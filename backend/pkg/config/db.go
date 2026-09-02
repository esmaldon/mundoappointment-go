package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

type DBClient struct {
	Supabase *supabase.Client
}

const url string = "SUPABASE_URL"
const key string = "SUPABASE_KEY"

func NewDBClient() (*DBClient, error) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Could not load .env file")
	}
	url := os.Getenv(url)
	key := os.Getenv(key)
	client, err := supabase.NewClient(url, key, nil)
	if err != nil {
		return nil, fmt.Errorf("error during creation of client supabase %v", err.Error())
	}
	return &DBClient{
		Supabase: client,
	}, nil
}

func (d *DBClient) FetchAll[T any](table string) ([]T, error) {
	var result []T
	data, _, err := d.Supabase.From(table).Select("*", "exact", false).Execute()
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (d *DBClient) FetchById[T any](table, id string) (T, error) {
	var result []T
	var zero T
	data, count, err := d.Supabase.From(table).Select("*", "exact", false).Filter("id", "eq", id).Execute()
	if err != nil {
		return zero, err
	}
	if count == 0 {
		return zero, fmt.Errorf("%w: id=%s", ErrorRecordNotFound, id)
	}
	if count > 1 {
		return zero, fmt.Errorf("warning, search returned more than one record. Total of records %v", count)
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return zero, err
	}
	if len(result) == 0 {
		return zero, fmt.Errorf("%w: id=%s", ErrorRecordNotFound, id)
	}

	return result[0], nil
}

func (d *DBClient) Create[T any](table string, obj T) ([]T, error) {
	var result []T
	data, count, err := d.Supabase.From(table).Insert(obj, false, "", "representation", "exact").Execute()
	if err != nil {
		return result, err
	}
	if count == 0 {
		return result, fmt.Errorf("warning, %v records saved in db", count)
	}
	if count > 1 {
		return result, fmt.Errorf("warning, more than one record was saved in db. Total count %v", count)
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (d *DBClient) Delete(table, id string) (string, error) {
	var zero string
	_, count, err := d.Supabase.From(table).Delete("", "exact").Filter("id", "eq", id).Execute()
	if err != nil {
		return zero, err
	}
	if count == 0 {
		return zero, fmt.Errorf("%w: id=%s", ErrorRecordNotFound, id)
	}
	if count > 1 {
		return zero, fmt.Errorf("warning, more than one record was deleted in db. Total count %v", count)
	}

	return id, nil
}

func (d *DBClient) Update[T any](table, id string, obj T) (T, error) {
	var result []T
	var zero T
	data, count, err := d.Supabase.From(table).Update(obj, "", "exact").Filter("id", "eq", id).Execute()
	if err != nil {
		return zero, err
	}
	if count == 0 {
		return zero, fmt.Errorf("%w: id=%s", ErrorRecordNotFound, id)
	}
	if count > 1 {
		return zero, fmt.Errorf("warning, more than one record was deleted in db. Total count %v", count)
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return zero, err
	}

	return result[0], nil
}
