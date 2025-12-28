# ---- Build stage ----
FROM golang:1.24-alpine AS build
LABEL authors="kevin fernandes"

WORKDIR /src

# Copy go module files first (better Docker caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api

# ---- Runtime stage ----
FROM alpine:3.20

WORKDIR /app

# Copy the compiled binary
COPY --from=build /out/api /app/api

EXPOSE 8080

CMD ["/app/api"]
