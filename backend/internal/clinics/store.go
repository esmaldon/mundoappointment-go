package clinics

import (
	"errors"

	"mundoappointment.com/pkg/config"
)

const clinicTableName = "clinics"

var ErrorClinicNotFound = errors.New("clinic not found")

type store struct {
	db *config.DBClient
}

func NewStore(db *config.DBClient) *store {
	return &store{
		db: db,
	}
}

func (s *store) fetchClinics() ([]Clinic, error) {
	return s.db.FetchAll[Clinic](clinicTableName)
}

func (s *store) fetchClinic(clinicId string) (Clinic, error) {
	clinic, err := s.db.FetchById[Clinic](clinicTableName, clinicId)
	if err != nil {
		if errors.Is(err, config.ErrorRecordNotFound) {
			return Clinic{}, ErrorClinicNotFound
		}
		return Clinic{}, err
	}

	return clinic, nil
}

func (s *store) createClinic(req CreateClinicRequest) (Clinic, error) {
	newClinic := Clinic{
		Name:     req.Name,
		Status:   "Active",
		Timezone: req.Timezone,
	}

	return s.db.Create[Clinic](clinicTableName, newClinic)
}

func (s *store) updateClinic(req UpdateClinicRequest, clinicId string) (Clinic, error) {
	clinicUpdated, err := s.db.UpdateById[Clinic](clinicTableName, clinicId, req)
	if err != nil {
		if errors.Is(err, config.ErrorRecordNotFound) {
			return Clinic{}, ErrorClinicNotFound
		}
		return Clinic{}, err
	}
	return clinicUpdated, nil
}

func (s *store) deleteClinic(clinicId string) (string, error) {
	patient, err := s.db.Delete(clinicTableName, clinicId)
	if err != nil {
		if errors.Is(err, config.ErrorRecordNotFound) {
			return clinicId, ErrorClinicNotFound
		}
		return clinicId, err
	}
	return patient, nil
}
