package patients

import (
	"errors"
	"time"
	"uuid"

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

func (s *store) createPatient(req CreatePatientRequest, clinicId string) ([]Patient, error) {
	cId, err := uuid.Parse(clinicId)
	if err != nil {
		return nil, err
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
	}
	return s.db.Create[Patient](patientTableName, newPatient)
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
