# Small make tasks for go
.PHONY: test

BINARY=main
BINARY_WIN=${BINARY}-win
BINARY_OSX=${BINARY}-darwin
BINARY_LINUX=${BINARY}-linux

ARCH_64=amd64
ARCH_32=386

LINUX_64=${BINARY_LINUX}-${ARCH_64}
LINUX_32=${BINARY_LINUX}-${ARCH_32}
WIN_64=${BINARY_WIN}-${ARCH_64}
WIN_32=${BINARY_WIN}-${ARCH_32}
OSX_64=${BINARY_LINUX}-${ARCH_64}


test:
	go test -v ./...

# Could be compared using
# make benchmark > a.old
# make benchmark > b.new
# benchcmp a.old b.new
benchmark: 
	go test ./... -bench=. -benchtime 100000x -count 5

# View profiling
# Could use Graphviz (https://graphviz.org/download/)
# eg. go tool pprof -web bin/main-linux-amd64 cpu.prof
# or just standard output: go tool pprof -top cpu.prof 

# For memory profiling type use:
# inuse_space	Display in-use memory size
# inuse_objects	Display in-use object counts
# alloc_space	Display allocated memory size
# alloc_objects	Display allocated object counts
# eg. go tool pprof --alloc_space -top prof.mem 

profiling: 
	go test -bench=. -benchtime 100000x -run=^$ -cpuprofile=cpu.prof
	go test -bench=. -benchtime 100000x -run=^$ -memprofile=prof.mem

coverage:
	go test -coverprofile coverage ./...
	
coverage-export: coverage
	go tool cover -html=coverage

build:
	go build -v ./...

code-check:
	go vet -v ./...

clean:
	go clean --cache ./... 
	rm -f mem.prof
	rm -f prof.mem
	rm -rf bin

compile-win:
	GOOS=windows GOARCH=amd64 go build -o bin/${WIN_64} main.go
	GOOS=windows GOARCH=386 go build -o bin/${WIN_32} main.go

#Go1.15 deprecates 32-bit macOS builds	
#GOOS=darwin GOARCH=386 go build -o bin/main-mac-386 main.go
compile-mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/${OSX_64} main.go

compile-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/${LINUX_64} main.go
	GOOS=linux GOARCH=386 go build -o bin/${LINUX_32} main.go

compile: compile-linux compile-win compile-mac
	echo "Compiling for every OS and Platform"
	

all: build test check-test-coverage code-check compile