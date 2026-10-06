package patients

import (
	"errors"
	"fmt"
	"time"
	"uuid"
)

var ErrorClinicIdEmpty = errors.New("clinic id is empty")
var ErrorPatientIdEmpty = errors.New("patient id is empty")
var ErrorPatientBadReq = errors.New("patient bad request")

type patientStore interface {
	fetchPatients(string) ([]Patient, error)
	fetchPatient(string, string) (Patient, error)
	createPatient(Patient, uuid.UUID) (Patient, error)
	createParent(Parent) (Parent, error)
	updatePatient(UpdatePatientRequest, string, string) (Patient, error)
	updateParent(UpdateParentRequest, string) (Parent, error)
	deletePatient(string, string) (string, error)
	deleteParent(string) (string, error)
}

type service struct {
	s patientStore
}

func NewService(ps patientStore) *service {
	return &service{
		s: ps,
	}
}

func (s *service) getPatients(clinicId string) ([]Patient, error) {
	if clinicId == "" {
		return nil, ErrorClinicIdEmpty
	}
	return s.s.fetchPatients(clinicId)
}

func (s *service) getPatient(patientId, clinicId string) (Patient, error) {
	if clinicId == "" {
		return Patient{}, ErrorClinicIdEmpty
	}
	if patientId == "" {
		return Patient{}, ErrorPatientIdEmpty
	}
	return s.s.fetchPatient(patientId, clinicId)
}

func (s *service) addPatient(req CreatePatientRequest, clinicId string) (Patient, error) {
	cId, err := uuid.Parse(clinicId)
	if err != nil {
		return Patient{}, fmt.Errorf("%w: invalid clinic ID", ErrorPatientBadReq)
	}
	newPatient := Patient{
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Birthday:      req.Birthday,
		Phone:         req.Phone,
		Email:         req.Email,
		Status:        "Active",
		AdmissionDate: time.Now().Format("2006-01-02"),
		ClinicId:      cId,
		ParentId:      nil,
	}
	minor, err := isMinor(req.Birthday)
	if err != nil {
		return Patient{}, fmt.Errorf("%w: invalid birthday", ErrorPatientBadReq)
	}
	if minor {
		if !req.HasCompleteParent() {
			return Patient{}, ErrorPatientBadReq
		}

		newParent, err := s.addParent(req)
		if err != nil {
			return Patient{}, err
		}
		if newParent.Id == nil || *newParent.Id == uuid.Nil() {
			return Patient{}, errors.New("created parent has no ID")
		}
		newPatient.ParentId = newParent.Id
	}
	p, err := s.s.createPatient(newPatient, cId)
	if err != nil {
		if newPatient.ParentId != nil {
			if _, cleanupErr := s.s.deleteParent(newPatient.ParentId.String()); cleanupErr != nil {
				return Patient{}, errors.Join(err, fmt.Errorf("cleaning up parent: %w", cleanupErr))
			}
		}
		return Patient{}, err
	}
	return p, nil
}

func (s *service) addParent(req CreatePatientRequest) (Parent, error) {
	if !req.HasCompleteParent() {
		return Parent{}, ErrorParentEmpty
	}
	newParent := Parent{
		FirstName: *req.ParentFirstName,
		LastName:  *req.ParentLastName,
		Birthday:  *req.ParentBirthday,
		Status:    "Active",
	}
	p, err := s.s.createParent(newParent)
	if err != nil {
		return Parent{}, err
	}
	return p, nil
}

func (s *service) changePatient(req UpdatePatientRequest, patientId, clinicId string) (Patient, error) {
	p, err := s.s.updatePatient(req, patientId, clinicId)
	if err != nil {
		return Patient{}, err
	}
	return p, nil
}

func (s *service) changeParent(req UpdateParentRequest, parentId, patientId, clinicId string) (Parent, error) {
	patient, err := s.getPatient(patientId, clinicId)
	if err != nil {
		return Parent{}, err
	}
	if patient.ParentId == nil || patient.ParentId.String() != parentId {
		return Parent{}, ErrorParentNotFound
	}

	p, err := s.s.updateParent(req, parentId)
	if err != nil {
		return Parent{}, err
	}
	return p, nil
}

func (s *service) removePatient(patientId, clinicId string) (string, error) {
	id, err := s.s.deletePatient(patientId, clinicId)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *service) removeParent(parentId string) (string, error) {
	id, err := s.s.deleteParent(parentId)
	if err != nil {
		return "", err
	}
	return id, nil
}

func isMinor(birthday string) (bool, error) {
	return isMinorAt(birthday, time.Now())
}

func isMinorAt(birthday string, today time.Time) (bool, error) {
	bday, err := time.ParseInLocation("2006-01-02", birthday, today.Location())
	if err != nil {
		return false, err
	}
	if bday.After(today) {
		return false, ErrorPatientBadReq
	}
	return today.Before(bday.AddDate(18, 0, 0)), nil
}
