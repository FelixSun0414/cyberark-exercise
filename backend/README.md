# URL Shortener Backend

This is a minimal FastAPI backend stub for the URL shortener exercise.
*This is a stub backend. It’s meant for testing frontend integration and demonstrating API behavior.

> Happy building! 🚀

## Running Locally

1. Install dependencies:
```bash
go mod download
```

2. Start the server:
```bash
go run ./cmd/main.go
```
The backend will be available at [http://localhost:5000]. To change the listening port, set it in the config.yaml file.

## API Endpoints (Stubbed)

### `POST /api/shorten`
- Accepts a JSON body with a URL.
- Responds with a hardcoded short code, a short URL and the original URL.

Example request:
```json
{ "url": "https://www.example.com" }
```

Example response:
```json
{
  "short_code": "abc123",
  "short_url": "http://localhost:5000/abc123",
  "original_url": "http://example.com"
}
```

### `GET /{short_code}`
- Redirects to a hardcoded URL (`https://example.com`) if the code is valid.
- Returns `404` if not found.

## Development Notes
- CORS is enabled for `http://localhost:3000` to support interaction with the frontend.
