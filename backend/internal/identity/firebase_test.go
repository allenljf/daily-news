package identity

import (
	"context"
	"errors"
	"testing"

	"firebase.google.com/go/v4/auth"
)

func TestFirebaseVerifierMapsClassifiedInvalidToken(t *testing.T) {
	t.Parallel()

	invalid := errors.New("invalid token")
	verifier := newFirebaseVerifier(fakeTokenClient{err: invalid}, func(err error) bool {
		return errors.Is(err, invalid)
	})

	_, err := verifier.Verify(context.Background(), "token")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Verify() error = %v, want ErrInvalidToken", err)
	}
}

func TestFirebaseVerifierPreservesOperationalFailure(t *testing.T) {
	t.Parallel()

	operational := errors.New("certificate service unavailable")
	verifier := newFirebaseVerifier(fakeTokenClient{err: operational}, func(error) bool { return false })

	_, err := verifier.Verify(context.Background(), "token")
	if !errors.Is(err, operational) {
		t.Fatalf("Verify() error = %v, want wrapped operational error", err)
	}
}

func TestFirebaseVerifierRequiresUIDAndEmailClaims(t *testing.T) {
	t.Parallel()

	verifier := newFirebaseVerifier(fakeTokenClient{token: &auth.Token{UID: "user", Claims: map[string]interface{}{}}}, func(error) bool { return false })

	_, err := verifier.Verify(context.Background(), "token")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Verify() error = %v, want ErrInvalidToken", err)
	}
}

type fakeTokenClient struct {
	token *auth.Token
	err   error
}

func (client fakeTokenClient) VerifyIDToken(_ context.Context, _ string) (*auth.Token, error) {
	return client.token, client.err
}
