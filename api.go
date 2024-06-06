package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"math/big"
	"net/http"
	"strings"
)

const registerRoute = "/register"

type api struct {
	store     store
	shortener shortener
}

func newAPI(store store) api {
	return api{
		store:     store,
		shortener: sha256ShortenerB62,
	}
}

func (api api) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post(registerRoute, api.register)
	r.Get("/*", api.redirect)
	return r
}

type Register struct {
	Long string `json:"long"`
}

type RegisterResponse struct {
	Short string `json:"short"`
}

func (api api) register(w http.ResponseWriter, r *http.Request) {
	var register Register
	if err := json.NewDecoder(r.Body).Decode(&register); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}
	short := api.shortener(register.Long)
	if exist, _ := api.store.Exist(short); exist {
		RespondWithJSON(w, http.StatusOK, RegisterResponse{Short: short})
		return
	}
	if err := api.store.Save(short, register.Long); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	RespondWithJSON(w, http.StatusOK, RegisterResponse{Short: short})
}

func (api api) redirect(w http.ResponseWriter, r *http.Request) {
	long, err := api.store.Long(strings.TrimPrefix(r.URL.Path, "/"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	http.Redirect(w, r, long, http.StatusSeeOther)
}

func RespondWithError(w http.ResponseWriter, code int, msg interface{}) {
	var message string
	switch m := msg.(type) {
	case error:
		message = m.Error()
	case string:
		message = m
	}
	RespondWithJSON(w, code, JSONError{Error: message})
}

type JSONError struct {
	Error string `json:"error,omitempty"`
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

type shortener func(string) string

// sha256ShortenerB62 use base62 encoding, it has the advantage of removing special char from base64
var sha256ShortenerB62 = func(long string) string {
	hasher := sha256.New()
	hasher.Write([]byte(long))
	var i big.Int
	i.SetBytes(hasher.Sum(nil)[:])
	return i.Text(62)[0:6]
}
