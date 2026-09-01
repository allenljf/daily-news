package news

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCursorRoundTripPreservesKeysetPosition(t *testing.T) {
	t.Parallel()

	position := cursorPosition{InsertedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), ArticleID: uuid.New()}
	decoded, err := decodeCursor(encodeCursor(position))
	if err != nil {
		t.Fatal(err)
	}
	if decoded != position {
		t.Fatalf("cursor = %#v, want %#v", decoded, position)
	}
}

func TestDecodeCursorRejectsMalformedValue(t *testing.T) {
	t.Parallel()
	if _, err := decodeCursor("not-a-cursor"); err == nil {
		t.Fatal("decodeCursor() error = nil")
	}
}
