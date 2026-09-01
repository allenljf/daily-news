// Command api runs the Daily News HTTP API for Cloud Run.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/allenljf/daily-news/backend/internal/api"
	"github.com/allenljf/daily-news/backend/internal/httpapi"
	"github.com/allenljf/daily-news/backend/internal/identity"
	"github.com/allenljf/daily-news/backend/internal/platform"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	address, err := platform.APIAddress(os.Getenv("PORT"))
	if err != nil {
		return err
	}
	database, err := platform.OpenDB(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer database.Close()
	verifier, err := identity.NewFirebaseVerifier(context.Background())
	if err != nil {
		return err
	}
	launcher, err := platform.NewCloudRunLauncher(context.Background(), os.Getenv("GCP_PROJECT_ID"), os.Getenv("GCP_REGION"), os.Getenv("CLOUD_RUN_JOB_NAME"))
	if err != nil {
		return err
	}

	server := httpapi.NewServerWithHandler(address, api.NewHandler(database, verifier, os.Getenv("ALLOWED_USER_EMAIL"), launcher))
	shutdownSignalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-shutdownSignalContext.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownContext)
	}
}
