package fileio

import (
	"bufio"
	"reflect"
	"testing"
)

func TestInvalidChunkSize(t *testing.T) {
	IO := IO{blockSize: 1 << 8} // To big chunks with small text
	_, err := IO.Open("../mock.txt")

	if err == nil {
		t.Fatalf("Expected error for 'at least 2 chunks are required'")
	}

}

func TestFileOpen(t *testing.T) {
	IO := IO{blockSize: 1 << 4}
	reader, err := IO.Open("invalid.txt")

	if err == nil {
		t.Fatalf("Expected error for invalid file directory 'invalid.txt'")
	}

	if reflect.TypeOf(reader) != reflect.TypeOf(new(bufio.Reader)) {
		t.Fatalf("Expected bufio.Reader as reader")
	}

}

func TestFileChunks(t *testing.T) {
	IO := IO{blockSize: 1 << 4}
	IO.Open("mock.txt")
	chunks := IO.Chunks(87)

	if 6 != chunks {
		t.Fatalf("Expected 6 as result for int(math.Ceil(float64(87) / float64(16)))")
	}

}
