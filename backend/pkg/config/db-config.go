package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

var Supabase *supabase.Client

func CreateClient() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Could not load .env file")
	}
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")
	Supabase, err = supabase.NewClient(url, key, nil)
	if err != nil {
		fmt.Printf("Error during creation of client supabase")
	}
}
