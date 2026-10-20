package complaint

import (
	"context"
	"io"
)

type Repository interface {
	Create(ctx context.Context, c *Complaint) error
	List(ctx context.Context, limit int) ([]Complaint, error)
	Get(ctx context.Context, id uint) (*Complaint, error)
	Delete(ctx context.Context, id uint) error
}

type Detector interface {
	Detect(ctx context.Context, img []byte) ([]byte, float64, int, error)
}

type PhotoStorage interface {
	Upload(ctx context.Context, key, contentType string, r io.Reader) error
	Delete(ctx context.Context, key string) error
	URL(key string) string
}
