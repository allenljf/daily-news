// Package news implements Category-scoped Article reads and global lifecycle mutations.
package news

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/allenljf/daily-news/backend/internal/httpapi"
	"github.com/allenljf/daily-news/backend/internal/identity"
	"github.com/google/uuid"
)

const pageSize = 20

var (
	ErrInvalidCursor = errors.New("invalid cursor")
	ErrNotFound      = errors.New("news resource not found")
)

type cursorPosition struct {
	InsertedAt time.Time `json:"inserted_at"`
	ArticleID  uuid.UUID `json:"article_id"`
}
type listItem struct {
	ID             uuid.UUID  `json:"id"`
	Title          string     `json:"title"`
	Summary        *string    `json:"summary"`
	CanonicalURL   string     `json:"canonical_url"`
	PublishedAt    *time.Time `json:"published_at"`
	InsertedAt     time.Time  `json:"inserted_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
	SourceTagID    uuid.UUID  `json:"source_tag_id"`
	SourceTagLabel string     `json:"source_tag_label"`
}
type detail struct {
	listItem
	FirstSeenAt time.Time `json:"first_seen_at"`
}
type page struct {
	Items      []listItem `json:"items"`
	NextCursor *string    `json:"next_cursor"`
}
type article struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Summary      *string    `json:"summary"`
	CanonicalURL string     `json:"canonical_url"`
	PublishedAt  *time.Time `json:"published_at"`
	FirstSeenAt  time.Time  `json:"first_seen_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

func encodeCursor(position cursorPosition) string {
	data, _ := json.Marshal(position)
	return base64.RawURLEncoding.EncodeToString(data)
}
func decodeCursor(value string) (cursorPosition, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return cursorPosition{}, ErrInvalidCursor
	}
	var position cursorPosition
	if err = json.Unmarshal(data, &position); err != nil || position.InsertedAt.IsZero() || position.ArticleID == uuid.Nil {
		return cursorPosition{}, ErrInvalidCursor
	}
	return position, nil
}

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (store *Store) List(ctx context.Context, categoryID uuid.UUID, cursor string, sourceTag *uuid.UUID) (page, error) {
	var exists bool
	if err := store.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM categories WHERE id=$1 AND deleted_at IS NULL)`, categoryID).Scan(&exists); err != nil {
		return page{}, err
	}
	if !exists {
		return page{}, ErrNotFound
	}
	query := `SELECT a.id,a.title,a.summary,a.canonical_url,a.published_at,ca.inserted_at,a.expires_at,ss.id,ss.label FROM category_articles ca JOIN articles a ON a.id=ca.article_id JOIN source_settings ss ON ss.id=ca.source_setting_id WHERE ca.category_id=$1 AND ca.deleted_at IS NULL AND a.deleted_at IS NULL AND (a.expires_at IS NULL OR a.expires_at > now()) AND ss.deleted_at IS NULL`
	args := []any{categoryID}
	if sourceTag != nil {
		args = append(args, *sourceTag)
		query += ` AND ca.source_setting_id=$2`
	}
	if cursor != "" {
		position, err := decodeCursor(cursor)
		if err != nil {
			return page{}, err
		}
		args = append(args, position.InsertedAt, position.ArticleID)
		if sourceTag != nil {
			query += ` AND (ca.inserted_at < $3 OR (ca.inserted_at = $3 AND ca.article_id < $4))`
		} else {
			query += ` AND (ca.inserted_at < $2 OR (ca.inserted_at = $2 AND ca.article_id < $3))`
		}
	}
	query += ` ORDER BY ca.inserted_at DESC,ca.article_id DESC LIMIT ` + strconv.Itoa(pageSize+1)
	rows, err := store.db.QueryContext(ctx, query, args...)
	if err != nil {
		return page{}, err
	}
	defer rows.Close()
	values := []listItem{}
	for rows.Next() {
		var value listItem
		if err := rows.Scan(&value.ID, &value.Title, &value.Summary, &value.CanonicalURL, &value.PublishedAt, &value.InsertedAt, &value.ExpiresAt, &value.SourceTagID, &value.SourceTagLabel); err != nil {
			return page{}, err
		}
		values = append(values, value)
	}
	if err = rows.Err(); err != nil {
		return page{}, err
	}
	result := page{Items: values}
	if len(values) > pageSize {
		result.Items = values[:pageSize]
		next := encodeCursor(cursorPosition{InsertedAt: result.Items[pageSize-1].InsertedAt, ArticleID: result.Items[pageSize-1].ID})
		result.NextCursor = &next
	}
	return result, nil
}

func (store *Store) Detail(ctx context.Context, categoryID, articleID uuid.UUID) (detail, error) {
	var value detail
	err := store.db.QueryRowContext(ctx, `SELECT a.id,a.title,a.summary,a.canonical_url,a.published_at,ca.inserted_at,a.expires_at,ss.id,ss.label,a.first_seen_at FROM category_articles ca JOIN articles a ON a.id=ca.article_id JOIN source_settings ss ON ss.id=ca.source_setting_id WHERE ca.category_id=$1 AND ca.article_id=$2 AND ca.deleted_at IS NULL AND a.deleted_at IS NULL AND (a.expires_at IS NULL OR a.expires_at>now()) AND ss.deleted_at IS NULL`, categoryID, articleID).Scan(&value.ID, &value.Title, &value.Summary, &value.CanonicalURL, &value.PublishedAt, &value.InsertedAt, &value.ExpiresAt, &value.SourceTagID, &value.SourceTagLabel, &value.FirstSeenAt)
	if errors.Is(err, sql.ErrNoRows) {
		return detail{}, ErrNotFound
	}
	return value, err
}
func (store *Store) SetPermanent(ctx context.Context, id uuid.UUID, permanent bool) (article, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return article{}, err
	}
	defer tx.Rollback()
	var value article
	err = tx.QueryRowContext(ctx, `SELECT id,title,summary,canonical_url,published_at,first_seen_at,expires_at FROM articles WHERE id=$1 AND deleted_at IS NULL FOR UPDATE`, id).Scan(&value.ID, &value.Title, &value.Summary, &value.CanonicalURL, &value.PublishedAt, &value.FirstSeenAt, &value.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return article{}, ErrNotFound
	}
	if err != nil {
		return article{}, err
	}
	if permanent {
		value.ExpiresAt = nil
	} else {
		expires := value.FirstSeenAt.AddDate(0, 0, 30)
		value.ExpiresAt = &expires
	}
	if _, err = tx.ExecContext(ctx, `UPDATE articles SET expires_at=$2 WHERE id=$1`, id, value.ExpiresAt); err != nil {
		return article{}, err
	}
	if err = tx.Commit(); err != nil {
		return article{}, err
	}
	return value, nil
}
func (store *Store) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := store.db.ExecContext(ctx, `UPDATE articles SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
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
	return nil
}

