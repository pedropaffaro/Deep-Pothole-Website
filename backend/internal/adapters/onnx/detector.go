package onnx

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"log/slog"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	ort "github.com/yalue/onnxruntime_go"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

//go:embed model/pothole.onnx
var modelData []byte

const (
	imgSize   = 960
	numBoxes  = 18900
	numFields = 37
	numProtos = 32
	maskSize  = 240
	confMin   = 0.5
	iouMax    = 0.5
	maskMin   = 0.5
	lineWidth = 4
	maskAlpha = 0.45
	debugTopK = 10
)

var boxColor = color.RGBA{255, 0, 0, 255}

type box struct {
	x1, y1, x2, y2, score float32
	coefs                 []float32
}

type Detector struct {
	session *ort.DynamicAdvancedSession
}

func New() (*Detector, error) {
	slog.Info("onnx: iniciando")

	ort.SetSharedLibraryPath(os.Getenv("ONNXRUNTIME_LIB_PATH"))
	if err := ort.InitializeEnvironment(); err != nil {
		return nil, err
	}
	slog.Info("onnx: ambiente pronto")

	session, err := ort.NewDynamicAdvancedSessionWithONNXData(
		modelData,
		[]string{"images"},
		[]string{"output0", "output1"},
		nil,
	)
	if err != nil {
		return nil, err
	}
	slog.Info("onnx: sessao criada", "modelo_bytes", len(modelData))

	return &Detector{session: session}, nil
}

func (d *Detector) Close() error {
	return d.session.Destroy()
}

func (d *Detector) Detect(ctx context.Context, img []byte) ([]byte, float64, int, error) {
	start := time.Now()
	slog.Info("onnx: detect inicio", "imagem_bytes", len(img))

	src, format, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		slog.Error("onnx: decode falhou", "tipo_detectado", http.DetectContentType(img), "primeiros_bytes", fmt.Sprintf("% x", img[:min(16, len(img))]), "erro", err)
		return nil, 0, 0, err
	}
	slog.Info("onnx: imagem decodificada", "formato", format, "largura", src.Bounds().Dx(), "altura", src.Bounds().Dy())
	dump("1-original.jpg", src)

	resized := resize(src)
	dump("2-entrada-960.jpg", resized)

	inTensor, err := ort.NewTensor(ort.NewShape(1, 3, imgSize, imgSize), preprocess(resized))
	if err != nil {
		return nil, 0, 0, err
	}
	defer inTensor.Destroy()
	slog.Info("onnx: preprocess ok", "ms", time.Since(start).Milliseconds())

	outTensor, err := ort.NewEmptyTensor[float32](ort.NewShape(1, numFields, numBoxes))
	if err != nil {
		return nil, 0, 0, err
	}
	defer outTensor.Destroy()

	protoTensor, err := ort.NewEmptyTensor[float32](ort.NewShape(1, numProtos, maskSize, maskSize))
	if err != nil {
		return nil, 0, 0, err
	}
	defer protoTensor.Destroy()

	slog.Info("onnx: run inicio")
	if err := d.session.Run([]ort.Value{inTensor}, []ort.Value{outTensor, protoTensor}); err != nil {
		return nil, 0, 0, err
	}
	slog.Info("onnx: run ok", "ms", time.Since(start).Milliseconds())

	raw := outTensor.GetData()
	protos := protoTensor.GetData()
	logTensor("output0", raw, numFields, numBoxes)
	logTensor("output1", protos, numProtos, maskSize*maskSize)
	logFields(raw)
	logScores(raw)
	logTopBoxes(raw)
	dumpOutput0(raw)
	dumpOutput1(protos)
	probability := float64(topScore(raw))

	cands := candidates(raw)
	slog.Info("onnx: candidatos acima do limiar", "total", len(cands), "limiar", confMin)
	dumpBoxes("3-antes-nms.jpg", src, cands, protos)

	boxes := nms(cands)
	slog.Info("onnx: boxes apos nms", "total", len(boxes), "iou_max", iouMax)
	for i, b := range boxes {
		slog.Info("onnx: box",
			"i", i,
			"score", b.score,
			"em_960", fmt.Sprintf("[%.0f %.0f %.0f %.0f]", b.x1, b.y1, b.x2, b.y2),
			"no_original", fmt.Sprintf("[%.0f %.0f %.0f %.0f]",
				b.x1*float32(src.Bounds().Dx())/imgSize,
				b.y1*float32(src.Bounds().Dy())/imgSize,
				b.x2*float32(src.Bounds().Dx())/imgSize,
				b.y2*float32(src.Bounds().Dy())/imgSize))
	}

	logMasks(boxes, protos)

	out, err := annotate(src, boxes, protos)
	if err != nil {
		return nil, 0, 0, err
	}
	slog.Info("onnx: detect fim", "saida_bytes", len(out), "probabilidade", probability, "ms", time.Since(start).Milliseconds())
	if dir := os.Getenv("DEBUG_DIR"); dir != "" {
		os.WriteFile(filepath.Join(dir, "4-final.jpg"), out, 0644)
		slog.Info("onnx: debug gravado", "dir", dir)
	}

	return out, probability, len(boxes), nil
}

