package model

type Configuration struct {
	SvcConfig ServiceConfig  `yaml:"urlShortenerService"`
	DbConfig  DatabaseConfig `yaml:"dbConfig"`
}

type ServiceConfig struct {
	Addr              string `yaml:"addr"`
	BaseURL           string `yaml:"base_url"`
	PermanentRedirect bool   `yaml:"permanent_redirect"`
	UrlLengthLimit    int    `yaml:"url_length_limit"`
}

type DatabaseConfig struct {
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	DatabaseHost string `yaml:"database_host"`
	DatabaseName string `yaml:"database_name"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
	OriginURL string `json:"origin_url"`
}
