FROM golang:1.26.2-alpine

WORKDIR /app

COPY notification-service/go.mod notification-service/go.sum ./notification-service/
COPY shared/go.mod shared/go.sum ./shared/

WORKDIR /app/notification-service
RUN go mod download

WORKDIR /app
COPY . .

WORKDIR /app/notification-service
RUN go build -o consumer ./cmd/consumer

CMD ["./consumer"]
