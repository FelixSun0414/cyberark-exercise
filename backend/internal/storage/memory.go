package storage

import (
	"crypto/rand"
	"math/big"
	"sync"
)

type Store interface {
	// CreateShortenUrlCode - Create a new shorten URL from the input URL.
	//                        If the mapping already exists, return the shorten URL code and true,
	//                        otherwise, create a new code and return it with false.
	CreateShortenUrlCode(url string) (string, bool)
	// GetOriginalUrlFromCode - Get original URL from the shorten URL code.
	//                          Return empty string and false if not found,
	//                          otherwise, return the original Url and true.
	GetOriginalUrlFromCode(code string) (string, bool)
}

type MemoryStore struct {
	mu        sync.RWMutex
	codeToURL map[string]string
	urlToCode map[string]string
	alphabet  []rune
	codeLen   int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		codeToURL: make(map[string]string),
		urlToCode: make(map[string]string),
		alphabet:  []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"),
		codeLen:   7,
	}
}

func (s *MemoryStore) CreateShortenUrlCode(url string) (string, bool) {
	// Lock before checking
	s.mu.RLock()
	if code, ok := s.urlToCode[url]; ok {
		s.mu.RUnlock()
		// The shorten Url was found
		return code, true
	}
	s.mu.RUnlock()

	// Shorten Url does not exist yet, create one
	var code string
	for {
		code = s.randCode()
		s.mu.RLock()
		_, exists := s.codeToURL[code]
		s.mu.RUnlock()
		if !exists {
			break
		}
	}

	s.mu.Lock()
	s.codeToURL[code] = url
	s.urlToCode[url] = code
	s.mu.Unlock()
	return code, false
}

func (s *MemoryStore) GetOriginalUrlFromCode(code string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.codeToURL[code]
	return url, ok
}

func (s *MemoryStore) randCode() string {
	b := make([]rune, s.codeLen)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(s.alphabet))))
		b[i] = s.alphabet[idx.Int64()]
	}
	return string(b)
}
