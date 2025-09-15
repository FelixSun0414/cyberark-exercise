package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"cyberark-shorten-url/internal/model"
	"cyberark-shorten-url/internal/storage"
	"cyberark-shorten-url/pkg/logger"
)

type ApiRouter struct {
	Store             storage.Store
	BaseURL           string
	PermanentRedirect bool
}

// NewApiRouter - Create and return a new API router
func NewApiRouter(store storage.Store, baseURL string, permanentRedirect bool) *ApiRouter {
	return &ApiRouter{
		Store:             store,
		BaseURL:           strings.TrimRight(baseURL, "/"),
		PermanentRedirect: permanentRedirect,
	}
}

// RegisterRoutes - registers all endpoints on a ServeMux.
func (ar *ApiRouter) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/shorten", ar.Shorten)
	mux.HandleFunc("/", ar.Redirect)
}

// Shorten - POST /api/shorten
func (ar *ApiRouter) Shorten(w http.ResponseWriter, r *http.Request) {
	logger.Info(fmt.Sprintf("New Shorten request: =>%v<=", r))

	if r.Method != http.MethodPost {
		logger.Info("Wrong method on POST /api/shorten")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Info(err.Error())
		writeErrorInJSON(w, http.StatusBadRequest, "invalid json")
		return
	}
	u, err := normalizeHTTPURL(req.URL)
	if err != nil {
		logger.Info(err.Error())
		writeErrorInJSON(w, http.StatusBadRequest, "invalid url")
		return
	}

	code, existed := ar.Store.CreateShortenUrlCode(u.String())
	res := model.ShortenResponse{
		ShortCode: code,
		ShortURL:  fmt.Sprintf("%s/%s", ar.BaseURL, code),
		OriginURL: u.String(),
	}

	status := http.StatusCreated
	if existed {
		status = http.StatusOK
	}
	logger.Info("shorten succeeded", "code", code, "url", u.String(), "existed", existed)
	writeJSON(w, status, res)
}

// Redirect - GET /{short_code}
func (ar *ApiRouter) Redirect(w http.ResponseWriter, r *http.Request) {
	logger.Info(fmt.Sprintf("New Redirect request: %v", r))

	if r.Method != http.MethodGet {
		logger.Info("Wrong method on GET /{shorten_url}")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := strings.TrimPrefix(r.URL.Path, "/")
	if code == "" || strings.HasPrefix(code, "api/") {
		logger.Info(fmt.Sprintf("Wrong request with path: %s", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	if u, ok := ar.Store.GetOriginalUrlFromCode(code); ok {
		if ar.PermanentRedirect {
			http.Redirect(w, r, u, http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, r, u, http.StatusFound)
		}
		logger.Info(fmt.Sprintf("Found the original Url [%s] for path: [%s]", u, r.URL.Path))
		return
	}
	logger.Info(fmt.Sprintf("Failed to find the original Url for path [%s]", r.URL.Path))
	http.NotFound(w, r)
}

func writeErrorInJSON(w http.ResponseWriter, status int, errMsg string) {
	rspBody := map[string]string{"error": errMsg}
	writeJSON(w, status, rspBody)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// normalizeHTTPURL - checks if a string is a valid HTTP/HTTPS URL
func normalizeHTTPURL(urlStr string) (*url.URL, error) {
	if urlStr == "" {
		return nil, errors.New("cannot get path from an empty URL")
	}

	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error parsing URL: %v", err))
	}

	if u.Scheme == "" || u.Host == "" {
		return nil, errors.New(fmt.Sprintf("URL %s does not have a valid scheme or host", urlStr))
	}

	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return nil, errors.New(fmt.Sprintf("Invalid URL: %s, only schema http and https are supported by now", urlStr))
	}

	return u, nil
}
