package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_api(t *testing.T) {
	tests := []struct {
		name      string
		store     string
		shortener shortener
	}{
		{
			name:      "mem with fake shortener",
			store:     "mem",
			shortener: strings.ToUpper, // use ToUpper for testing purpose
		},
		{
			name:      "mem with base32 shortener",
			store:     "mem",
			shortener: sha256ShortenerB62,
		},
		{
			name:      "postgres with fake shortener",
			store:     "postgres",
			shortener: strings.ToUpper, // use ToUpper for testing purpose
		},
		{
			name:      "postgres with base32 shortener",
			store:     "postgres",
			shortener: sha256ShortenerB62,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := api{
				store:     storeFactory(t, tt.store),
				shortener: tt.shortener,
			}
			server := httptest.NewServer(api.Routes())
			t.Cleanup(func() { server.Close() })

			t.Run("register", func(t *testing.T) {
				long := "foo"
				short := register(t, server.URL, long)
				if got, want := short, tt.shortener(long); got != want {
					t.Errorf("got %v want %v", got, want)
				}
			})

			t.Run("register two time", func(t *testing.T) {
				long := "foo"
				short := register(t, server.URL, long)
				if got, want := short, tt.shortener(long); got != want {
					t.Errorf("got %v want %v", got, want)
				}
				short2 := register(t, server.URL, long)
				if got, want := short2, tt.shortener(long); got != want {
					t.Errorf("got %v want %v", got, want)
				}
			})

			t.Run("redirect", func(t *testing.T) {
				long := "https://www.google.com"
				short := register(t, server.URL, long)

				resp, err := http.Get(fmt.Sprintf("%s/%s", server.URL, short))
				if err != nil {
					t.Fatal(err)
				}
				if got, want := resp.Request.URL.String(), long; got != want {
					t.Errorf("got %v want %v", got, want)
				}
			})

			t.Run("not found", func(t *testing.T) {
				short := "foobar"
				resp, err := http.Get(fmt.Sprintf("%s/%s", server.URL, short))
				if err != nil {
					t.Fatal(err)
				}
				b, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(b), ErrNotFound.Error()) {
					t.Errorf("invalid error message: %s", b)
				}
			})
		})
	}
}

func register(t *testing.T, url string, long string) string {
	r := Register{Long: long}
	b, _ := json.Marshal(r)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", url, registerRoute), bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("invalid http response %d: %s", res.StatusCode, b)
	}

	var resp RegisterResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	return resp.Short
}