func stats(data []float32) (float32, float32, float32) {
	lo, hi := data[0], data[0]
	var sum float64
	for _, v := range data {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
		sum += float64(v)
	}
	return lo, hi, float32(sum / float64(len(data)))
}

func logTensor(name string, data []float32, rows, cols int) {
	lo, hi, mean := stats(data)
	slog.Info("onnx: tensor", "nome", name, "valores", len(data), "forma", fmt.Sprintf("[1 %d %d]", rows, cols), "min", lo, "max", hi, "media", mean)
}

func logFields(out []float32) {
	names := []string{"cx", "cy", "w", "h", "score"}
	for j := 0; j < numFields; j++ {
		name := fmt.Sprintf("coef%d", j-5)
		if j < len(names) {
			name = names[j]
		}
		lo, hi, mean := stats(out[j*numBoxes : (j+1)*numBoxes])
		slog.Info("onnx: campo", "indice", j, "nome", name, "min", lo, "max", hi, "media", mean)
	}
}

func logTopBoxes(out []float32) {
	idx := make([]int, numBoxes)
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(a, b int) bool { return out[4*numBoxes+idx[a]] > out[4*numBoxes+idx[b]] })

	for r := 0; r < debugTopK && r < numBoxes; r++ {
		i := idx[r]
		coefs := make([]string, numProtos)
		for k := 0; k < numProtos; k++ {
			coefs[k] = fmt.Sprintf("%.4f", out[(5+k)*numBoxes+i])
		}
		cx, cy := out[i], out[numBoxes+i]
		w, h := out[2*numBoxes+i], out[3*numBoxes+i]
		slog.Info("onnx: bruto",
			"rank", r,
			"indice", i,
			"score", out[4*numBoxes+i],
			"cx", cx, "cy", cy, "w", w, "h", h,
			"xyxy", fmt.Sprintf("[%.1f %.1f %.1f %.1f]", cx-w/2, cy-h/2, cx+w/2, cy+h/2),
			"coefs", "["+strings.Join(coefs, " ")+"]")
	}
}

func logMasks(boxes []box, protos []float32) {
	for i, b := range boxes {
		mask := buildMask(b, protos)
		acima := 0
		var hi float32
		for _, v := range mask {
			if v >= maskMin {
				acima++
			}
			if v > hi {
				hi = v
			}
		}
		slog.Info("onnx: mascara", "i", i, "score", b.score, "pixels_acima", acima, "total_pixels", len(mask), "max", hi, "limiar", maskMin)
	}
}

