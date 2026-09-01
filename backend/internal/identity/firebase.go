// Package identity verifies Firebase identities and applies Daily News policy.
package identity

import (
	"context"
	"errors"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
)

// ErrInvalidToken indicates that Firebase rejected a supplied ID token or its
// required identity claims were absent.
var ErrInvalidToken = errors.New("invalid Firebase ID token")

// Identity is the verified Firebase information required by Daily News.
type Identity struct {
	UID   string
	Email string
}

// TokenVerifier verifies a Firebase ID token without coupling HTTP to Firebase.
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (Identity, error)
}

type firebaseTokenClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

// FirebaseVerifier verifies tokens using Firebase Admin.
type FirebaseVerifier struct {
	client    firebaseTokenClient
	isInvalid func(error) bool
}

// NewFirebaseVerifier initializes Firebase Admin with Application Default Credentials.
func NewFirebaseVerifier(ctx context.Context) (*FirebaseVerifier, error) {
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase Admin: %w", err)
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase Auth: %w", err)
	}
	return newFirebaseVerifier(client, auth.IsIDTokenInvalid), nil
}

func newFirebaseVerifier(client firebaseTokenClient, isInvalid func(error) bool) *FirebaseVerifier {
	return &FirebaseVerifier{client: client, isInvalid: isInvalid}
}

// Verify validates a Firebase ID token and extracts its required claims.
func (verifier *FirebaseVerifier) Verify(ctx context.Context, token string) (Identity, error) {
	verifiedToken, err := verifier.client.VerifyIDToken(ctx, token)
	if err != nil {
		if verifier.isInvalid(err) {
			return Identity{}, ErrInvalidToken
		}
		return Identity{}, fmt.Errorf("verify Firebase ID token: %w", err)
	}

	email, ok := verifiedToken.Claims["email"].(string)
	if verifiedToken.UID == "" || !ok || email == "" {
		return Identity{}, ErrInvalidToken
	}
	return Identity{UID: verifiedToken.UID, Email: email}, nil
}
