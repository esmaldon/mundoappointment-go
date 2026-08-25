package patients

import (
	"log"

	"mundoappointment.com/pkg/config"
)

func fetchPatients() ([]Patient, error) {
	var patients []Patient
	_, err := config.GetDBClient().From("patients").Select("*", "exact", false).ExecuteTo(&patients)
	if err != nil {
		log.Printf("Could not fetch patients data %v", err.Error())
		return nil, err
	}
	return patients, nil
}
