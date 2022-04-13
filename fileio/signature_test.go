package fileio

import (
	"os"
	"reflect"
	"testing"

	"github.com/geolffreym/rolling-sync/sync"
)

func TestSignatureReadWrite(t *testing.T) {
	// Read file to split in chunks
	signature := sync.Table{Weak: 0000, Strong: "abc123"}
	signatures := []sync.Table{signature}
	WriteSignature("signature.bin", signatures)
	out, _ := ReadSignature("signature.bin")

	if !reflect.DeepEqual(signatures, out) {
		t.Errorf("Expected written signatures equal to out signatures")
	}

}

func TestSignatureBadWrite(t *testing.T) {
	signatures := []sync.Table{}
	err := WriteSignature("signature.bin", signatures)

	if err == nil {
		t.Error("Expected error with invalid signatures to write")
	}
}

func TestSignatureBadFileWrite(t *testing.T) {
	signatures := []sync.Table{}
	err := WriteSignature("notexists.bin", signatures)

	if err == nil {
		t.Error("Expected error with invalid file to write")
	}
}

func TestSignatureBadFileRead(t *testing.T) {
	_, err := ReadSignature("notexists.bin")

	if err == nil {
		t.Error("Expected error with invalid file to read")
	}
}

func TestSignatureBadDataRead(t *testing.T) {
	file := "invalid.bin"

	//  Performed writing operations
	f, _ := os.Create(file)
	f.WriteString("I am invalid gob")
	_, err := ReadSignature(file)

	if err == nil {
		t.Error("Expected error with invalid file gob data content")
	}
}
