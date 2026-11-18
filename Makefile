.PHONY: build test run clean

build:
	cd frontend && npm ci && npm run build
	rm -rf internal/web/dist && mkdir -p internal/web/dist
	cp -r frontend/dist/. internal/web/dist/
	touch internal/web/dist/.gitkeep
	CGO_ENABLED=0 go build -trimpath -o vibelog ./cmd/vibelog

test:
	CGO_ENABLED=0 go test -count=1 ./...
	cd frontend && npm ci && npm run typecheck

run:
	./scripts/run.sh

clean:
	rm -f vibelog
	rm -rf frontend/dist frontend/node_modules
