package d1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pedropaffaro/deep-pothole-backend/internal/complaint"
)

const apiBase = "https://api.cloudflare.com/client/v4"

type Repo struct {
	client     *http.Client
	accountID  string
	databaseID string
	token      string
}

func New(accountID, databaseID, token string) *Repo {
	return &Repo{
		client:     &http.Client{},
		accountID:  accountID,
		databaseID: databaseID,
		token:      token,
	}
}

func (r *Repo) Create(ctx context.Context, c *complaint.Complaint) error {
	const query = `INSERT INTO complaints (city, street, latitude, longitude, photo_key)
	               VALUES (?, ?, ?, ?, ?)`

	res, err := r.query(ctx, query, c.City, c.Street, c.Latitude, c.Longitude, c.PhotoKey)
	if err != nil {
		return err
	}
	if len(res) == 0 {
		return fmt.Errorf("d1: INSERT não devolveu resultado")
	}
	if res[0].Meta.LastRowID <= 0 {
		return fmt.Errorf("d1: INSERT não devolveu last_row_id")
	}

	c.ID = uint(res[0].Meta.LastRowID)
	return nil
}

func (r *Repo) List(ctx context.Context, limit int) ([]complaint.Complaint, error) {
	const query = `SELECT id, city, street, latitude, longitude, photo_key
	               FROM complaints ORDER BY id DESC LIMIT ?`

	res, err := r.query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}

	var rows []struct {
		ID        uint    `json:"id"`
		City      string  `json:"city"`
		Street    string  `json:"street"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		PhotoKey  string  `json:"photo_key"`
	}
	if err := json.Unmarshal(res[0].Results, &rows); err != nil {
		return nil, fmt.Errorf("d1: decodificar linhas: %w", err)
	}

	out := make([]complaint.Complaint, 0, len(rows))
	for _, row := range rows {
		out = append(out, complaint.Complaint{
			ID:        row.ID,
			City:      row.City,
			Street:    row.Street,
			Latitude:  row.Latitude,
			Longitude: row.Longitude,
			PhotoKey:  row.PhotoKey,
		})
	}
	return out, nil
}

type queryResult struct {
	Results json.RawMessage `json:"results"`
	Meta    struct {
		LastRowID int64 `json:"last_row_id"`
	} `json:"meta"`
}

type apiResponse struct {
	Result  []queryResult `json:"result"`
	Success bool          `json:"success"`
	Errors  []apiError    `json:"errors"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (r *Repo) query(ctx context.Context, sql string, params ...any) ([]queryResult, error) {
	if params == nil {
		params = []any{}
	}
	body, err := json.Marshal(map[string]any{"sql": sql, "params": params})
	if err != nil {
		return nil, fmt.Errorf("d1: montar corpo: %w", err)
	}

	url := fmt.Sprintf("%s/accounts/%s/d1/database/%s/query", apiBase, r.accountID, r.databaseID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("d1: montar requisição: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+r.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("d1: chamar API: %w", err)
	}
	defer resp.Body.Close()

	var parsed apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("d1: decodificar resposta (HTTP %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK || !parsed.Success {
		return nil, fmt.Errorf("d1: HTTP %d: %s", resp.StatusCode, joinErrors(parsed.Errors))
	}
	return parsed.Result, nil
}

func joinErrors(errs []apiError) string {
	if len(errs) == 0 {
		return "sem detalhe na resposta"
	}
	msgs := make([]string, 0, len(errs))
	for _, e := range errs {
		msgs = append(msgs, fmt.Sprintf("[%d] %s", e.Code, e.Message))
	}
	return strings.Join(msgs, "; ")
}
