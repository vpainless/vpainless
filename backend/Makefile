dist = dist
$(dist):
	mkdir -p dist

generate:
	go generate ./api

build: $(dist) generate
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o $(dist)/server cmd/server/main.go

run: dist generate
	go run cmd/server/main.go

build-mock-server: dist generate
	go build -o $(dist)/mockserver cmd/mockserver/*.go

test:
	go test -v -count=1 ./...

test-opa:
	opa test .

opa-lint:
	opa fmt -w .

lint: opa-lint
	golangci-lint run --fix --timeout=3m

clean_db:
	rm -rf data/*
