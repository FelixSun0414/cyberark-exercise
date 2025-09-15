package model

type Config struct {
	Addr              string `yaml:"addr"`
	BaseURL           string `yaml:"base_url"`
	PermanentRedirect bool   `yaml:"permanent_redirect"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
	OriginURL string `json:"origin_url"`
}
