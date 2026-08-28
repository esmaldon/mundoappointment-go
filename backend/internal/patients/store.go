package patients

import (
	"encoding/json"
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

func createPatient(patient PatientDB) ([]Patient, error) {
	var newPatient []Patient

	result, _, err := config.GetDBClient().From("patients").Insert(patient, false, "", "", "").Execute()
	if err != nil {
		log.Printf("Could not create patient %v", patient)
		return newPatient, err
	}

	err = json.Unmarshal(result, &newPatient)
	if err != nil {
		log.Printf("Could not unmarshal json from create response")
		return newPatient, err
	}
	return newPatient, nil
}
