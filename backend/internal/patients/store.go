package patients

import (
	"log"

	"mundoappointment.com/pkg/config"
)

func fetchPatients() []Patient {
	var patients []Patient
	_, err := config.Supabase.From("patients").Select("*", "exact", false).ExecuteTo(&patients)
	if err != nil {
		log.Fatalf("Could not fetch patients data %v", err.Error())
	}
	log.Printf("result %v", patients)
	return patients
}
