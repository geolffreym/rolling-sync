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


compile-win:
	GOOS=windows GOARCH=amd64 go build -o bin/main-windows-amd64 main.go
	GOOS=windows GOARCH=386 go build -o bin/main-windows-386 main.go

#Go1.15 deprecates 32-bit macOS builds	
#GOOS=darwin GOARCH=386 go build -o bin/main-mac-386 main.go
compile-mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/main-mac-amd64 main.go

compile-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/main-linux-amd64 main.go
	GOOS=linux GOARCH=386 go build -o bin/main-linux-386 main.go

compile: compile-linux compile-win compile-mac
	echo "Compiling for every OS and Platform"
	

all: build test check-test-coverage code-check compile