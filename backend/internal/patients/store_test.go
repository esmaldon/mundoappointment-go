package patients

import (
	"encoding/json"
	"github.com/supabase-community/supabase-go"
	"io"
	"mundoappointment.com/pkg/config"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"uuid"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestParentPatchUsesDatabaseColumnNames(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	id := uuid.New()
	called := false
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		called = true
		if r.Method != "PATCH" || r.URL.Path != "/rest/v1/parents" || r.URL.Query().Get("id") != "eq."+id.String() {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		want := map[string]any{"first_name": "Updated", "birthday": "1980-02-03", "status": "Inactive"}
		if !reflect.DeepEqual(payload, want) {
			t.Fatalf("payload=%v want=%v", payload, want)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Range": []string{"0-0/1"}}, Body: io.NopCloser(strings.NewReader(`[{"id":"` + id.String() + `","first_name":"Updated"}]`)), Request: r}, nil
	})
	client, err := supabase.NewClient("https://test.invalid", "test-key", nil)
	if err != nil {
		t.Fatal(err)
	}
	s := NewStore(&config.DBClient{Supabase: client})
	p, err := s.updateParent(UpdateParentRequest{ParentFirstName: stringPointer("Updated"), ParentBirthday: stringPointer("1980-02-03"), ParentStatus: stringPointer("Inactive")}, id.String())
	if err != nil || !called || p.Id == nil || *p.Id != id || p.FirstName != "Updated" {
		t.Fatalf("result=%+v error=%v", p, err)
	}
}
func TestAdultParentIDSerializesAsNull(t *testing.T) {
	data, err := json.Marshal(Patient{})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	v, exists := fields["parent_id"]
	if !exists || v != nil {
		t.Fatalf("parent_id must be null: %s", data)
	}
}

func TestParentCreateUsesDatabaseColumnNames(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	id := uuid.New()
	called := false
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		called = true
		if r.Method != "POST" || r.URL.Path != "/rest/v1/parents" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		want := map[string]any{"first_name": "Parent", "last_name": "Test", "birthday": "1980-02-03", "status": "Active"}
		if !reflect.DeepEqual(payload, want) {
			t.Fatalf("payload=%v want=%v", payload, want)
		}
		return &http.Response{StatusCode: 201, Header: http.Header{"Content-Range": []string{"0-0/1"}}, Body: io.NopCloser(strings.NewReader(`[{"id":"` + id.String() + `","first_name":"Parent","last_name":"Test","birthday":"1980-02-03","status":"Active"}]`)), Request: r}, nil
	})
	client, err := supabase.NewClient("https://test.invalid", "test-key", nil)
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewStore(&config.DBClient{Supabase: client}).createParent(Parent{FirstName: "Parent", LastName: "Test", Birthday: "1980-02-03", Status: "Active"})
	if err != nil || !called || p.Id == nil || *p.Id != id || p.FirstName != "Parent" || p.LastName != "Test" {
		t.Fatalf("result=%+v error=%v", p, err)
	}
}
