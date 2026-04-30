FROM golang:1.26.2-alpine

WORKDIR /app

# copy module files (from repo root context)
COPY order-service/go.mod order-service/go.sum ./order-service/

WORKDIR /app/order-service
RUN go mod download

# copy everything
WORKDIR /app
COPY . .

WORKDIR /app/order-service

# build
RUN go build -o app ./cmd/server

EXPOSE 8080

CMD ["./app"]
