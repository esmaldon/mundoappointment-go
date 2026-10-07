//go:build integration

package patients

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
	"uuid"

	"mundoappointment.com/pkg/config"
)

func TestStorePatientCRUD(t *testing.T) {
	s := newIntegrationStore(t)

	clinicUUID := createIntegrationClinic(t, s)
	otherClinicUUID := createIntegrationClinic(t, s)
	input := Patient{
		ClinicId: clinicUUID, Status: "Active", AdmissionDate: time.Now().Format("2006-01-02"),
		FirstName: "Integration",
		LastName:  "Test",
		Birthday:  "1990-01-01",
		Phone:     "5555555555",
		Email:     fmt.Sprintf("integration-%d@example.com", time.Now().UnixNano()),
	}

	created, err := s.createPatient(input, clinicUUID)
	if err != nil {
		t.Fatalf("error creating patient. %v", err)
	}
	if created.Id == nil {
		t.Fatalf("expected id in patient created")
	}
	id := created.Id.String()
	clinicId := created.ClinicId.String()
	deleted := false
	t.Cleanup(func() {
		if deleted {
			return
		}

		_, cleanupErr := s.deletePatient(id, clinicId)
		if cleanupErr != nil && !errors.Is(cleanupErr, ErrorPatientNotFound) {
			t.Errorf("expected cleanup of patient created. %v", cleanupErr)
		}
	})

	fetched, err := s.fetchPatient(id, clinicId)
	if err != nil {
		t.Fatalf("error during fetch of patient. %v", err)
	}

	if fetched.Email != input.Email {
		t.Errorf("expected email %q, got %q", input.Email, fetched.Email)
	}

	// A patient cannot be read, changed or deleted through another clinic.
	if _, err := s.fetchPatient(id, otherClinicUUID.String()); !errors.Is(err, ErrorPatientNotFound) {
		t.Fatalf("cross-clinic fetch: %v", err)
	}
	wrongPhone := "1111111111"
	if _, err := s.updatePatient(UpdatePatientRequest{Phone: &wrongPhone}, id, otherClinicUUID.String()); !errors.Is(err, ErrorPatientNotFound) {
		t.Fatalf("cross-clinic update: %v", err)
	}
	if _, err := s.deletePatient(id, otherClinicUUID.String()); !errors.Is(err, ErrorPatientNotFound) {
		t.Fatalf("cross-clinic delete: %v", err)
	}

	newPhone := "9999999999"
	updateReq := UpdatePatientRequest{
		Phone: &newPhone,
	}
	updated, err := s.updatePatient(updateReq, id, clinicId)
	if err != nil {
		t.Fatalf("error during update of patient. %v", err)
	}

	if updated.Phone != "9999999999" {
		t.Errorf("expected phone 9999999999, got %q", updated.Phone)
	}

	allPatients, err := s.fetchPatients(clinicId)
	if err != nil {
		t.Fatalf("error during fetch all patients. %v", err)
	}
	found := false
	for _, patient := range allPatients {
		if patient.Id != nil && *patient.Id == *created.Id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected patient created during fetch all patients")
	}

	deletedId, err := s.deletePatient(id, clinicId)
	if err != nil {
		t.Fatalf("error during delete of patient. %v", err)
	}
	if deletedId != id {
		t.Errorf("expected id %q, got %q", id, deletedId)
	}
	deleted = true

	_, err = s.fetchPatient(id, clinicId)
	if !errors.Is(err, ErrorPatientNotFound) {
		t.Fatalf("expected ErrorPatientNotFound after deletion. got %v", err)
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

func createIntegrationClinic(t *testing.T, s *store) uuid.UUID {
	t.Helper()
	type clinicFixture struct {
		ID       *uuid.UUID `json:"id,omitempty"`
		Name     string     `json:"name"`
		Status   string     `json:"status"`
		Timezone string     `json:"timezone"`
	}
	clinic, err := s.db.Create[clinicFixture]("clinics", clinicFixture{Name: fmt.Sprintf("Patient test %d", time.Now().UnixNano()), Status: "Active", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("creating test clinic: %v", err)
	}
	if clinic.ID == nil || *clinic.ID == uuid.Nil() {
		t.Fatal("missing clinic UUID")
	}
	t.Cleanup(func() {
		if _, err := s.db.Delete("clinics", clinic.ID.String()); err != nil && !errors.Is(err, config.ErrorRecordNotFound) {
			t.Errorf("cleaning up test clinic: %v", err)
		}
	})
	return *clinic.ID
}

func TestStoreMinorPatientAndParentCRUD(t *testing.T) {
	s := newIntegrationStore(t)
	clinicID := createIntegrationClinic(t, s)
	req := minorRequest()
	req.Email = fmt.Sprintf("minor-%d@example.com", time.Now().UnixNano())
	service := NewService(s)
	p, err := service.addPatient(req, clinicID.String())
	if err != nil {
		t.Fatalf("creating minor and parent: %v", err)
	}
	if p.ParentId != nil {
		parentID := p.ParentId.String()
		t.Cleanup(func() {
			if _, err := s.deleteParent(parentID); err != nil && !errors.Is(err, ErrorParentNotFound) {
				t.Errorf("cleaning up parent: %v", err)
			}
		})
	}
	if p.Id != nil {
		patientID := p.Id.String()
		t.Cleanup(func() {
			if _, err := s.deletePatient(patientID, clinicID.String()); err != nil && !errors.Is(err, ErrorPatientNotFound) {
				t.Errorf("cleaning up patient: %v", err)
			}
		})
	}
	if p.Id == nil || p.ParentId == nil {
		t.Fatal("missing patient or parent UUID")
	}
	parent, err := s.db.FetchById[Parent](parentTableName, p.ParentId.String())
	if err != nil || parent.Birthday != *req.ParentBirthday {
		t.Fatalf("parent birthday=%q error=%v", parent.Birthday, err)
	}
	name := "Updated parent"
	updated, err := service.changeParent(UpdateParentRequest{ParentFirstName: &name}, p.ParentId.String(), p.Id.String(), clinicID.String())
	if err != nil || updated.FirstName != name {
		t.Fatalf("parent update=%+v error=%v", updated, err)
	}
	fetched, err := s.db.FetchById[Parent](parentTableName, p.ParentId.String())
	if err != nil || fetched.FirstName != name || fetched.Birthday != *req.ParentBirthday {
		t.Fatalf("persisted parent=%+v error=%v", fetched, err)
	}
}

func TestStoreParentCRUD(t *testing.T) {
	s := newIntegrationStore(t)
	p, err := s.createParent(Parent{FirstName: "Integration", LastName: "Parent", Birthday: "1980-02-03", Status: "Active"})
	if err != nil {
		t.Fatalf("creating parent: %v", err)
	}
	if p.Id == nil {
		t.Fatal("missing parent ID")
	}
	id := p.Id.String()
	t.Cleanup(func() {
		if _, err := s.deleteParent(id); err != nil && !errors.Is(err, ErrorParentNotFound) {
			t.Errorf("cleaning up parent: %v", err)
		}
	})
	if p.FirstName != "Integration" || p.LastName != "Parent" || p.Birthday != "1980-02-03" {
		t.Fatalf("unexpected parent %+v", p)
	}
	updated, err := s.updateParent(UpdateParentRequest{ParentLastName: stringPointer("Updated")}, id)
	if err != nil || updated.LastName != "Updated" || updated.FirstName != p.FirstName {
		t.Fatalf("updated parent %+v error=%v", updated, err)
	}
	if _, err := s.deleteParent(id); err != nil {
		t.Fatal(err)
	}
	if _, err := s.updateParent(UpdateParentRequest{ParentFirstName: stringPointer("Missing")}, id); !errors.Is(err, ErrorParentNotFound) {
		t.Fatalf("missing parent: %v", err)
	}
}
