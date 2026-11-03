.PHONY: build test run clean

build:
	cd backend && go build -o vibelog-api .
	cd frontend && npm ci && npm run build

test:
	cd backend && go test -count=1 ./...
	cd frontend && npm ci && npm run typecheck

run:
	./scripts/run.sh

clean:
	rm -f backend/vibelog-api
	rm -rf frontend/dist frontend/node_modules
