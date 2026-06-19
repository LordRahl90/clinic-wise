all: clean lint test
run:
	go run ./cmd/api
test:
	@go test ./... --cover

lint:
	golangci-lint run ./... --timeout=3m

test-with-race:
	@go test -race ./... --cover

shuffle-test:
	@go test -shuffle=on --count=2 ./... -v

clean:
	go clean --testcache

build-image:
	docker build -t lordrahl/clinic-wise .