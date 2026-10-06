package config

import (
	"github.com/supabase-community/supabase-go"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCreateRequiresOneReturnedRecord(t *testing.T) {
	// No network or credentials: the Supabase transport delegates to this fake.
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	for _, tc := range []struct {
		name, body, count string
		fail              bool
	}{
		{"one", `[{"name":"Created"}]`, "0-0/1", false},
		{"empty with count one", `[]`, "0-0/1", true},
		{"null with count one", `null`, "0-0/1", true},
		{"multiple with count one", `[{},{}]`, "0-0/1", true},
		{"zero count", `[]`, "*/0", true},
		{"multiple count", `[{},{}]`, "0-1/2", true},
		{"malformed", `{`, "0-0/1", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method != "POST" || r.URL.Path != "/rest/v1/clinics" {
					t.Fatalf("unexpected request %s %s", r.Method, r.URL)
				}
				return &http.Response{StatusCode: 201, Header: http.Header{"Content-Range": []string{tc.count}}, Body: io.NopCloser(strings.NewReader(tc.body)), Request: r}, nil
			})
			client, err := supabase.NewClient("https://test.invalid", "test-key", nil)
			if err != nil {
				t.Fatal(err)
			}
			db := &DBClient{Supabase: client}
			type record struct {
				Name string `json:"name"`
			}
			got, err := db.Create[record]("clinics", record{Name: "Created"})
			if (err != nil) != tc.fail {
				t.Fatalf("result=%+v error=%v", got, err)
			}
			if !tc.fail && got.Name != "Created" {
				t.Fatal(got)
			}
		})
	}
}
