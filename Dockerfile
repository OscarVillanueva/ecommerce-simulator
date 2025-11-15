FROM golang:latest

WORKDIR /app

COPY backend/ ./api/

WORKDIR /app/api

RUN go mod tidy

CMD ["go", "run", "cmd/api/main.go"]

