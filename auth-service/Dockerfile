FROM golang:1.26.2-alpine

WORKDIR /app

COPY auth-service/go.mod auth-service/go.sum ./auth-service/
COPY shared/go.mod shared/go.sum ./shared/

WORKDIR /app/auth-service
RUN go mod download

WORKDIR /app
COPY . .

WORKDIR /app/auth-service
RUN go build -o app ./cmd/server

EXPOSE 8081

CMD ["./app"]
