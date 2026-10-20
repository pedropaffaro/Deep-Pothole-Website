package d1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	const query = `INSERT INTO complaints (city, street, latitude, longitude, photo_key, pothole_probability, pothole_count, created_at)
	               VALUES (?, ?, ?, ?, ?, ?, ?, ?)` //o fato de já definir aqui como é a estrutura da query, remove o perigo de query injection

	res, err := r.query(ctx, query, []string{
		c.City,
		c.Street,
		strconv.FormatFloat(c.Latitude, 'f', -1, 64),
		strconv.FormatFloat(c.Longitude, 'f', -1, 64),
		c.PhotoKey,
		strconv.FormatFloat(c.Probability, 'f', -1, 64),
		strconv.Itoa(c.Count),
		c.CreatedAt,
	})
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
	const query = `SELECT id, city, street, latitude, longitude, photo_key, pothole_probability, pothole_count, state, created_at
	               FROM complaints ORDER BY id DESC LIMIT ?`

	res, err := r.query(ctx, query, []string{strconv.Itoa(limit)})
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}

	var rows []complaintRow
	if err := json.Unmarshal(res[0].Results, &rows); err != nil {
		return nil, fmt.Errorf("d1: decodificar linhas: %w", err)
	}

	out := make([]complaint.Complaint, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toDomain())
	}
	return out, nil
}

func (r *Repo) Get(ctx context.Context, id uint) (*complaint.Complaint, error) {
	const query = `SELECT id, city, street, latitude, longitude, photo_key, pothole_probability, pothole_count, state, created_at
	               FROM complaints WHERE id = ?`

	res, err := r.query(ctx, query, []string{strconv.FormatUint(uint64(id), 10)})
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, complaint.ErrNotFound
	}

	var rows []complaintRow
	if err := json.Unmarshal(res[0].Results, &rows); err != nil {
		return nil, fmt.Errorf("d1: decodificar linha: %w", err)
	}
	if len(rows) == 0 {
		return nil, complaint.ErrNotFound
	}

	out := rows[0].toDomain()
	return &out, nil
}

func (r *Repo) Delete(ctx context.Context, id uint) error {
	const query = `DELETE FROM complaints WHERE id = ?`

	res, err := r.query(ctx, query, []string{strconv.FormatUint(uint64(id), 10)})
	if err != nil {
		return err
	}
	if len(res) == 0 || res[0].Meta.Changes == 0 {
		return complaint.ErrNotFound
	}
	return nil
}

type complaintRow struct {
	ID          uint    `json:"id"`
	City        string  `json:"city"`
	Street      string  `json:"street"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	PhotoKey    string  `json:"photo_key"`
	Probability float64 `json:"pothole_probability"`
	Count       int     `json:"pothole_count"`
	State       string  `json:"state"`
	CreatedAt   string  `json:"created_at"`
}

func (r complaintRow) toDomain() complaint.Complaint {
	return complaint.Complaint{
		ID:          r.ID,
		City:        r.City,
		Street:      r.Street,
		Latitude:    r.Latitude,
		Longitude:   r.Longitude,
		PhotoKey:    r.PhotoKey,
		Probability: r.Probability,
		Count:       r.Count,
		State:       r.State,
		CreatedAt:   r.CreatedAt,
	}
}

type queryResult struct {
	Results json.RawMessage `json:"results"`
	Meta    struct {
		LastRowID int64 `json:"last_row_id"`
		Changes   int64 `json:"changes"`
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

func (r *Repo) query(ctx context.Context, sql string, params []string) ([]queryResult, error) {
	if params == nil {
		params = []string{}
	}
	body, err := json.Marshal(map[string]any{"sql": sql, "params": params})
	if err != nil {
		return nil, fmt.Errorf("d1: montar corpo: %w", err)
	}

	url := fmt.Sprintf("%s/accounts/%s/d1/database/%s/query", apiBase, r.accountID, r.databaseID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body)) //para isso que precisava do "ctx"
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
