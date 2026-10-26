"""Profundidade de buracos em relação ao asfalto plano, a partir de uma foto.

Uso: uv run uvicorn main:app --port 8000
"""

import io
import math
from contextlib import asynccontextmanager
from pathlib import Path

import numpy as np
import onnxruntime as ort
import torch
from fastapi import FastAPI, File, Form, HTTPException, UploadFile
from PIL import Image, ImageOps
from scipy import ndimage
from transformers import AutoImageProcessor, AutoModelForDepthEstimation

POTHOLE_MODEL = Path(__file__).parent.parent / "backend/internal/adapters/onnx/model/pothole.onnx"
DEPTH_MODEL = "depth-anything/Depth-Anything-V2-Metric-Outdoor-Small-hf"
DEFAULT_HFOV_DEG = 70  # usado quando não há distância focal
MAX_SIDE = 1280
SEG_SIZE = 960
CENTER_POINTS = 10  # pixels centrais da máscara que viram o "centro único"
MIN_RING_PIXELS = 200


def detect(session, img):
    """Máscaras (bool, tamanho da imagem) e scores dos buracos, mesma decodificação do detector em Go."""
    w, h = img.size
    x = np.asarray(img.resize((SEG_SIZE, SEG_SIZE), Image.Resampling.BICUBIC), np.float32) / 255
    raw, protos = (o[0] for o in session.run(None, {"images": x.transpose(2, 0, 1)[None]}))

    raw = raw[:, raw[4] >= 0.5]  # linhas: cx, cy, w, h, score, coeficientes
    raw = raw[:, np.argsort(-raw[4])]
    boxes = np.stack([raw[0] - raw[2] / 2, raw[1] - raw[3] / 2, raw[0] + raw[2] / 2, raw[1] + raw[3] / 2], 1)

    area = lambda a: (a[..., 2] - a[..., 0]) * (a[..., 3] - a[..., 1])
    kept = []
    for i, b in enumerate(boxes):
        inter = np.clip(np.minimum(b[2:], boxes[kept, 2:]) - np.maximum(b[:2], boxes[kept, :2]), 0, None).prod(1)
        if not len(kept) or (inter / (area(b) + area(boxes[kept]) - inter)).max() <= 0.5:
            kept.append(i)

    out = []
    for i in kept:
        logits = (raw[5:, i] @ protos.reshape(len(protos), -1)).reshape(protos.shape[1:])
        prob = 1 / (1 + np.exp(-logits))
        s = protos.shape[1] / SEG_SIZE
        x1, y1, x2, y2 = (np.array(boxes[i]) * s).astype(int).clip(0)
        crop = np.zeros_like(prob)
        crop[y1:y2, x1:x2] = prob[y1:y2, x1:x2]
        mask = np.asarray(Image.fromarray(crop, "F").resize((w, h), Image.Resampling.BILINEAR)) >= 0.5
        if mask.any():
            out.append((float(raw[4, i]), mask))
    return out


def estimate_depth(processor, model, img):
    """Mapa (H, W) em metros ao longo do eixo da câmera."""
    inputs = processor(images=img, return_tensors="pt")
    with torch.inference_mode():
        depth = model(**inputs).predicted_depth
    resized = torch.nn.functional.interpolate(depth[None], size=img.size[::-1], mode="bilinear")
    return resized[0, 0].numpy()


def fit_plane(pts, rel_threshold=0.01, iterations=300):
    """Plano n·p + d = 0 do asfalto (RANSAC), com a câmera do lado positivo. Devolve também os inliers."""
    rng = np.random.default_rng(0)
    threshold = max(0.003, rel_threshold * np.median(pts[:, 2]))

    best = None
    for _ in range(iterations):
        a, b, c = pts[rng.choice(len(pts), 3, replace=False)]
        n = np.cross(b - a, c - a)
        norm = np.linalg.norm(n)
        if norm < 1e-9:
            continue
        inliers = np.abs((pts - a) @ n) / norm < threshold
        if best is None or inliers.sum() > best.sum():
            best = inliers

    centroid = pts[best].mean(0)
    n = np.linalg.svd(pts[best] - centroid)[2][-1]
    d = -n @ centroid
    return (n, d, best) if d >= 0 else (-n, -d, best)


