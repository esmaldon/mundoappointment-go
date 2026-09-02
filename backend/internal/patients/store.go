package patients

import (
	"errors"

	"mundoappointment.com/pkg/config"
)

const patientTableName = "patients"

var ErrorPatientNotFound = errors.New("patient not found")

type store struct {
	db *config.DBClient
}

func NewStore(db *config.DBClient) *store {
	return &store{
		db: db,
	}
}

func (s *store) fetchPatients() ([]Patient, error) {
	return s.db.FetchAll[Patient](patientTableName)
}

func (s *store) fetchPatient(patientId string) (Patient, error) {
	patient, err := s.db.FetchById[Patient](patientTableName, patientId)
	if err != nil {
		if errors.Is(err, config.ErrorRecordNotFound) {
			return Patient{}, ErrorPatientNotFound
		}
		return Patient{}, err
	}

	return patient, nil
}

func (s *store) createPatient(patient Patient) ([]Patient, error) {
	return s.db.Create(patientTableName, patient)
}

func (s *store) deletePatient(patientId string) (string, error) {
	patient, err := s.db.Delete(patientTableName, patientId)
	if err != nil {
		if errors.Is(err, config.ErrorRecordNotFound) {
			return patientId, ErrorPatientNotFound
		}
		return patientId, err
	}
	return patient, nil
}

func (s *store) updatePatient(patientId string, patient Patient) (Patient, error) {
	patientUpdated, err := s.db.Update(patientTableName, patientId, patient)
	if err != nil {
		if errors.Is(err, config.ErrorRecordNotFound) {
			return Patient{}, ErrorPatientNotFound
		}
		return Patient{}, err
	}
	return patientUpdated, nil
}
