package patients

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type patientStoreStub struct {
	createPatientFunc func(CreatePatientRequest) ([]Patient, error)
	updatePatientFunc func(string, UpdatePatientRequest) (Patient, error)
}

func (s *patientStoreStub) fetchPatients() ([]Patient, error) {
	panic("unexpected call to fetchPatients")
}

func (s *patientStoreStub) fetchPatient(string) (Patient, error) {
	panic("unexpected call to fetchPatient")
}

func (s *patientStoreStub) createPatient(req CreatePatientRequest) ([]Patient, error) {
	if s.createPatientFunc == nil {
		panic("unexpected call to createPatient")
	}
	return s.createPatientFunc(req)
}

func (s *patientStoreStub) deletePatient(string) (string, error) {
	panic("unexpected call to deletePatient")
}

func (s *patientStoreStub) updatePatient(id string, req UpdatePatientRequest) (Patient, error) {
	if s.updatePatientFunc == nil {
		panic("unexpected call to updatePatient")
	}
	return s.updatePatientFunc(id, req)
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func TestAddPatientValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing first name",
			body: `{"lastname":"Test","birthday":"1990-01-01","phone":"5555555555","email":"patient@example.com"}`,
		},
		{
			name: "missing last name",
			body: `{"firstname":"Integration","birthday":"1990-01-01","phone":"5555555555","email":"patient@example.com"}`,
		},
		{
			name: "missing birthday",
			body: `{"firstname":"Integration","lastname":"Test","phone":"5555555555","email":"patient@example.com"}`,
		},
		{
			name: "missing phone",
			body: `{"firstname":"Integration","lastname":"Test","birthday":"1990-01-01","email":"patient@example.com"}`,
		},
		{
			name: "missing email",
			body: `{"firstname":"Integration","lastname":"Test","birthday":"1990-01-01","phone":"5555555555"}`,
		},
		{
			name: "first name too long",
			body: `{"firstname":"` + strings.Repeat("a", 101) + `","lastname":"Test","birthday":"1990-01-01","phone":"5555555555","email":"patient@example.com"}`,
		},
		{
			name: "last name too long",
			body: `{"firstname":"Integration","lastname":"` + strings.Repeat("a", 101) + `","birthday":"1990-01-01","phone":"5555555555","email":"patient@example.com"}`,
		},
		{
			name: "invalid birthday",
			body: `{"firstname":"Integration","lastname":"Test","birthday":"1990/01/01","phone":"5555555555","email":"patient@example.com"}`,
		},
		{
			name: "short phone",
			body: `{"firstname":"Integration","lastname":"Test","birthday":"1990-01-01","phone":"123","email":"patient@example.com"}`,
		},
		{
			name: "long phone",
			body: `{"firstname":"Integration","lastname":"Test","birthday":"1990-01-01","phone":"1234567890123456","email":"patient@example.com"}`,
		},
		{
			name: "invalid email",
			body: `{"firstname":"Integration","lastname":"Test","birthday":"1990-01-01","phone":"5555555555","email":"invalid-email"}`,
		},
		{
			name: "malformed json",
			body: `{"firstname":`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := performPatientRequest(
				t,
				&patientStoreStub{},
				http.MethodPost,
				"/patients",
				tt.body,
			)

			assertErrorResponse(t, recorder, http.StatusBadRequest, "BAD_REQUEST")
		})
	}
}

func TestAddPatientReturnsCreatedPatient(t *testing.T) {
	id := 42
	store := &patientStoreStub{
		createPatientFunc: func(req CreatePatientRequest) ([]Patient, error) {
			if req.Email != "patient@example.com" {
				t.Fatalf("expected request email patient@example.com, got %q", req.Email)
			}

			return []Patient{{
				Id:        &id,
				FirstName: req.FirstName,
				LastName:  req.LastName,
				Birthday:  req.Birthday,
				Phone:     req.Phone,
				Email:     req.Email,
				Status:    "Active",
			}}, nil
		},
	}

	recorder := performPatientRequest(
		t,
		store,
		http.MethodPost,
		"/patients",
		`{"firstname":"Integration","lastname":"Test","birthday":"1990-01-01","phone":"5555555555","email":"patient@example.com"}`,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}

	var response []Patient
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(response) != 1 || response[0].Id == nil || *response[0].Id != id {
		t.Fatalf("expected created patient with id %d, got %+v", id, response)
	}
}

