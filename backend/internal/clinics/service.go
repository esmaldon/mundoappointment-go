package clinics

type clinicStore interface {
	fetchClinics() ([]Clinic, error)
	fetchClinic(string) (Clinic, error)
	createClinic(CreateClinicRequest) (Clinic, error)
	updateClinic(UpdateClinicRequest, string) (Clinic, error)
	deleteClinic(string) (string, error)
}

type service struct {
	s clinicStore
}

func NewService(store clinicStore) *service {
	return &service{
		s: store,
	}
}

func (s *service) getClinics() ([]Clinic, error) {
	return s.s.fetchClinics()
}

func (s *service) getClinic(clinicId string) (Clinic, error) {
	return s.s.fetchClinic(clinicId)
}

func (s *service) addClinic(req CreateClinicRequest) (Clinic, error) {
	return s.s.createClinic(req)
}

func (s *service) changeClinic(req UpdateClinicRequest, clinicId string) (Clinic, error) {
	return s.s.updateClinic(req, clinicId)
}

func (s *service) removeClinic(clinicId string) (string, error) {
	return s.s.deleteClinic(clinicId)
}
