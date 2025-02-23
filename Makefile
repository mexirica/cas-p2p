build:
	@go build -o cmd/api/fs

run: build
	@./cmd/api/fs

test:
	@go test ./...