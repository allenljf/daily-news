// Package category implements Category and Source Setting operations.
package category

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/allenljf/daily-news/backend/internal/httpapi"
	"github.com/allenljf/daily-news/backend/internal/identity"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("category not found")

const DefaultContentLanguage = "zh-Hant"

type SourceSettingInput struct {
	Label        string `json:"label"`
	WebsiteInput string `json:"website_input"`
	Kind         string `json:"kind"`
}
type Request struct {
	Name                string               `json:"name"`
	SearchKeywords      *string              `json:"search_keywords"`
	SpecialRequirements *string              `json:"special_requirements"`
	SourceSettings      []SourceSettingInput `json:"source_settings"`
	ContentLanguage     string               `json:"content_language"`
	contentLanguageSet  bool
}

// UnmarshalJSON records whether content_language was supplied so an explicit
// unsupported empty value cannot be mistaken for a backward-compatible omission.
func (request *Request) UnmarshalJSON(data []byte) error {
	var payload struct {
		Name                string               `json:"name"`
		SearchKeywords      *string              `json:"search_keywords"`
		SpecialRequirements *string              `json:"special_requirements"`
		SourceSettings      []SourceSettingInput `json:"source_settings"`
		ContentLanguage     json.RawMessage      `json:"content_language"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("trailing JSON request data")
	}
	request.Name = payload.Name
	request.SearchKeywords = payload.SearchKeywords
	request.SpecialRequirements = payload.SpecialRequirements
	request.SourceSettings = payload.SourceSettings
	request.contentLanguageSet = payload.ContentLanguage != nil
	request.ContentLanguage = ""
	if request.contentLanguageSet {
		if err := json.Unmarshal(payload.ContentLanguage, &request.ContentLanguage); err != nil {
			return err
		}
	}
	return nil
}

type SourceSetting struct {
	ID             uuid.UUID `json:"id"`
	Label          string    `json:"label"`
	WebsiteInput   string    `json:"website_input"`
	NormalizedHost *string   `json:"normalized_host"`
	Kind           string    `json:"kind"`
	Position       int       `json:"position"`
}
type Response struct {
	ID                  uuid.UUID       `json:"id"`
	Name                string          `json:"name"`
	SearchKeywords      *string         `json:"search_keywords"`
	SpecialRequirements *string         `json:"special_requirements"`
	SourceSettings      []SourceSetting `json:"source_settings"`
	ContentLanguage     string          `json:"content_language"`
}

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (store *Store) List(ctx context.Context) ([]Response, error) {
	rows, err := store.db.QueryContext(ctx, `SELECT id,name,search_keywords,special_requirements,content_language FROM categories WHERE deleted_at IS NULL ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]Response, 0)
	for rows.Next() {
		var value Response
		if err := rows.Scan(&value.ID, &value.Name, &value.SearchKeywords, &value.SpecialRequirements, &value.ContentLanguage); err != nil {
			return nil, err
		}
		settings, err := store.settings(ctx, store.db, value.ID)
		if err != nil {
			return nil, err
		}
		value.SourceSettings = settings
		results = append(results, value)
	}
	return results, rows.Err()
}

func (store *Store) Create(ctx context.Context, request Request) (Response, error) {
	if err := validate(request); err != nil {
		return Response{}, err
	}
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return Response{}, err
	}
	defer tx.Rollback()
	contentLanguage, err := normalizedContentLanguage(request.ContentLanguage)
	if err != nil {
		return Response{}, err
	}
	result := Response{ID: uuid.New(), Name: strings.TrimSpace(request.Name), SearchKeywords: trimOptional(request.SearchKeywords), SpecialRequirements: trimOptional(request.SpecialRequirements), ContentLanguage: contentLanguage}
	if _, err = tx.ExecContext(ctx, `INSERT INTO categories (id,name,search_keywords,special_requirements,content_language) VALUES ($1,$2,$3,$4,$5)`, result.ID, result.Name, result.SearchKeywords, result.SpecialRequirements, result.ContentLanguage); err != nil {
		return Response{}, err
	}
	if result.SourceSettings, err = insertSettings(ctx, tx, result.ID, request.SourceSettings); err != nil {
		return Response{}, err
	}
	if err = tx.Commit(); err != nil {
		return Response{}, err
	}
	return result, nil
}

func (store *Store) Update(ctx context.Context, id uuid.UUID, request Request) (Response, error) {
	if err := validate(request); err != nil {
		return Response{}, err
	}
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return Response{}, err
	}
	defer tx.Rollback()
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM categories WHERE id=$1 AND deleted_at IS NULL)`, id).Scan(&exists); err != nil {
		return Response{}, err
	}
	if !exists {
		return Response{}, ErrNotFound
	}
	contentLanguage, err := normalizedContentLanguage(request.ContentLanguage)
	if err != nil {
		return Response{}, err
	}
	result := Response{ID: id, Name: strings.TrimSpace(request.Name), SearchKeywords: trimOptional(request.SearchKeywords), SpecialRequirements: trimOptional(request.SpecialRequirements), ContentLanguage: contentLanguage}
	if _, err = tx.ExecContext(ctx, `UPDATE categories SET name=$2,search_keywords=$3,special_requirements=$4,content_language=$5,updated_at=now() WHERE id=$1`, id, result.Name, result.SearchKeywords, result.SpecialRequirements, result.ContentLanguage); err != nil {
		return Response{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE source_settings SET deleted_at=now() WHERE category_id=$1 AND deleted_at IS NULL`, id); err != nil {
		return Response{}, err
	}
	if result.SourceSettings, err = insertSettings(ctx, tx, id, request.SourceSettings); err != nil {
		return Response{}, err
	}
	if err = tx.Commit(); err != nil {
		return Response{}, err
	}
	return result, nil
}

