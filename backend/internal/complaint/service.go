package complaint

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"
)

var allowedExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".heic": true,
}

type Service struct {
	repo     Repository
	photos   PhotoStorage
	detector Detector
}

func NewService(repo Repository, photos PhotoStorage, detector Detector) *Service {
	return &Service{repo: repo, photos: photos, detector: detector}
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*Complaint, error) {
	city := strings.TrimSpace(in.City)
	street := strings.TrimSpace(in.Street)

	if city == "" {
		return nil, fmt.Errorf("%w: cidade é obrigatória", ErrValidation)
	}
	if street == "" {
		return nil, fmt.Errorf("%w: rua é obrigatória", ErrValidation)
	}
	if in.Latitude < -90 || in.Latitude > 90 {
		return nil, fmt.Errorf("%w: latitude %v fora do intervalo [-90, 90]", ErrValidation, in.Latitude)
	}
	if in.Longitude < -180 || in.Longitude > 180 {
		return nil, fmt.Errorf("%w: longitude %v fora do intervalo [-180, 180]", ErrValidation, in.Longitude)
	}
	if in.Latitude == 0 && in.Longitude == 0 {
		return nil, fmt.Errorf("%w: localização não informada", ErrValidation)
	}
	if in.Photo == nil {
		return nil, fmt.Errorf("%w: foto é obrigatória", ErrValidation)
	}

	key, err := s.photoKey(in.PhotoName)
	if err != nil {
		return nil, err
	}

	raw, err := io.ReadAll(in.Photo)
	if err != nil {
		return nil, fmt.Errorf("ler foto: %w", err)
	}

	if err := s.photos.Upload(ctx, key, in.ContentType, bytes.NewReader(raw)); err != nil {
		return nil, fmt.Errorf("gravar foto: %w", err)
	}

	out, err := s.detector.Detect(ctx, raw)
	if err != nil {
		return nil, fmt.Errorf("detectar: %w", err)
	}
	outKey := key + "-detect.jpg"
	if err := s.photos.Upload(ctx, outKey, "image/jpeg", bytes.NewReader(out)); err != nil {
		return nil, fmt.Errorf("gravar saida do modelo: %w", err)
	}
	slog.Info("modelo: saida gravada", "chave", outKey, "bytes", len(out))

	c := &Complaint{
		City:      city,
		Street:    street,
		Latitude:  in.Latitude,
		Longitude: in.Longitude,
		PhotoKey:  key,
	}

	if err := s.repo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("salvar denúncia: %w", err)
	}
	return c, nil
}

func (s *Service) List(ctx context.Context, limit int) ([]Complaint, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return s.repo.List(ctx, limit)
}

func (s *Service) PhotoURL(key string) string {
	return s.photos.URL(key)
}

func (s *Service) photoKey(originalName string) (string, error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	if !allowedExtensions[ext] {
		return "", fmt.Errorf("%w: extensão de imagem %q não suportada", ErrValidation, ext)
	}

	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("gerar nome da foto: %w", err)
	}

	return fmt.Sprintf("complaints/%d-%s%s", time.Now().Unix(), hex.EncodeToString(buf[:]), ext), nil
}
