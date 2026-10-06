package patients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"github.com/gin-gonic/gin"
)

type patientServiceStub struct {
	parentFunc        func(UpdateParentRequest, string, string, string) (Parent, error)
	createPatientFunc func(CreatePatientRequest) (Patient, error)
	updatePatientFunc func(string, UpdatePatientRequest) (Patient, error)
}

func (s *patientServiceStub) getPatients(clinicId string) ([]Patient, error) {
	panic("unexpected call to fetchPatients")
}

func (s *patientServiceStub) getPatient(id, clinicId string) (Patient, error) {
	panic("unexpected call to fetchPatient")
}

func (s *patientServiceStub) addPatient(req CreatePatientRequest, clinicId string) (Patient, error) {
	if s.createPatientFunc == nil {
		panic("unexpected call to createPatient")
	}
	if clinicId != testClinicID {
		panic("incorrect clinic ID")
	}
	return s.createPatientFunc(req)
}

func (s *patientServiceStub) removePatient(id, clinicId string) (string, error) {
	panic("unexpected call to deletePatient")
}

func (s *patientServiceStub) changePatient(req UpdatePatientRequest, id, clinicId string) (Patient, error) {
	if s.updatePatientFunc == nil {
		panic("unexpected call to updatePatient")
	}
	if clinicId != testClinicID {
		panic("incorrect clinic ID")
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
				&patientServiceStub{},
				http.MethodPost,
				"/patients",
				tt.body,
			)

			assertErrorResponse(t, recorder, http.StatusBadRequest, "BAD_REQUEST")
		})
	}
}

func TestAddPatientReturnsCreatedPatient(t *testing.T) {
	id := uuid.MustParse(testPatientID)
	store := &patientServiceStub{
		createPatientFunc: func(req CreatePatientRequest) (Patient, error) {
			if req.Email != "patient@example.com" {
				t.Fatalf("expected request email patient@example.com, got %q", req.Email)
			}

			return Patient{
				Id:        &id,
				FirstName: req.FirstName,
				LastName:  req.LastName,
				Birthday:  req.Birthday,
				Phone:     req.Phone,
				Email:     req.Email,
				Status:    "Active",
			}, nil
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

	var response Patient
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if response.Id == nil || *response.Id != id {
		t.Fatalf("expected created patient with id %s, got %+v", id, response)
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
				&patientServiceStub{},
				http.MethodPatch,
				"/patients/"+testPatientID,
				tt.body,
			)

			assertErrorResponse(t, recorder, http.StatusBadRequest, tt.expectedCode)
		})
	}
}

func TestChangePatientAcceptsPartialRequest(t *testing.T) {
	newPhone := "9999999999"
	store := &patientServiceStub{
		updatePatientFunc: func(id string, req UpdatePatientRequest) (Patient, error) {
			if id != testPatientID {
				t.Fatalf("expected patient id %s, got %q", testPatientID, id)
			}
			if req.Phone == nil || *req.Phone != newPhone {
				t.Fatalf("expected phone %q, got %+v", newPhone, req.Phone)
			}
			if req.FirstName != nil || req.LastName != nil || req.Birthday != nil || req.Email != nil || req.Status != nil {
				t.Fatalf("expected only phone in partial update, got %+v", req)
			}

			return Patient{Id: uuidPointer(testPatientID), Phone: newPhone}, nil
		},
	}

	recorder := performPatientRequest(
		t,
		store,
		http.MethodPatch,
		"/patients/"+testPatientID,
		`{"phone":"9999999999"}`,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	var response Patient
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if response.Id == nil || response.Id.String() != testPatientID || response.Phone != newPhone {
		t.Fatalf("unexpected update response: %+v", response)
	}
}

func performPatientRequest(
	t *testing.T,
	store patientService,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	registerPatientsRoutes(router.Group("/api/v1"), store)

	request := httptest.NewRequest(method, "/api/v1/clinic/"+testClinicID+path, bytes.NewBufferString(body))
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

func uuidPointer(value string) *uuid.UUID {
	id := uuid.MustParse(value)
	return &id
}

const testPatientID = "ed1618d9-cc28-463b-aa93-b2a9d583459a"
const testClinicID = "ed1618d9-cc28-463b-aa93-b2a9d583459b"

func (s *patientServiceStub) addParent(CreatePatientRequest) (Parent, error) {
	panic("unexpected addParent")
}
func (s *patientServiceStub) changeParent(req UpdateParentRequest, parent, patient, clinic string) (Parent, error) {
	return s.parentFunc(req, parent, patient, clinic)
}
func (s *patientServiceStub) removeParent(string) (string, error) { panic("unexpected removeParent") }

func TestAddPatientServiceErrors(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{ErrorPatientBadReq, 400, "BAD_REQUEST"}, {ErrorParentEmpty, 400, "BAD_REQUEST"}, {errors.New("database unavailable"), 500, "INTERNAL_ERROR"},
	} {
		t.Run(tc.code+tc.err.Error(), func(t *testing.T) {
			stub := &patientServiceStub{createPatientFunc: func(CreatePatientRequest) (Patient, error) { return Patient{}, fmt.Errorf("create: %w", tc.err) }}
			w := performPatientRequest(t, stub, "POST", "/patients", `{"firstname":"Test","lastname":"Test","birthday":"1990-01-01","phone":"5555555555","email":"test@example.com"}`)
			assertErrorResponse(t, w, tc.status, tc.code)
		})
	}
}
func TestParentHandler(t *testing.T) {
	parentID := uuid.New()
	path := "/patients/" + testPatientID + "/parent/" + parentID.String()
	for _, tc := range []struct{ body, code string }{
		{`{}`, "EMPTY_REQUEST"}, {`{"parent_firstname":null}`, "EMPTY_REQUEST"}, {`{"parent_firstname":""}`, "BAD_REQUEST"}, {`{"parent_status":"Deleted"}`, "BAD_REQUEST"}, {`{"parent_birthday":"invalid"}`, "BAD_REQUEST"}, {`{"parent_lastname":`, "BAD_REQUEST"},
	} {
		w := performPatientRequest(t, &patientServiceStub{}, "PATCH", path, tc.body)
		assertErrorResponse(t, w, 400, tc.code)
	}
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{nil, 200, ""}, {ErrorParentNotFound, 404, "NOT_FOUND"}, {ErrorPatientNotFound, 404, "NOT_FOUND"}, {errors.New("database unavailable"), 500, "INTERNAL_ERROR"},
	} {
		called := false
		stub := &patientServiceStub{parentFunc: func(req UpdateParentRequest, parent, patient, clinic string) (Parent, error) {
			called = true
			if parent != parentID.String() || patient != testPatientID || clinic != testClinicID || req.ParentFirstName == nil || *req.ParentFirstName != "Updated" {
				t.Fatal("incorrect parent update or scope")
			}
			if tc.err != nil {
				return Parent{}, fmt.Errorf("update: %w", tc.err)
			}
			return Parent{Id: &parentID, FirstName: *req.ParentFirstName}, nil
		}}
		w := performPatientRequest(t, stub, "PATCH", path, `{"parent_firstname":"Updated"}`)
		if !called {
			t.Fatal("handler was not invoked")
		}
		if tc.err != nil {
			assertErrorResponse(t, w, tc.status, tc.code)
		} else {
			var got Parent
			if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &got) != nil || got.Id == nil || *got.Id != parentID || got.FirstName != "Updated" {
				t.Fatalf("unexpected response %s", w.Body.String())
			}
		}
	}
}
