package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/pedropaffaro/deep-pothole-backend/internal/complaint"
)

type Handler struct {
	svc     *complaint.Service
	timeout time.Duration
	log     *slog.Logger
}

func NewHandler(svc *complaint.Service, timeout time.Duration, log *slog.Logger) *Handler {
	return &Handler{svc: svc, timeout: timeout, log: log}
}

func (h *Handler) Register(r fiber.Router) {
	r.Post("/complaint", h.CreateComplaint)
	r.Get("/complaints", h.ListComplaints)
	r.Delete("/complaints/:id", h.DeleteComplaint)
	r.Get("/health", h.Health)
}

type complaintResponse struct {
	ID          uint    `json:"id"`
	City        string  `json:"city"`
	Street      string  `json:"street"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	PhotoKey    string  `json:"photo_key"`
	PhotoURL    string  `json:"photo_url"`
	DetectKey   string  `json:"detect_key"`
	DetectURL   string  `json:"detect_url"`
	Probability float64 `json:"pothole_probability"`
	Count       int     `json:"pothole_count"`
	State       string  `json:"state"`
	CreatedAt   string  `json:"created_at"`
}

func (h *Handler) toResponse(c complaint.Complaint) complaintResponse {
	return complaintResponse{
		ID:          c.ID,
		City:        c.City,
		Street:      c.Street,
		Latitude:    c.Latitude,
		Longitude:   c.Longitude,
		PhotoKey:    c.PhotoKey,
		PhotoURL:    h.svc.PhotoURL(c.PhotoKey),
		DetectKey:   complaint.DetectPhotoKey(c.PhotoKey),
		DetectURL:   h.svc.DetectPhotoURL(c.PhotoKey),
		Probability: c.Probability,
		Count:       c.Count,
		State:       c.State,
		CreatedAt:   c.CreatedAt,
	}
}

func (h *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *Handler) CreateComplaint(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), h.timeout)
	defer cancel()

	lat, err := strconv.ParseFloat(c.FormValue("latitude"), 64)
	if err != nil {
		return badRequest(c, "latitude inválida")
	}
	lng, err := strconv.ParseFloat(c.FormValue("longitude"), 64)
	if err != nil {
		return badRequest(c, "longitude inválida")
	}

	fileHeader, err := c.FormFile("foto")
	if err != nil {
		return badRequest(c, "foto é obrigatória")
	}
	file, err := fileHeader.Open()
	if err != nil {
		h.log.Error("abrir foto enviada", "erro", err)
		return internalError(c)
	}
	defer file.Close()

	out, err := h.svc.Create(ctx, complaint.CreateInput{
		City:        c.FormValue("cidade"),
		Street:      c.FormValue("rua"),
		Latitude:    lat,
		Longitude:   lng,
		Photo:       file,
		PhotoName:   fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
	})
	if err != nil {
		return h.writeError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(h.toResponse(*out))
}

func (h *Handler) ListComplaints(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), h.timeout)
	defer cancel()

	limit, _ := strconv.Atoi(c.Query("limit"))

	items, err := h.svc.List(ctx, limit)
	if err != nil {
		return h.writeError(c, err)
	}

	out := make([]complaintResponse, 0, len(items))
	for _, item := range items {
		out = append(out, h.toResponse(item))
	}
	return c.JSON(out)
}

func (h *Handler) DeleteComplaint(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), h.timeout)
	defer cancel()

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return badRequest(c, "id inválido")
	}

	if err := h.svc.Delete(ctx, uint(id)); err != nil {
		return h.writeError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) writeError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, complaint.ErrValidation):
		return badRequest(c, err.Error())
	case errors.Is(err, complaint.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, context.DeadlineExceeded):
		h.log.Warn("requisição estourou o tempo limite", "rota", c.Path())
		return c.Status(fiber.StatusGatewayTimeout).
			JSON(fiber.Map{"error": "tempo limite excedido"})
	case errors.Is(err, context.Canceled):
		return c.SendStatus(499)
	default:
		h.log.Error("falha ao processar requisição", "rota", c.Path(), "erro", err)
		return internalError(c)
	}
}

func badRequest(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
}

func internalError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).
		JSON(fiber.Map{"error": "erro interno"})
}
