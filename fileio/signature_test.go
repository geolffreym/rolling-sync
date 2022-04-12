package fileio

import (
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/geolffreym/rolling-sync/sync"
)

func TestSignatureReadWrite(t *testing.T) {
	// Read file to split in chunks
	blockSize := 1 << 4 // 16 bytes
	io := IO{blockSize: 1 << 4}
	sync := sync.New(blockSize)

	// Memory performance improvement using bufio.Reader
	reader, err := io.Open("../mock.txt")
	if err != nil {
		log.Fatal(err)
	}

	// For each block slice from file
	sync.FillTable(reader)
	signatures := sync.Signatures()
	io.Signature.Write("signature.bin", signatures)
	out, _ := io.Signature.Read("signature.bin")

	if !reflect.DeepEqual(signatures, out) {
		t.Errorf("Expected written signatures equal to out signatures")
	}

}

func TestSignatureBadWrite(t *testing.T) {
	io := IO{blockSize: 1 << 4}
	signatures := []sync.Table{}
	err := io.Signature.Write("signature.bin", signatures)

	if err == nil {
		t.Error("Expected error with invalid signatures to write")
	}
}

func TestSignatureBadFileWrite(t *testing.T) {
	io := IO{blockSize: 1 << 4}
	signatures := []sync.Table{}
	err := io.Signature.Write("notexists.bin", signatures)

	if err == nil {
		t.Error("Expected error with invalid file to write")
	}
}

func TestSignatureBadFileRead(t *testing.T) {
	io := IO{blockSize: 1 << 4}
	_, err := io.Signature.Read("notexists.bin")

	if err == nil {
		t.Error("Expected error with invalid file to read")
	}
}

func TestSignatureBadDataRead(t *testing.T) {
	file := "invalid.bin"
	io := IO{blockSize: 1 << 4}

	//  Performed writing operations
	f, _ := os.Create(file)
	f.WriteString("I am invalid gob")
	_, err := io.Signature.Read(file)

	if err == nil {
		t.Error("Expected error with invalid file gob data content")
	}
}
