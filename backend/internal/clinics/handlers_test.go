package clinics

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"uuid"

	"github.com/gin-gonic/gin"
)

const testClinicID = "ed1618d9-cc28-463b-aa93-b2a9d583459b"

// Missing callbacks deliberately fail: rejected requests must not reach the store.
type clinicStoreStub struct {
	list   func() ([]Clinic, error)
	get    func(string) (Clinic, error)
	create func(CreateClinic) ([]Clinic, error)
	update func(UpdateClinic, string) (Clinic, error)
	delete func(string) (string, error)
}

func (s clinicStoreStub) fetchClinics() ([]Clinic, error)                 { return s.list() }
func (s clinicStoreStub) fetchClinic(id string) (Clinic, error)           { return s.get(id) }
func (s clinicStoreStub) createClinic(req CreateClinic) ([]Clinic, error) { return s.create(req) }
func (s clinicStoreStub) updateClinic(req UpdateClinic, id string) (Clinic, error) {
	return s.update(req, id)
}
func (s clinicStoreStub) deleteClinic(id string) (string, error) { return s.delete(id) }

func clinicRequest(s clinicStore, method, path, body string) *httptest.ResponseRecorder {
	router := gin.New()
	registerClinicRoutes(router.Group("/api/v1"), s)
	req := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func assertClinicJSON(t *testing.T, w *httptest.ResponseRecorder, status int, want any) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("status: got %d, want %d; body=%s", w.Code, status, w.Body.String())
	}
	if !strings.HasPrefix(w.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("expected JSON content type, got %q", w.Header().Get("Content-Type"))
	}
	wantBytes, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var gotJSON, wantJSON any
	if err := json.Unmarshal(w.Body.Bytes(), &gotJSON); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(wantBytes, &wantJSON); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotJSON, wantJSON) {
		t.Fatalf("response: got %s, want %s", w.Body.String(), wantBytes)
	}
}

func TestClinicRequestValidation(t *testing.T) {
	cases := []struct{ name, method, body, code, message string }{
		{"create empty", "POST", `{}`, "BAD_REQUEST", "required"},
		{"create missing name", "POST", `{"timezone":"UTC"}`, "BAD_REQUEST", "Name"},
		{"create missing timezone", "POST", `{"name":"Test"}`, "BAD_REQUEST", "Timezone"},
		{"create empty name", "POST", `{"name":"","timezone":"UTC"}`, "BAD_REQUEST", "required"},
		{"create long name", "POST", `{"name":"` + strings.Repeat("a", 51) + `","timezone":"UTC"}`, "BAD_REQUEST", "max"},
		{"create long timezone", "POST", `{"name":"Test","timezone":"` + strings.Repeat("a", 51) + `"}`, "BAD_REQUEST", "max"},
		{"create wrong type", "POST", `{"name":42,"timezone":"UTC"}`, "BAD_REQUEST", ""},
		{"create malformed", "POST", `{"name":`, "BAD_REQUEST", ""},
		{"create missing body", "POST", ``, "BAD_REQUEST", ""},
		{"patch empty", "PATCH", `{}`, "EMPTY_REQUEST", "at least one field"},
		{"patch unknown only", "PATCH", `{"id":"ignored"}`, "EMPTY_REQUEST", "at least one field"},
		{"patch null only", "PATCH", `{"name":null}`, "EMPTY_REQUEST", "at least one field"},
		{"patch empty name", "PATCH", `{"name":""}`, "BAD_REQUEST", "min"},
		{"patch empty timezone", "PATCH", `{"timezone":""}`, "BAD_REQUEST", "min"},
		{"patch long name", "PATCH", `{"name":"` + strings.Repeat("a", 51) + `"}`, "BAD_REQUEST", "max"},
		{"patch long timezone", "PATCH", `{"timezone":"` + strings.Repeat("a", 51) + `"}`, "BAD_REQUEST", "max"},
		{"patch bad status", "PATCH", `{"status":"Deleted"}`, "BAD_REQUEST", "oneof"},
		{"patch empty status", "PATCH", `{"status":""}`, "BAD_REQUEST", "oneof"},
		{"patch malformed", "PATCH", `{"status":`, "BAD_REQUEST", ""},
		{"patch wrong type", "PATCH", `{"name":42}`, "BAD_REQUEST", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := "/clinics/"
			if tc.method == "PATCH" {
				path = "/clinic/" + testClinicID
			}
			w := clinicRequest(clinicStoreStub{}, tc.method, path, tc.body)
			var response struct{ Code, Message string }
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Message == "" || !strings.Contains(response.Message, tc.message) {
				t.Fatalf("unexpected validation message %q", response.Message)
			}
			assertClinicJSON(t, w, http.StatusBadRequest, map[string]string{"code": tc.code, "message": response.Message})
		})
	}
}