func dumpOutput0(out []float32) {
	dir := os.Getenv("DEBUG_DIR")
	if dir == "" {
		return
	}
	f, err := os.Create(filepath.Join(dir, "5-output0.csv"))
	if err != nil {
		slog.Error("onnx: debug output0 falhou", "erro", err)
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	fmt.Fprint(w, "box,cx,cy,w,h,score")
	for k := 0; k < numProtos; k++ {
		fmt.Fprintf(w, ",coef%d", k)
	}
	fmt.Fprintln(w)

	for i := 0; i < numBoxes; i++ {
		fmt.Fprintf(w, "%d", i)
		for j := 0; j < numFields; j++ {
			fmt.Fprintf(w, ",%.6f", out[j*numBoxes+i])
		}
		fmt.Fprintln(w)
	}
	slog.Info("onnx: debug output0 gravado", "arquivo", filepath.Join(dir, "5-output0.csv"), "linhas", numBoxes)
}

func dumpOutput1(protos []float32) {
	dir := os.Getenv("DEBUG_DIR")
	if dir == "" {
		return
	}
	f, err := os.Create(filepath.Join(dir, "6-output1.bin"))
	if err != nil {
		slog.Error("onnx: debug output1 falhou", "erro", err)
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	if err := binary.Write(w, binary.LittleEndian, protos); err != nil {
		slog.Error("onnx: debug output1 falhou", "erro", err)
		return
	}
	slog.Info("onnx: debug output1 gravado", "arquivo", filepath.Join(dir, "6-output1.bin"), "float32", len(protos), "forma", fmt.Sprintf("[%d %d %d]", numProtos, maskSize, maskSize))
}

func topScore(out []float32) float32 {
	var top float32
	for i := 0; i < numBoxes; i++ {
		if s := out[4*numBoxes+i]; s > top {
			top = s
		}
	}
	return top
}

func resize(src image.Image) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, imgSize, imgSize))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Src, nil)
	return dst
}

func preprocess(dst *image.RGBA) []float32 {
	plane := imgSize * imgSize
	data := make([]float32, 3*plane)
	for y := 0; y < imgSize; y++ {
		for x := 0; x < imgSize; x++ {
			i := dst.PixOffset(x, y)
			p := y*imgSize + x
			data[p] = float32(dst.Pix[i]) / 255
			data[plane+p] = float32(dst.Pix[i+1]) / 255
			data[2*plane+p] = float32(dst.Pix[i+2]) / 255
		}
	}
	return data
}

func logScores(out []float32) {
	var maxScore float32
	acima := map[string]int{"0.10": 0, "0.25": 0, "0.50": 0, "0.75": 0}
	for i := 0; i < numBoxes; i++ {
		s := out[4*numBoxes+i]
		if s > maxScore {
			maxScore = s
		}
		for _, t := range []struct {
			k string
			v float32
		}{{"0.10", 0.10}, {"0.25", 0.25}, {"0.50", 0.50}, {"0.75", 0.75}} {
			if s > t.v {
				acima[t.k]++
			}
		}
	}
	slog.Info("onnx: scores", "max", maxScore, "acima_0.10", acima["0.10"], "acima_0.25", acima["0.25"], "acima_0.50", acima["0.50"], "acima_0.75", acima["0.75"])
}

func candidates(out []float32) []box {
	var boxes []box
	for i := 0; i < numBoxes; i++ {
		score := out[4*numBoxes+i]
		if score < confMin {
			continue
		}
		cx, cy := out[i], out[numBoxes+i]
		w, h := out[2*numBoxes+i], out[3*numBoxes+i]

		coefs := make([]float32, numProtos)
		for k := 0; k < numProtos; k++ {
			coefs[k] = out[(5+k)*numBoxes+i]
		}
		boxes = append(boxes, box{cx - w/2, cy - h/2, cx + w/2, cy + h/2, score, coefs})
	}
	sort.Slice(boxes, func(a, b int) bool { return boxes[a].score > boxes[b].score })
	return boxes
}

func nms(boxes []box) []box {
	var kept []box
	for _, b := range boxes {
		keep := true
		for _, k := range kept {
			if iou(b, k) > iouMax {
				keep = false
				break
			}
		}
		if keep {
			kept = append(kept, b)
		}
	}
	return kept
}

