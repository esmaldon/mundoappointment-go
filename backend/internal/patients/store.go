package patients

import (
	"errors"
	"uuid"

	"mundoappointment.com/pkg/config"
)

const patientTableName = "patients"
const parentTableName = "parents"

var ErrorPatientNotFound = errors.New("patient not found")
var ErrorParentNotFound = errors.New("parent not found")
var ErrorParentEmpty = errors.New("parent is empty")

type store struct {
	db *config.DBClient
}

func NewStore(db *config.DBClient) *store {
	return &store{
		db: db,
	}
}

func (s *store) fetchPatients(clinicId string) ([]Patient, error) {
	return s.db.FetchAllByClinic[Patient](patientTableName, clinicId)
}

func (s *store) fetchPatient(patientId, clinicId string) (Patient, error) {
	patient, err := s.db.FetchByIdAndClinicId[Patient](patientTableName, patientId, clinicId)
	if err != nil {
		if errors.Is(err, config.ErrorRecordNotFound) {
			return Patient{}, ErrorPatientNotFound
		}
		return Patient{}, err
	}

	return patient, nil
}

func (s *store) createPatient(p Patient, clinicId uuid.UUID) (Patient, error) {
	return s.db.Create[Patient](patientTableName, p)
}

// Keep the public API independent of the parents table's column names.
type parentRecord struct {
	Id        *uuid.UUID `json:"id,omitempty"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Birthday  string     `json:"birthday"`
	Status    string     `json:"status"`
}

func (p parentRecord) parent() Parent {
	return Parent{Id: p.Id, FirstName: p.FirstName, LastName: p.LastName, Birthday: p.Birthday, Status: p.Status}
}
func (s *store) createParent(p Parent) (Parent, error) {
	row, err := s.db.Create[parentRecord](parentTableName, parentRecord{Id: p.Id, FirstName: p.FirstName, LastName: p.LastName, Birthday: p.Birthday, Status: p.Status})
	return row.parent(), err
}

func (s *store) updatePatient(req UpdatePatientRequest, patientId, clinicId string) (Patient, error) {
	patientUpdated, err := s.db.UpdateByIdAndClinicId[Patient](patientTableName, patientId, clinicId, req)
	if err != nil {
		if errors.Is(err, config.ErrorRecordNotFound) {
			return Patient{}, ErrorPatientNotFound
		}
		return Patient{}, err
	}
	return patientUpdated, nil
}

func (s *store) updateParent(req UpdateParentRequest, parentId string) (Parent, error) {
	// Translate API fields to the parents table column names.
	patch := struct {
		FirstName *string `json:"first_name,omitempty"`
		LastName  *string `json:"last_name,omitempty"`
		Birthday  *string `json:"birthday,omitempty"`
		Status    *string `json:"status,omitempty"`
	}{req.ParentFirstName, req.ParentLastName, req.ParentBirthday, req.ParentStatus}
	parentUpdated, err := s.db.UpdateById[parentRecord](parentTableName, parentId, patch)
	if err != nil {
		if errors.Is(err, config.ErrorRecordNotFound) {
			return Parent{}, ErrorParentNotFound
		}
		return Parent{}, err
	}
	return parentUpdated.parent(), nil
}

func (s *store) deletePatient(patientId, clinicId string) (string, error) {
	patient, err := s.db.DeleteByClinicId(patientTableName, patientId, clinicId)
	if err != nil {
		if errors.Is(err, config.ErrorRecordNotFound) {
			return patientId, ErrorPatientNotFound
		}
		return patientId, err
	}
	return patient, nil
}

func (s *store) deleteParent(parentId string) (string, error) {
	parent, err := s.db.Delete(parentTableName, parentId)
	if err != nil {
		if errors.Is(err, config.ErrorRecordNotFound) {
			return parentId, ErrorParentNotFound
		}
		return parentId, err
	}
	return parent, nil
}
