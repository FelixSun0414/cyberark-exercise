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
	"cyberark-shorten-url/internal/utility"
	"cyberark-shorten-url/pkg/logger"
)

type ApiRouter struct {
	DbClient          *storage.DbClient
	BaseURL           string
	PermanentRedirect bool
	UrlLengthLimit    int
}

// NewApiRouter - Create and return a new API router
func NewApiRouter(dbClient *storage.DbClient, baseURL string, permanentRedirect bool, urlLengthLimit int) *ApiRouter {
	return &ApiRouter{
		DbClient:          dbClient,
		BaseURL:           strings.TrimRight(baseURL, "/"),
		PermanentRedirect: permanentRedirect,
		UrlLengthLimit:    urlLengthLimit,
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

	// get original url from body
	var req model.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Info(err.Error())
		writeErrorInJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// validate if the URL is in acceptable pattern
	u, err := normalizeHTTPURL(req.URL)
	if err != nil {
		logger.Info(err.Error())
		writeErrorInJSON(w, http.StatusBadRequest, "invalid url format")
		return
	}
	// If there is a URL length limitation, check it
	if ar.UrlLengthLimit > 0 {
		if len(u.String()) > ar.UrlLengthLimit {
			logger.Info(fmt.Sprintf("Url length limit exceeded, total length of the original URL is %d", len(u.String())))
			writeErrorInJSON(w, http.StatusBadRequest, "url length exceeded")
			return
		}
	}

	// do not allow user to set any URL to redirect back to our service
	if strings.HasSuffix(u.String(), ar.BaseURL) {
		logger.Info(fmt.Sprintf("new URL is trying to redirect to our site: %s", u.String()))
		writeErrorInJSON(w, http.StatusBadRequest, "invalid url")
		return
	}

	// TODO: further security validation
	//       - check if the URL is a local loop to avoid Server Side Request Forgery
	//       - check if the URL link to a potential website which has security issue

	id, exists, code, err := storage.UpsertURL(ar.DbClient, u.String())
	if err != nil {
		// the UpsertURL is only doing insert and query, there should be no error at all
		// when this happens, something unexpected happened
		logger.Error(fmt.Sprintf("failed to query new record details, err: %s"), err.Error())
		http.Error(w, "Something unexpected happened, retry later", http.StatusInternalServerError)
		return
	}

	if id == 0 || id >= utility.Pow62_8 {
		// the id got from database is invalid, something unexpected happened
		logger.Error(fmt.Sprintf("failed to create a new record, new record has an id: %d"), id)
		http.Error(w, "Something unexpected happened, retry later", http.StatusInternalServerError)
		return
	}

	// the shorten code does not exist for this URL yet, generate a new code
	if !exists || code == "" {
		code, _ = utility.EncodeNumberToShortenCode(id)
		refreshedCode, claimed, updateErr := storage.SetCodeForId(ar.DbClient, id, code)
		// Something bad happened
		if updateErr != nil {
			logger.Error(fmt.Sprintf("failed to set short code for the new record, new record id: %d, err: %s"), id, updateErr.Error())
			http.Error(w, "Something unexpected happened, retry later", http.StatusInternalServerError)
			return
		}

		code = refreshedCode
		exists = !claimed
	}

	res := model.ShortenResponse{
		ShortCode: code,
		ShortURL:  fmt.Sprintf("%s/%s", ar.BaseURL, code),
		OriginURL: u.String(),
	}

	status := http.StatusCreated
	if exists {
		status = http.StatusOK
	}
	logger.Info("shorten succeeded", "new code", code, "url", u.String(), "exists", exists)
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
	if code == "" || strings.HasPrefix(code, "api/") || len(code) < 5 || len(code) > 8 {
		logger.Info(fmt.Sprintf("Wrong request with path: %s", r.URL.Path))
		writeErrorInJSON(w, http.StatusBadRequest, "invalid shorten code")
		return
	}

	// fetch original URL
	u, found, err := storage.GetURLByCode(ar.DbClient, code)
	if err != nil {
		logger.Error(fmt.Sprintf("Unexpected error during fetching original URL for code: %s, err: %s"), code, err.Error())
		http.Error(w, "Something unexpected happened, retry later", http.StatusInternalServerError)
		return
	}

	if found {
		// found original URL, redirect
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
