package platform

import "testing"

func TestAPIAddressUsesCloudRunPort(t *testing.T) {
	t.Parallel()

	got, err := APIAddress("9090")
	if err != nil {
		t.Fatalf("APIAddress() error = %v", err)
	}
	if want := "0.0.0.0:9090"; got != want {
		t.Fatalf("APIAddress() = %q, want %q", got, want)
	}
}

func TestAPIAddressDefaultsToCloudRunPort(t *testing.T) {
	t.Parallel()

	got, err := APIAddress("")
	if err != nil {
		t.Fatalf("APIAddress() error = %v", err)
	}
	if want := "0.0.0.0:8080"; got != want {
		t.Fatalf("APIAddress() = %q, want %q", got, want)
	}
}

func TestAPIAddressRejectsInvalidPort(t *testing.T) {
	t.Parallel()

	if _, err := APIAddress("not-a-port"); err == nil {
		t.Fatal("APIAddress() error = nil, want error")
	}
}
