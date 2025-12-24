FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Самое важное — теги netgo + osusergo + CGO_ENABLED=0
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -a \
    -tags netgo,osusergo \
    -ldflags="-s -w" \
    -o /app/geo_not ./cmd/app/main.go

FROM scratch

WORKDIR /app
COPY --from=builder /app/geo_not /app/geo_not
COPY .env ./

EXPOSE 8080
CMD ["/app/geo_not"]