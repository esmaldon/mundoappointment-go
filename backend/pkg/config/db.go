package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

var Supabase *supabase.Client

const url string = "SUPABASE_URL"
const key string = "SUPABASE_KEY"

func CreateClient() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Could not load .env file")
	}
	url := os.Getenv(url)
	key := os.Getenv(key)
	Supabase, err = supabase.NewClient(url, key, nil)
	if err != nil {
		log.Fatalf("Error during creation of client supabase %v", err.Error())
	}
	log.Println("Connection to DB succesfully")
}

func GetDBClient() *supabase.Client {
	return Supabase
}
