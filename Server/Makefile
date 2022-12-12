build:
	@go build -o bin/server -v
run: build
	@bin/server -l=:8080
test: 
	@go test -v ./...