func TestChangePatientValidation(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		expectedCode string
	}{
		{name: "empty request", body: `{}`, expectedCode: "EMPTY_REQUEST"},
		{name: "unknown field only", body: `{"unknown":"value"}`, expectedCode: "EMPTY_REQUEST"},
		{name: "empty first name", body: `{"firstname":""}`, expectedCode: "BAD_REQUEST"},
		{name: "first name too long", body: `{"firstname":"` + strings.Repeat("a", 101) + `"}`, expectedCode: "BAD_REQUEST"},
		{name: "last name too long", body: `{"lastname":"` + strings.Repeat("a", 101) + `"}`, expectedCode: "BAD_REQUEST"},
		{name: "invalid birthday", body: `{"birthday":"1990/01/01"}`, expectedCode: "BAD_REQUEST"},
		{name: "short phone", body: `{"phone":"123"}`, expectedCode: "BAD_REQUEST"},
		{name: "long phone", body: `{"phone":"1234567890123456"}`, expectedCode: "BAD_REQUEST"},
		{name: "invalid email", body: `{"email":"invalid-email"}`, expectedCode: "BAD_REQUEST"},
		{name: "invalid status", body: `{"status":"Deleted"}`, expectedCode: "BAD_REQUEST"},
		{name: "malformed json", body: `{"phone":`, expectedCode: "BAD_REQUEST"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := performPatientRequest(
				t,
				&patientStoreStub{},
				http.MethodPatch,
				"/patients/42",
				tt.body,
			)

			assertErrorResponse(t, recorder, http.StatusBadRequest, tt.expectedCode)
		})
	}
}

func TestChangePatientAcceptsPartialRequest(t *testing.T) {
	newPhone := "9999999999"
	store := &patientStoreStub{
		updatePatientFunc: func(id string, req UpdatePatientRequest) (Patient, error) {
			if id != "42" {
				t.Fatalf("expected patient id 42, got %q", id)
			}
			if req.Phone == nil || *req.Phone != newPhone {
				t.Fatalf("expected phone %q, got %+v", newPhone, req.Phone)
			}
			if req.FirstName != nil || req.LastName != nil || req.Birthday != nil || req.Email != nil || req.Status != nil {
				t.Fatalf("expected only phone in partial update, got %+v", req)
			}

			return Patient{Id: intPointer(42), Phone: newPhone}, nil
		},
	}

	recorder := performPatientRequest(
		t,
		store,
		http.MethodPatch,
		"/patients/42",
		`{"phone":"9999999999"}`,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	var response Patient
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if response.Id == nil || *response.Id != 42 || response.Phone != newPhone {
		t.Fatalf("unexpected update response: %+v", response)
	}
}

func performPatientRequest(
	t *testing.T,
	store patientStore,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	h := NewHandler(store)
	router := gin.New()
	router.POST("/patients", h.addPatient)
	router.PATCH("/patients/:id", h.changePatient)

	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	return recorder
}

func assertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, status int, code string) {
	t.Helper()

	if recorder.Code != status {
		t.Fatalf("expected status %d, got %d; body=%s", status, recorder.Code, recorder.Body.String())
	}

	var response errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decoding error response: %v", err)
	}
	if response.Code != code {
		t.Errorf("expected error code %q, got %q", code, response.Code)
	}
	if response.Message == "" {
		t.Error("expected a non-empty error message")
	}
}

func intPointer(value int) *int {
	return &value
}
