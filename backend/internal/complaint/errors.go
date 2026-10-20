package complaint

import "errors"

var (
	ErrValidation = errors.New("dados inválidos")
	ErrNotFound   = errors.New("denúncia não encontrada")
)
