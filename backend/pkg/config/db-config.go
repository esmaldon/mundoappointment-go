package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

var Supabase *supabase.Client

func CreateClient() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalln("Could not load .env file")
	}
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")
	Supabase, err = supabase.NewClient(url, key, nil)
	if err != nil {
		log.Fatalf("Error during creation of client supabase %v", err.Error())
	}
	log.Println("Connection to DB succesfully")
}
