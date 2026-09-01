package patients

import (
	"encoding/json"
	"log"
	"strconv"

	"mundoappointment.com/pkg/config"
)

const PATIENT_TABLE_NAME = "patients"

func fetchPatients() ([]Patient, error) {
	var patients []Patient
	_, err := config.GetDBClient().From(PATIENT_TABLE_NAME).Select("*", "exact", false).ExecuteTo(&patients)
	if err != nil {
		return nil, err
	}
	return patients, nil
}

func fetchPatient(patientId string) (Patient, error) {
	var fetchedPatient []Patient
	result, count, err := config.GetDBClient().From(PATIENT_TABLE_NAME).Select("*", "exact", false).Filter("id", "eq", patientId).Execute()
	if err != nil {
		return Patient{}, err
	}
	if count > 1 {
		log.Printf("Result of fetch patient returned more than one record. Total of records %v", count)
		return Patient{}, nil
	}
	if count == 0 {
		log.Printf("No patient found")
		return Patient{}, nil
	}
	err = json.Unmarshal(result, &fetchedPatient)
	if err != nil {
		return Patient{}, err
	}
	return fetchedPatient[0], nil
}

func createPatient(patient PatientDB) ([]Patient, error) {
	var newPatient []Patient

	result, _, err := config.GetDBClient().From(PATIENT_TABLE_NAME).Insert(patient, false, "", "", "").Execute()
	if err != nil {
		return newPatient, err
	}

	err = json.Unmarshal(result, &newPatient)
	if err != nil {
		return newPatient, err
	}
	return newPatient, nil
}

func deletePatient(patientId string) (int, error) {
	var patient Patient
	result, count, err := config.GetDBClient().From(PATIENT_TABLE_NAME).Delete("", "exact").Filter("id", "eq", patientId).Execute()
	if err != nil {
		return 0, err
	}
	if count == 0 {
		// no rows affected then not found
		return 0, nil
	}
	err = json.Unmarshal(result, &patient)
	if err != nil {
		return 0, err
	}
	return patient.Id, nil
}

func updatePatient(patient Patient) (Patient, error) {
	var updatedPatient Patient
	result, count, err := config.GetDBClient().From(PATIENT_TABLE_NAME).Update(patient, "", "exact").Filter("id", "eq", strconv.Itoa(patient.Id)).Execute()
	if err != nil {
		return updatedPatient, err
	}
	if count == 0 {
		// no rows affected then not found
		return patient, nil
	}
	if count > 1 {
		log.Printf("Rows affected during update patient %v", count)
	}
	err = json.Unmarshal(result, &updatedPatient)
	if err != nil {
		return updatedPatient, err
	}
	return patient, nil
}