func iou(a, b box) float32 {
	x1 := max(a.x1, b.x1)
	y1 := max(a.y1, b.y1)
	x2 := min(a.x2, b.x2)
	y2 := min(a.y2, b.y2)
	if x2 <= x1 || y2 <= y1 {
		return 0
	}
	inter := (x2 - x1) * (y2 - y1)
	areaA := (a.x2 - a.x1) * (a.y2 - a.y1)
	areaB := (b.x2 - b.x1) * (b.y2 - b.y1)
	return inter / (areaA + areaB - inter)
}

func annotate(src image.Image, boxes []box, protos []float32) ([]byte, error) {
	bounds := src.Bounds()
	out := image.NewRGBA(bounds)
	draw.Draw(out, bounds, src, bounds.Min, draw.Src)

	sx := float32(bounds.Dx()) / imgSize
	sy := float32(bounds.Dy()) / imgSize

	for _, b := range boxes {
		r := image.Rect(
			bounds.Min.X+int(b.x1*sx),
			bounds.Min.Y+int(b.y1*sy),
			bounds.Min.X+int(b.x2*sx),
			bounds.Min.Y+int(b.y2*sy),
		).Intersect(bounds)

		mask := buildMask(b, protos)
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				mx := int(float32(x-bounds.Min.X) / sx / 4)
				my := int(float32(y-bounds.Min.Y) / sy / 4)
				if mx < 0 || mx >= maskSize || my < 0 || my >= maskSize {
					continue
				}
				if mask[my*maskSize+mx] < maskMin {
					continue
				}
				i := out.PixOffset(x, y)
				out.Pix[i] = blend(out.Pix[i], boxColor.R)
				out.Pix[i+1] = blend(out.Pix[i+1], boxColor.G)
				out.Pix[i+2] = blend(out.Pix[i+2], boxColor.B)
			}
		}

		for i := 0; i < lineWidth; i++ {
			edge := r.Inset(i)
			if edge.Empty() {
				break
			}
			draw.Draw(out, image.Rect(edge.Min.X, edge.Min.Y, edge.Max.X, edge.Min.Y+1), &image.Uniform{boxColor}, image.Point{}, draw.Src)
			draw.Draw(out, image.Rect(edge.Min.X, edge.Max.Y-1, edge.Max.X, edge.Max.Y), &image.Uniform{boxColor}, image.Point{}, draw.Src)
			draw.Draw(out, image.Rect(edge.Min.X, edge.Min.Y, edge.Min.X+1, edge.Max.Y), &image.Uniform{boxColor}, image.Point{}, draw.Src)
			draw.Draw(out, image.Rect(edge.Max.X-1, edge.Min.Y, edge.Max.X, edge.Max.Y), &image.Uniform{boxColor}, image.Point{}, draw.Src)
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, out, &jpeg.Options{Quality: 85}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildMask(b box, protos []float32) []float32 {
	mask := make([]float32, maskSize*maskSize)
	for k := 0; k < numProtos; k++ {
		c := b.coefs[k]
		if c == 0 {
			continue
		}
		off := k * maskSize * maskSize
		for p := range mask {
			mask[p] += c * protos[off+p]
		}
	}
	for p, v := range mask {
		mask[p] = float32(1 / (1 + math.Exp(float64(-v))))
	}
	return mask
}

func blend(base, over uint8) uint8 {
	return uint8(float32(base)*(1-maskAlpha) + float32(over)*maskAlpha)
}

func dump(name string, img image.Image) {
	dir := os.Getenv("DEBUG_DIR")
	if dir == "" {
		return
	}
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		slog.Error("onnx: debug falhou", "arquivo", name, "erro", err)
		return
	}
	defer f.Close()
	jpeg.Encode(f, img, &jpeg.Options{Quality: 85})
}

func dumpBoxes(name string, src image.Image, boxes []box, protos []float32) {
	if os.Getenv("DEBUG_DIR") == "" {
		return
	}
	out, err := annotate(src, boxes, protos)
	if err != nil {
		return
	}
	os.WriteFile(filepath.Join(os.Getenv("DEBUG_DIR"), name), out, 0644)
}
