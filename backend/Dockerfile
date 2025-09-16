# build
FROM golang:1.23.2 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/cyberark-shorten-url ./cmd/server/main.go

# runtime
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /src/bin/cyberark-shorten-url /app/cyberark-shorten-url
COPY configs ./configs

USER 65532:65532
ENTRYPOINT ["/app/cyberark-shorten-url"]