func TestClinicSuccessResponses(t *testing.T) {
	id, err := uuid.Parse(testClinicID)
	if err != nil {
		t.Fatal(err)
	}
	clinic := Clinic{Id: &id, Name: "Test clinic", Status: "Active", Timezone: "UTC"}
	var receivedID string
	var receivedCreate CreateClinic
	checkID := func(got string) { receivedID = got }
	cases := []struct {
		name, method, path, body string
		status                   int
		want                     any
		store                    clinicStoreStub
	}{
		{"list", "GET", "/clinics", "", 200, []Clinic{clinic}, clinicStoreStub{list: func() ([]Clinic, error) { return []Clinic{clinic}, nil }}},
		{"empty list", "GET", "/clinics", "", 200, []Clinic{}, clinicStoreStub{list: func() ([]Clinic, error) { return []Clinic{}, nil }}},
		{"get", "GET", "/clinic/" + testClinicID, "", 200, clinic, clinicStoreStub{get: func(id string) (Clinic, error) { checkID(id); return clinic, nil }}},
		{"create", "POST", "/clinics/", `{"name":"Test clinic","timezone":"UTC"}`, 201, []Clinic{clinic}, clinicStoreStub{create: func(req CreateClinic) ([]Clinic, error) {
			receivedCreate = req
			return []Clinic{clinic}, nil
		}}},
		{"delete", "DELETE", "/clinic/" + testClinicID, "", 200, testClinicID, clinicStoreStub{delete: func(id string) (string, error) { checkID(id); return id, nil }}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			receivedID = ""
			receivedCreate = CreateClinic{}
			assertClinicJSON(t, clinicRequest(tc.store, tc.method, tc.path, tc.body), tc.status, tc.want)
			if (tc.name == "get" || tc.name == "delete") && receivedID != testClinicID {
				t.Fatalf("clinic ID: got %q", receivedID)
			}
			if tc.name == "create" && receivedCreate != (CreateClinic{Name: "Test clinic", Timezone: "UTC"}) {
				t.Fatalf("unexpected create request: %+v", receivedCreate)
			}
		})
	}
}

func TestClinicPartialUpdates(t *testing.T) {
	for _, body := range []string{`{"name":"Updated"}`, `{"timezone":"America/Mexico_City"}`, `{"status":"Inactive"}`, `{"status":"Active"}`} {
		t.Run(body, func(t *testing.T) {
			calls := 0
			result := Clinic{Name: "Updated", Timezone: "UTC", Status: "Active"}
			s := clinicStoreStub{update: func(req UpdateClinic, id string) (Clinic, error) {
				calls++
				if id != testClinicID {
					t.Fatalf("clinic ID: got %q", id)
				}
				data, err := json.Marshal(req)
				if err != nil {
					t.Fatal(err)
				}
				if string(data) != body {
					t.Fatalf("omitted fields must stay omitted: got %s, want %s", data, body)
				}
				return result, nil
			}}
			assertClinicJSON(t, clinicRequest(s, "PATCH", "/clinic/"+testClinicID, body), 200, result)
			if calls != 1 {
				t.Fatalf("expected one store call, got %d", calls)
			}
		})
	}
}

func TestClinicStoreErrors(t *testing.T) {
	for _, operation := range []string{"list", "get", "create", "update", "delete"} {
		for _, notFound := range []bool{false, true} {
			if notFound && (operation == "list" || operation == "create") {
				continue
			}
			t.Run(fmt.Sprintf("%s/not_found=%t", operation, notFound), func(t *testing.T) {
				storeErr := errors.New("database unavailable")
				status, code := 500, "INTERNAL_ERROR"
				if notFound {
					storeErr = fmt.Errorf("lookup: %w", ErrorClinicNotFound)
					status, code = 404, "NOT_FOUND"
				}
				calls := 0
				s := clinicStoreStub{
					list:   func() ([]Clinic, error) { calls++; return nil, storeErr },
					get:    func(string) (Clinic, error) { calls++; return Clinic{}, storeErr },
					create: func(CreateClinic) ([]Clinic, error) { calls++; return nil, storeErr },
					update: func(UpdateClinic, string) (Clinic, error) { calls++; return Clinic{}, storeErr },
					delete: func(string) (string, error) { calls++; return "", storeErr },
				}
				method, path, body := "GET", "/clinic/"+testClinicID, ""
				switch operation {
				case "list":
					path = "/clinics"
				case "create":
					method, path, body = "POST", "/clinics/", `{"name":"Test","timezone":"UTC"}`
				case "update":
					method, body = "PATCH", `{"status":"Inactive"}`
				case "delete":
					method = "DELETE"
				}
				assertClinicJSON(t, clinicRequest(s, method, path, body), status, map[string]string{"code": code, "message": storeErr.Error()})
				if calls != 1 {
					t.Fatalf("expected one store call, got %d", calls)
				}
			})
		}
	}
}
