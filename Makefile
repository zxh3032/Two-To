FRONTEND_DIR := frontend
BACKEND_DIR := backend

.PHONY: frontend-install frontend-dev frontend-build frontend-lint frontend-test
frontend-install:
	cd $(FRONTEND_DIR) && npm install

frontend-dev:
	cd $(FRONTEND_DIR) && npm run dev

frontend-build:
	cd $(FRONTEND_DIR) && npm run build

frontend-lint:
	cd $(FRONTEND_DIR) && npm run lint

frontend-test:
	cd $(FRONTEND_DIR) && npm run test

.PHONY: backend-dev backend-test backend-test-linux backend-proto
backend-dev:
	cd $(BACKEND_DIR) && go run .

backend-test:
	cd $(BACKEND_DIR) && go test ./...

backend-test-linux:
	cd $(BACKEND_DIR) && GOOS=linux GOARCH=amd64 go test ./...

backend-proto:
	cd $(BACKEND_DIR) && ./scripts/gen-proto.sh

.PHONY: check
check: backend-test-linux frontend-lint frontend-test frontend-build