func NewHandler(store *Store, verifier identity.TokenVerifier, allowedEmail string) *http.ServeMux {
	mux := http.NewServeMux()
	require := func(next http.Handler) http.Handler { return httpapi.RequireAllowed(verifier, allowedEmail, next) }
	mux.Handle("GET /v1/categories/{categoryID}/news", require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		categoryID, err := uuid.Parse(r.PathValue("categoryID"))
		if err != nil {
			problem(w, http.StatusUnprocessableEntity, "Validation failed")
			return
		}
		var source *uuid.UUID
		if value := r.URL.Query().Get("sourceTagId"); value != "" {
			parsed, err := uuid.Parse(value)
			if err != nil {
				problem(w, http.StatusUnprocessableEntity, "Validation failed")
				return
			}
			source = &parsed
		}
		result, err := store.List(r.Context(), categoryID, r.URL.Query().Get("cursor"), source)
		if errors.Is(err, ErrInvalidCursor) {
			problem(w, http.StatusBadRequest, "Invalid cursor")
			return
		}
		if errors.Is(err, ErrNotFound) {
			problem(w, http.StatusNotFound, "Category not found")
			return
		}
		if err != nil {
			problem(w, http.StatusServiceUnavailable, "Database unavailable")
			return
		}
		jsonResponse(w, http.StatusOK, result)
	})))
	mux.Handle("GET /v1/categories/{categoryID}/news/{articleID}", require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		categoryID, err := uuid.Parse(r.PathValue("categoryID"))
		if err != nil {
			problem(w, http.StatusUnprocessableEntity, "Validation failed")
			return
		}
		articleID, err := uuid.Parse(r.PathValue("articleID"))
		if err != nil {
			problem(w, http.StatusUnprocessableEntity, "Validation failed")
			return
		}
		value, err := store.Detail(r.Context(), categoryID, articleID)
		if errors.Is(err, ErrNotFound) {
			problem(w, http.StatusNotFound, "Article not found")
			return
		}
		if err != nil {
			problem(w, http.StatusServiceUnavailable, "Database unavailable")
			return
		}
		jsonResponse(w, http.StatusOK, value)
	})))
	mux.Handle("PATCH /v1/news/{articleID}", require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("articleID"))
		if err != nil {
			problem(w, http.StatusUnprocessableEntity, "Validation failed")
			return
		}
		var request struct {
			Permanent *bool `json:"permanent"`
		}
		if err = httpapi.DecodeJSON(r, &request); err != nil || request.Permanent == nil {
			problem(w, http.StatusUnprocessableEntity, "Validation failed")
			return
		}
		value, err := store.SetPermanent(r.Context(), id, *request.Permanent)
		if errors.Is(err, ErrNotFound) {
			problem(w, http.StatusNotFound, "Article not found")
			return
		}
		if err != nil {
			problem(w, http.StatusServiceUnavailable, "Database unavailable")
			return
		}
		jsonResponse(w, http.StatusOK, value)
	})))
	mux.Handle("DELETE /v1/news/{articleID}", require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("articleID"))
		if err != nil {
			problem(w, http.StatusUnprocessableEntity, "Validation failed")
			return
		}
		err = store.Delete(r.Context(), id)
		if errors.Is(err, ErrNotFound) {
			problem(w, http.StatusNotFound, "Article not found")
			return
		}
		if err != nil {
			problem(w, http.StatusServiceUnavailable, "Database unavailable")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})))
	return mux
}
func problem(w http.ResponseWriter, status int, title string) {
	httpapi.WriteProblem(w, httpapi.Problem{Title: title, Status: status})
}
func jsonResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
