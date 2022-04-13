package sync

import (
	"bufio"
	"bytes"
	"testing"
)

/**
Test for matching differences

a mock.txt
b mockV2.txt
a= i am here guys how are you doing this is a small test for chunk split and rolling hash
b= i here guys how are you doing this is a small test chunk split and rolling hash

0 = i am here guys h
1 = ow are you doing
2 = this is a small
3 =  test for chunk
4 = split and rollin
5 = g hash

Signatures:
[
	0: {768804235 5d7b9b82d3dd8c4d13a576f004318130},
	1: {828311020 eb535617e82301559e56a18993cdbe39},
	2: {763037070 f38bd8f1d59e45f4a7bdaa6311064573},
	3: {800720288 6f2fcd27d23f5e98f486ff34ad580d09},
	4: {880805423 5f19d42bfb610b9861ec0704b6467910},
	5: {489488949 3024133c176e89ed84401db125a62ed0}
]


**/

func CalculateDelta(a []byte, b []byte) map[int]Bytes {

	sync := New(1 << 4) // 16 bytes

	// Memory performance improvement using bufio.Reader
	bytesA := bytes.NewReader(a)
	bytesB := bytes.NewReader(b)

	bufioA := bufio.NewReader(bytesA)
	bufioB := bufio.NewReader(bytesB)

	// For each block slice from file
	sync.FillTable(bufioA)
	signatures := sync.Signatures()
	// TODO here signatures could be written and read from file
	// TODO using same signatures directly
	return sync.Delta(signatures, bufioB)
}

func CheckMatch(delta map[int]Bytes, expected map[int][]byte, t *testing.T) {

	for i := range expected {
		// Index not matched in delta
		if _, ok := delta[i]; !ok {
			t.Errorf("Expected match corresponding index for delta %d", i)
		}

		literal := delta[i].Lit
		expect := expected[i]
		if string(literal) != string(expect) {
			t.Errorf("Expected match difference %s = %s ", literal, expect)
		}
	}
}

func TestDetectChunkChange(t *testing.T) {
	a := []byte("i am here guys how are you doing this is a small test for chunk split and rolling hash")
	b := []byte("i here guys how are you doing this is a mall test chunk split and rolling hash")
	expect := map[int][]byte{
		1: []byte("i here guys h"),               // Match first block change
		4: []byte(" this is a mall test chunk "), // Match blocks 4 changed

	}

	delta := CalculateDelta(a, b)
	CheckMatch(delta, expect, t)

}

func TestDetectChunkAdd(t *testing.T) {
	a := []byte("i am here guys how are you doing this is a small test for chunk split and rolling hash")
	b := []byte("i am here guys how are you doingadded this is a small test for chunk split and rolling hash")
	expect := map[int][]byte{
		2: []byte("added"), // Match blocks 2 changed

	}
	delta := CalculateDelta(a, b)
	CheckMatch(delta, expect, t)

}

func TestDetectChunkRemoval(t *testing.T) {
	a := []byte("i am here guys how are you doing this is a small test for chunk split and rolling hash")
	b := []byte("ow are you doing test for chunk split and rolling hash")
	delta := CalculateDelta(a, b)

	if delta[0].Missing == false {
		t.Errorf("Expected delta first block missing")
	}

	if delta[0].Start != 1 && delta[0].Offset != 16 {
		t.Errorf("Expected delta range for missing block = 0-16")
	}
}

func TestDetectChunkShifted(t *testing.T) {
	o := []byte("i am here guys how are you doing this is a small test for chunk split and rolling hash")
	c := []byte("i am here guys   how are you doing test for chunk split and rolling hash")
	expect := map[int][]byte{
		1: []byte("i am here guys   h"), // Match first block change
	}

	delta := CalculateDelta(o, c)
	CheckMatch(delta, expect, t)
}