func (store *Store) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE categories SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	if _, err = tx.ExecContext(ctx, `UPDATE source_settings SET deleted_at=now() WHERE category_id=$1 AND deleted_at IS NULL`, id); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE category_articles SET deleted_at=now() WHERE category_id=$1 AND deleted_at IS NULL`, id); err != nil {
		return err
	}
	return tx.Commit()
}

type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (store *Store) settings(ctx context.Context, source queryer, id uuid.UUID) ([]SourceSetting, error) {
	rows, err := source.QueryContext(ctx, `SELECT id,label,website_input,normalized_host,kind,position FROM source_settings WHERE category_id=$1 AND deleted_at IS NULL ORDER BY position`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var values []SourceSetting
	for rows.Next() {
		var value SourceSetting
		if err := rows.Scan(&value.ID, &value.Label, &value.WebsiteInput, &value.NormalizedHost, &value.Kind, &value.Position); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}
func insertSettings(ctx context.Context, tx *sql.Tx, categoryID uuid.UUID, inputs []SourceSettingInput) ([]SourceSetting, error) {
	values := make([]SourceSetting, 0, len(inputs))
	for _, input := range inputs {
		input.Label = strings.TrimSpace(input.Label)
		input.WebsiteInput = strings.TrimSpace(input.WebsiteInput)
		input.Kind = strings.TrimSpace(input.Kind)
		if input.Label == "" && input.WebsiteInput == "" {
			continue
		}
		if input.Kind == "" {
			input.Kind = "unspecified"
		}
		host := normalizedHost(input.WebsiteInput)
		value := SourceSetting{ID: uuid.New(), Label: input.Label, WebsiteInput: input.WebsiteInput, NormalizedHost: host, Kind: input.Kind, Position: len(values)}
		if _, err := tx.ExecContext(ctx, `INSERT INTO source_settings (id,category_id,label,website_input,normalized_host,kind,position) VALUES ($1,$2,$3,$4,$5,$6,$7)`, value.ID, categoryID, value.Label, value.WebsiteInput, value.NormalizedHost, value.Kind, value.Position); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}
func validate(request Request) error {
	if strings.TrimSpace(request.Name) == "" {
		return errors.New("category name is blank")
	}
	if request.contentLanguageSet && request.ContentLanguage != DefaultContentLanguage {
		return errors.New("unsupported content language")
	}
	if _, err := normalizedContentLanguage(request.ContentLanguage); err != nil {
		return err
	}
	return nil
}
func normalizedContentLanguage(value string) (string, error) {
	if value == "" {
		return DefaultContentLanguage, nil
	}
	if value != DefaultContentLanguage {
		return "", errors.New("unsupported content language")
	}
	return value, nil
}
func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}
func normalizedHost(input string) *string {
	if input == "" {
		return nil
	}
	parsed, err := url.Parse(input)
	if err != nil || parsed.Hostname() == "" {
		parsed, _ = url.Parse("//" + input)
	}
	host := parsed.Hostname()
	if host == "" {
		return nil
	}
	return &host
}

func NewHandler(store *Store, verifier identity.TokenVerifier, allowedEmail string) *http.ServeMux {
	mux := http.NewServeMux()
	require := func(next http.Handler) http.Handler { return httpapi.RequireAllowed(verifier, allowedEmail, next) }
	mux.Handle("GET /v1/categories", require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, err := store.List(r.Context())
		if err != nil {
			httpapi.WriteProblem(w, httpapi.Problem{Title: "Database unavailable", Status: http.StatusServiceUnavailable})
			return
		}
		writeJSON(w, http.StatusOK, values)
	})))
	mux.Handle("POST /v1/categories", require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request Request
		if err := httpapi.DecodeJSON(r, &request); err != nil || validate(request) != nil {
			httpapi.WriteProblem(w, httpapi.ValidationProblem(err))
			return
		}
		value, err := store.Create(r.Context(), request)
		if err != nil {
			httpapi.WriteProblem(w, httpapi.Problem{Title: "Database unavailable", Status: http.StatusServiceUnavailable})
			return
		}
		writeJSON(w, http.StatusCreated, value)
	})))
	mux.Handle("PATCH /v1/categories/{id}", require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpapi.WriteProblem(w, httpapi.ValidationProblem(err))
			return
		}
		var request Request
		if err = httpapi.DecodeJSON(r, &request); err != nil || validate(request) != nil {
			httpapi.WriteProblem(w, httpapi.ValidationProblem(err))
			return
		}
		value, err := store.Update(r.Context(), id, request)
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteProblem(w, httpapi.Problem{Title: "Category not found", Status: http.StatusNotFound})
			return
		}
		if err != nil {
			httpapi.WriteProblem(w, httpapi.Problem{Title: "Database unavailable", Status: http.StatusServiceUnavailable})
			return
		}
		writeJSON(w, http.StatusOK, value)
	})))
	mux.Handle("DELETE /v1/categories/{id}", require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpapi.WriteProblem(w, httpapi.ValidationProblem(err))
			return
		}
		err = store.Delete(r.Context(), id)
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteProblem(w, httpapi.Problem{Title: "Category not found", Status: http.StatusNotFound})
			return
		}
		if err != nil {
			httpapi.WriteProblem(w, httpapi.Problem{Title: "Database unavailable", Status: http.StatusServiceUnavailable})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})))
	return mux
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
