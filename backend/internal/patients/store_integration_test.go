//go:build integration

package patients

import (
	"errors"
	"os"
	"strconv"
	"testing"

	"mundoappointment.com/pkg/config"
)

func TestStorePatientCRUD(t *testing.T) {
	s := newIntegrationStore(t)

	input := Patient{
		FirstName:     "Integration",
		LastName:      "Test",
		Birthday:      "1990-01-01",
		Phone:         "5555555555",
		Email:         "integration@example.com",
		Status:        "Active",
		AdmissionDate: "2026-01-01",
	}

	created, err := s.createPatient(input)
	if err != nil {
		t.Fatalf("error creating patient. %v", err)
	}
	if len(created) != 1 {
		t.Fatalf("expected 1 record created got %d", len(created))
	}
	if created[0].Id == nil {
		t.Fatalf("expected id in patient created")
	}
	id := strconv.Itoa(*created[0].Id)
	deleted := false
	t.Cleanup(func() {
		if deleted {
			return
		}

		_, cleanupErr := s.deletePatient(id)
		if cleanupErr != nil && !errors.Is(cleanupErr, ErrorPatientNotFound) {
			t.Errorf("expected cleanup of patient created. %v", cleanupErr)
		}
	})

	fetched, err := s.fetchPatient(id)
	if err != nil {
		t.Fatalf("error during fetch of patient. %v", err)
	}

	if fetched.Email != input.Email {
		t.Errorf("expected email %q, got %q", input.Email, fetched.Email)
	}

	fetched.Phone = "9999999999"
	updated, err := s.updatePatient(id, fetched)
	if err != nil {
		t.Fatalf("error during update of patient. %v", err)
	}

	if updated.Phone != "9999999999" {
		t.Errorf("expected phone 9999999999, got %q", updated.Phone)
	}

	allPatients, err := s.fetchPatients()
	if err != nil {
		t.Fatalf("error during fetch all patients. %v", err)
	}
	found := false
	for _, patient := range allPatients {
		if patient.Id != nil && *patient.Id == *created[0].Id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected patient created during fetch all patients")
	}

	deletedId, err := s.deletePatient(id)
	if err != nil {
		t.Fatalf("error during delete of patient. %v", err)
	}
	if deletedId != id {
		t.Errorf("expected id %q, got %q", id, deletedId)
	}
	deleted = true

	_, err = s.fetchPatient(id)
	if err != nil && !errors.Is(err, ErrorPatientNotFound) {
		t.Fatalf("error during fetch patient by id. %v", err)
	}
}

func newIntegrationStore(t *testing.T) *store {
	t.Helper()

	if os.Getenv("RUN_SUPABASE_INTEGRATION") != "true" {
		t.Skip("integration test are disabled")
	}

	testUrl := os.Getenv("SUPABASE_TEST_URL")
	testKey := os.Getenv("SUPABASE_TEST_KEY")

	if testUrl == "" || testKey == "" {
		t.Fatal("SUPABASE_TEST_URL and SUPABASE_TEST_KEY are required")
	}

	t.Setenv("SUPABASE_URL", testUrl)
	t.Setenv("SUPABASE_KEY", testKey)

	db, err := config.NewDBClient()
	if err != nil {
		t.Fatalf("error creating db client. %v", err)
	}

	return NewStore(db)
}
