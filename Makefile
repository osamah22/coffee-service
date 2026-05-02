.PHONY: test test-order test-notification test-shared frontend-build docker-config check

test: test-order test-notification test-shared

test-order:
	cd order-service && go test ./...

test-notification:
	cd notification-service && go test ./...

test-shared:
	cd shared && go test ./...

frontend-build:
	cd frontend && npm run build

docker-config:
	docker compose config

check: test frontend-build docker-config
