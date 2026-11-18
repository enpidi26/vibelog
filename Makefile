.PHONY: build test run clean

build:
	cd frontend && npm ci && npm run build
	rm -rf backend/web/dist && mkdir -p backend/web/dist
	cp -r frontend/dist/. backend/web/dist/
	touch backend/web/dist/.gitkeep
	cd backend && CGO_ENABLED=0 go build -trimpath -o vibelog .

test:
	cd backend && CGO_ENABLED=0 go test -count=1 ./...
	cd frontend && npm ci && npm run typecheck

run:
	./scripts/run.sh

clean:
	rm -f backend/vibelog backend/vibelog-api
	rm -rf frontend/dist frontend/node_modules
