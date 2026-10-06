package patients

import (
	"errors"
	"testing"
	"time"
	"uuid"
)

// Embedding the interface makes unexpected calls fail immediately.
type patientStoreStub struct {
	patientStore
	create           func(Patient, uuid.UUID) (Patient, error)
	parent           func(Parent) (Parent, error)
	get              func(string, string) (Patient, error)
	patchParent      func(UpdateParentRequest, string) (Parent, error)
	deleteParentFunc func(string) (string, error)
}

func (s patientStoreStub) createPatient(p Patient, id uuid.UUID) (Patient, error) {
	return s.create(p, id)
}
func (s patientStoreStub) createParent(p Parent) (Parent, error)           { return s.parent(p) }
func (s patientStoreStub) fetchPatient(id, clinic string) (Patient, error) { return s.get(id, clinic) }
func (s patientStoreStub) updateParent(req UpdateParentRequest, id string) (Parent, error) {
	return s.patchParent(req, id)
}
func (s patientStoreStub) deleteParent(id string) (string, error) { return s.deleteParentFunc(id) }
func stringPointer(s string) *string                              { return &s }
func minorRequest() CreatePatientRequest {
	return CreatePatientRequest{FirstName: "Child", LastName: "Test", Birthday: time.Now().AddDate(-10, 0, 0).Format("2006-01-02"), Phone: "5555555555", Email: "test@example.com", ParentFirstName: stringPointer("Parent"), ParentLastName: stringPointer("Test"), ParentBirthday: stringPointer("1980-02-03")}
}
func TestIsMinorAt(t *testing.T) {
	today := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name, birthday string
		minor, fail    bool
	}{
		{"birthday tomorrow", "2008-10-06", true, false}, {"birthday today", "2008-10-05", false, false},
		{"birthday yesterday", "2008-10-04", false, false}, {"child", "2016-01-01", true, false},
		{"future", "2026-10-06", false, true}, {"invalid", "bad", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := isMinorAt(tc.birthday, today)
			if (err != nil) != tc.fail || got != tc.minor {
				t.Fatalf("minor=%t error=%v", got, err)
			}
		})
	}
}
func TestAddPatientAdult(t *testing.T) {
	req := minorRequest()
	req.Birthday = "1990-01-01"
	s := NewService(patientStoreStub{create: func(p Patient, c uuid.UUID) (Patient, error) {
		if p.ParentId != nil || p.ClinicId.String() != testClinicID || c != p.ClinicId || p.Status != "Active" || p.AdmissionDate == "" || p.FirstName != req.FirstName {
			t.Fatalf("unexpected patient: %+v", p)
		}
		p.Id = uuidPointer(testPatientID)
		return p, nil
	}})
	p, err := s.addPatient(req, testClinicID)
	if err != nil || p.Id == nil {
		t.Fatalf("patient=%+v error=%v", p, err)
	}
}
func TestAddPatientMinor(t *testing.T) {
	req := minorRequest()
	parentID := uuid.New()
	order := 0
	s := NewService(patientStoreStub{parent: func(p Parent) (Parent, error) {
		if order != 0 || p.FirstName != *req.ParentFirstName || p.LastName != *req.ParentLastName || p.Birthday != *req.ParentBirthday || p.Status != "Active" {
			t.Fatalf("unexpected parent %+v", p)
		}
		order++
		p.Id = &parentID
		return p, nil
	}, create: func(p Patient, c uuid.UUID) (Patient, error) {
		if order != 1 || p.ParentId == nil || *p.ParentId != parentID || c.String() != testClinicID {
			t.Fatalf("unexpected patient %+v", p)
		}
		order++
		return p, nil
	}})
	if _, err := s.addPatient(req, testClinicID); err != nil || order != 2 {
		t.Fatalf("order=%d error=%v", order, err)
	}
}
func TestRejectIncompleteParentBeforeWriting(t *testing.T) {
	for mask := 0; mask < 7; mask++ {
		req := minorRequest()
		if mask&1 == 0 {
			req.ParentFirstName = nil
		}
		if mask&2 == 0 {
			req.ParentLastName = nil
		}
		if mask&4 == 0 {
			req.ParentBirthday = nil
		}
		s := NewService(patientStoreStub{})
		if _, err := s.addPatient(req, testClinicID); !errors.Is(err, ErrorPatientBadReq) {
			t.Fatalf("mask %d: %v", mask, err)
		}
		if _, err := s.addParent(req); !errors.Is(err, ErrorParentEmpty) {
			t.Fatalf("mask %d: %v", mask, err)
		}
	}
}
func TestInvalidPatientInputBeforeWriting(t *testing.T) {
	s := NewService(patientStoreStub{})
	if _, err := s.addPatient(minorRequest(), "bad"); !errors.Is(err, ErrorPatientBadReq) {
		t.Fatal(err)
	}
	req := minorRequest()
	req.Birthday = "invalid"
	if _, err := s.addPatient(req, testClinicID); !errors.Is(err, ErrorPatientBadReq) {
		t.Fatal(err)
	}
}
func TestParentCreationFailureStopsPatient(t *testing.T) {
	failure := errors.New("insert failed")
	for _, tc := range []struct {
		name string
		p    Parent
		err  error
	}{{"store error", Parent{}, failure}, {"missing ID", Parent{}, nil}, {"zero ID", Parent{Id: new(uuid.UUID)}, nil}} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(patientStoreStub{parent: func(Parent) (Parent, error) { return tc.p, tc.err }})
			if _, err := s.addPatient(minorRequest(), testClinicID); err == nil {
				t.Fatal("expected failure")
			}
		})
	}
}
func TestPatientFailureCleansUpParent(t *testing.T) {
	for _, cleanupFails := range []bool{false, true} {
		failure, cleanupFailure := errors.New("patient insert failed"), errors.New("cleanup failed")
		id := uuid.New()
		deleted := false
		s := NewService(patientStoreStub{parent: func(p Parent) (Parent, error) { p.Id = &id; return p, nil }, create: func(Patient, uuid.UUID) (Patient, error) { return Patient{}, failure }, deleteParentFunc: func(got string) (string, error) {
			if got != id.String() {
				t.Fatal(got)
			}
			deleted = true
			if cleanupFails {
				return "", cleanupFailure
			}
			return got, nil
		}})
		_, err := s.addPatient(minorRequest(), testClinicID)
		if !deleted || !errors.Is(err, failure) || (cleanupFails && !errors.Is(err, cleanupFailure)) {
			t.Fatalf("deleted=%t error=%v", deleted, err)
		}
	}
}
func TestChangeParentEnforcesPatientAndClinic(t *testing.T) {
	id := uuid.New()
	dbError := errors.New("database unavailable")
	for _, tc := range []struct {
		name           string
		p              Patient
		err, errorWant error
		update         bool
	}{
		{"matching", Patient{ParentId: &id}, nil, nil, true},
		{"other parent", Patient{ParentId: uuidPointer(testPatientID)}, nil, ErrorParentNotFound, false},
		{"no parent", Patient{}, nil, ErrorParentNotFound, false},
		{"other clinic or patient", Patient{}, ErrorPatientNotFound, ErrorPatientNotFound, false},
		{"store failure", Patient{}, dbError, dbError, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			s := NewService(patientStoreStub{get: func(patient, clinic string) (Patient, error) {
				if patient != testPatientID || clinic != testClinicID {
					t.Fatal("scope lost")
				}
				return tc.p, tc.err
			}, patchParent: func(req UpdateParentRequest, parent string) (Parent, error) {
				called = true
				if parent != id.String() || req.ParentFirstName == nil {
					t.Fatal("incorrect patch")
				}
				return Parent{Id: &id}, nil
			}})
			_, err := s.changeParent(UpdateParentRequest{ParentFirstName: stringPointer("Updated")}, id.String(), testPatientID, testClinicID)
			if !errors.Is(err, tc.errorWant) || called != tc.update {
				t.Fatalf("updated=%t error=%v", called, err)
			}
		})
	}
}
