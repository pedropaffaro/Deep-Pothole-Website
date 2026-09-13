package complaint

import (
	"context"
	"io"
)

type Repository interface {
	Create(ctx context.Context, c *Complaint) error
	List(ctx context.Context, limit int) ([]Complaint, error)
}

type Detector interface {
	Detect(ctx context.Context, img []byte) ([]byte, error)
}

type PhotoStorage interface {
	Upload(ctx context.Context, key, contentType string, r io.Reader) error
	URL(key string) string
}
