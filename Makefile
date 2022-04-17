# Small make tasks for go
.PHONY: test
test:
	go test -v ./...

check-test-coverage:
	go test -coverprofile coverage ./...

build:
	go build -v ./...

code-check:
	go vet -v ./...

compile: 
	echo "Compiling for every OS and Platform"
	GOOS=freebsd GOARCH=386 go build -o bin/main-freebsd-386 main.go
	GOOS=linux GOARCH=386 go build -o bin/main-linux-386 main.go
	GOOS=windows GOARCH=386 go build -o bin/main-windows-386 main.go

all: build test check-test-coverage code-check compile