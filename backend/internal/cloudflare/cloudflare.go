package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

type Config struct {
	AccountID    string // CF_ACCOUNT_ID
	APIToken     string // CF_API_TOKEN
	R2Bucket     string // CF_R2_BUCKET
	D1DatabaseID string // CF_D1_DATABASE_ID
}

func LoadConfig() Config {
	return Config{
		AccountID:    os.Getenv("CF_ACCOUNT_ID"),
		APIToken:     os.Getenv("CF_API_TOKEN"),
		R2Bucket:     os.Getenv("CF_R2_BUCKET"),
		D1DatabaseID: os.Getenv("CF_D1_DATABASE_ID"),
	}
}
func PostToR2(cfg Config, objectKey string, fileData []byte, filename string) (string, error) {
	// Detecta o Content-Type pelo nome do arquivo
	ext := filepath.Ext(filename)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	url := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/accounts/%s/r2/buckets/%s/objects/%s",
		cfg.AccountID, cfg.R2Bucket, objectKey,
	)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(fileData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar request R2: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao enviar arquivo para R2: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("R2 retornou status %d: %s", resp.StatusCode, string(body))
	}

	// Monta a URL pública (requer que o bucket tenha domínio público configurado).
	// Caso use um domínio customizado, ajuste aqui.
	publicURL := fmt.Sprintf(
		"https://%s.r2.cloudflarestorage.com/%s",
		cfg.R2Bucket, objectKey,
	)

	return publicURL, nil
}

// D1 — execução de queries SQL

type d1Request struct {
	SQL    string `json:"sql"`
	Params []any  `json:"params"`
}

type d1Response struct {
	Success bool `json:"success"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
	Result []struct {
		Meta struct {
			LastRowID int64 `json:"last_row_id"`
		} `json:"meta"`
	} `json:"result"`
}


func PostToD1(cfg Config, sql string, params []any) (int64, error) {
	payload := d1Request{SQL: sql, Params: params}

	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("erro ao serializar query D1: %w", err)
	}

	url := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s/query",
		cfg.AccountID, cfg.D1DatabaseID,
	)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("erro ao criar request D1: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("erro ao chamar API D1: %w", err)
	}
	defer resp.Body.Close()

	var d1Resp d1Response
	if err := json.NewDecoder(resp.Body).Decode(&d1Resp); err != nil {
		return 0, fmt.Errorf("erro ao decodificar resposta D1: %w", err)
	}

	if !d1Resp.Success {
		if len(d1Resp.Errors) > 0 {
			return 0, fmt.Errorf("D1 erro: %s", d1Resp.Errors[0].Message)
		}
		return 0, fmt.Errorf("D1 retornou sucesso=false sem mensagem de erro")
	}

	if len(d1Resp.Result) > 0 {
		return d1Resp.Result[0].Meta.LastRowID, nil
	}

	return 0, nil
}
