//go:build integration

package clinics

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
	"uuid"

	"mundoappointment.com/pkg/config"
)

func newClinicIntegrationStore(t *testing.T) *store {
	t.Helper()
	if os.Getenv("RUN_SUPABASE_INTEGRATION") != "true" {
		t.Skip("integration tests are disabled; set RUN_SUPABASE_INTEGRATION=true")
	}
	testURL, testKey := os.Getenv("SUPABASE_TEST_URL"), os.Getenv("SUPABASE_TEST_KEY")
	if testURL == "" || testKey == "" {
		t.Fatal("SUPABASE_TEST_URL and SUPABASE_TEST_KEY are required; use a test database")
	}
	t.Setenv("SUPABASE_URL", testURL)
	t.Setenv("SUPABASE_KEY", testKey)
	db, err := config.NewDBClient()
	if err != nil {
		t.Fatalf("creating test DB client: %v", err)
	}
	return NewStore(db)
}

func TestStoreClinicCRUD(t *testing.T) {
	s := newClinicIntegrationStore(t)
	createFixture := func(name string) Clinic {
		t.Helper()
		created, err := s.createClinic(CreateClinicRequest{Name: name, Timezone: "UTC"})
		if err != nil {
			t.Fatalf("creating clinic: %v", err)
		}
		clinic := created
		if clinic.Id != nil && *clinic.Id != uuid.Nil() {
			id := clinic.Id.String()
			t.Cleanup(func() {
				if _, err := s.deleteClinic(id); err != nil && !errors.Is(err, ErrorClinicNotFound) {
					t.Errorf("cleaning up clinic: %v", err)
				}
			})
		}
		if clinic.Id == nil || *clinic.Id == (uuid.UUID{}) {
			t.Fatal("expected a nonzero database-generated clinic UUID")
		}
		if clinic.Name != name || clinic.Timezone != "UTC" || clinic.Status != "Active" {
			t.Fatal("created clinic fields or default Active status do not match")
		}
		return clinic
	}

	stamp := time.Now().UnixNano()
	created := createFixture(fmt.Sprintf("Integration clinic A %d", stamp))
	other := createFixture(fmt.Sprintf("Integration clinic B %d", stamp))
	if *created.Id == *other.Id {
		t.Fatal("different clinics must have different generated IDs")
	}
	id := created.Id.String()
	assertClinic := func(got, want Clinic) {
		t.Helper()
		if got.Id == nil || want.Id == nil || *got.Id != *want.Id || got.Name != want.Name || got.Status != want.Status || got.Timezone != want.Timezone {
			t.Fatal("clinic fields do not match expected fixture")
		}
	}
	fetched, err := s.fetchClinic(id)
	if err != nil {
		t.Fatalf("fetching clinic: %v", err)
	}
	assertClinic(fetched, created)

	// Each PATCH changes one field and must preserve the others, including the ID.
	name, timezone, status := fmt.Sprintf("Updated clinic %d", stamp), "America/Mexico_City", "Inactive"
	want := created
	for _, patch := range []UpdateClinicRequest{{Name: &name}, {Timezone: &timezone}, {Status: &status}} {
		if patch.Name != nil {
			want.Name = *patch.Name
		}
		if patch.Timezone != nil {
			want.Timezone = *patch.Timezone
		}
		if patch.Status != nil {
			want.Status = *patch.Status
		}
		updated, err := s.updateClinic(patch, id)
		if err != nil {
			t.Fatalf("updating clinic: %v", err)
		}
		assertClinic(updated, want)
		persisted, err := s.fetchClinic(id)
		if err != nil {
			t.Fatalf("reading persisted clinic update: %v", err)
		}
		assertClinic(persisted, want)
	}
	untouched, err := s.fetchClinic(other.Id.String())
	if err != nil {
		t.Fatalf("fetching untouched clinic: %v", err)
	}
	assertClinic(untouched, other)

	all, err := s.fetchClinics()
	if err != nil {
		t.Fatalf("listing clinics: %v", err)
	}
	found := map[string]bool{}
	for _, clinic := range all {
		if clinic.Id != nil {
			found[clinic.Id.String()] = true
		}
	}
	if !found[id] || !found[other.Id.String()] {
		t.Fatal("expected both test clinics in list response")
	}

	deletedID, err := s.deleteClinic(id)
	if err != nil {
		t.Fatalf("deleting clinic: %v", err)
	}
	if deletedID != id {
		t.Fatalf("deleted ID: got %q, want %q", deletedID, id)
	}
	_, err = s.fetchClinic(id)
	if !errors.Is(err, ErrorClinicNotFound) {
		t.Fatalf("fetch after delete: expected ErrorClinicNotFound, got %v", err)
	}
	_, err = s.updateClinic(UpdateClinicRequest{Status: &status}, id)
	if !errors.Is(err, ErrorClinicNotFound) {
		t.Fatalf("update after delete: expected ErrorClinicNotFound, got %v", err)
	}
	_, err = s.deleteClinic(id)
	if !errors.Is(err, ErrorClinicNotFound) {
		t.Fatalf("delete after delete: expected ErrorClinicNotFound, got %v", err)
	}
	untouched, err = s.fetchClinic(other.Id.String())
	if err != nil {
		t.Fatalf("deleting one clinic must not delete another: %v", err)
	}
	assertClinic(untouched, other)
}