def measure(depth, mask, others, focal_px):
    """Profundidade do centro do buraco abaixo do plano do asfalto ao redor dele."""
    h, w = depth.shape
    u, v = np.meshgrid(np.arange(w), np.arange(h))
    pts = np.stack([(u - w / 2) * depth / focal_px, (v - h / 2) * depth / focal_px, depth], -1)

    # Anel de asfalto: folga da borda do buraco e longe dos outros buracos
    radius = math.sqrt(mask.sum() / math.pi)
    gap = max(3, 0.25 * radius)
    dist = ndimage.distance_transform_edt(~mask)
    ring = (dist > gap) & (dist <= gap + max(8, radius)) & (ndimage.distance_transform_edt(~others) > gap) & (depth > 0)
    if ring.sum() < MIN_RING_PIXELS:
        return {"error": "asfalto insuficiente ao redor do buraco"}

    n, d, inliers = fit_plane(pts[ring])
    rmse = float(np.sqrt(np.mean((pts[ring][inliers] @ n + d) ** 2)))

    # Centro único: média dos pontos 3D dos pixels da máscara mais próximos do centróide
    valid = mask & (depth > 0)
    rows, cols = np.nonzero(valid)
    if len(rows) < CENTER_POINTS:
        return {"error": "máscara pequena demais"}
    nearest = np.argsort((rows - rows.mean()) ** 2 + (cols - cols.mean()) ** 2)[:CENTER_POINTS]
    center = pts[rows[nearest], cols[nearest]].mean(0)
    depth_center = float(-(center @ n + d))

    return {
        "depth_center_m": depth_center,
        "depth_max_m": float(np.percentile(-(pts[valid] @ n + d), 95)),
        "plane_rmse_m": rmse,
        "plane_inlier_ratio": float(inliers.mean()),
        # Sem isso a profundidade se confunde com o ruído do próprio asfalto
        "reliable": bool(inliers.mean() >= 0.5 and depth_center > 2 * rmse),
    }


@asynccontextmanager
async def lifespan(app):
    app.state.seg = ort.InferenceSession(str(POTHOLE_MODEL), providers=["CPUExecutionProvider"])
    app.state.processor = AutoImageProcessor.from_pretrained(DEPTH_MODEL)
    app.state.depth = AutoModelForDepthEstimation.from_pretrained(DEPTH_MODEL).eval()
    yield


app = FastAPI(lifespan=lifespan)


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/v1/depth")
def depth(foto: UploadFile = File(...), focal_35mm: float | None = Form(None)):
    try:
        img = Image.open(io.BytesIO(foto.file.read()))
    except OSError:
        raise HTTPException(400, "imagem inválida")

    # Distância focal equivalente em 35mm: do formulário ou do EXIF
    focal_35mm = focal_35mm or img.getexif().get_ifd(0x8769).get(0xA405)
    img = ImageOps.exif_transpose(img).convert("RGB")
    img.thumbnail((MAX_SIDE, MAX_SIDE))

    if focal_35mm:
        focal_px = focal_35mm * math.hypot(*img.size) / math.hypot(36, 24)
    else:
        focal_px = (img.width / 2) / math.tan(math.radians(DEFAULT_HFOV_DEG) / 2)

    found = detect(app.state.seg, img)
    if not found:
        return {"focal_px": focal_px, "potholes": []}

    depth_map = estimate_depth(app.state.processor, app.state.depth, img)
    everything = np.logical_or.reduce([m for _, m in found])
    potholes = [{"score": score, **measure(depth_map, mask, everything, focal_px)} for score, mask in found]
    return {"focal_px": focal_px, "potholes": potholes}
