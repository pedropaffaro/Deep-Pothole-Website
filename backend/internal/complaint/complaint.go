package complaint

import "io"

type Complaint struct {
	ID          uint
	City        string
	Street      string
	Latitude    float64
	Longitude   float64
	PhotoKey    string
	Probability float64
	Count       int
	State       string
	CreatedAt   string
}

type CreateInput struct {
	City        string
	Street      string
	Latitude    float64
	Longitude   float64
	Photo       io.Reader
	PhotoName   string
	ContentType string
}

// DetectPhotoKey devolve a chave da versão processada pelo modelo,
// gravada no bucket ao lado da foto original.
func DetectPhotoKey(photoKey string) string {
	if photoKey == "" {
		return ""
	}
	return photoKey + "-detect.jpg"
}
