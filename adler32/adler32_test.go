package adler32

import (
	"testing"
)

func TestWriteSum(t *testing.T) {
	rolling := New()

	rolling = rolling.Write([]byte("how are you doing"))
	w0 := rolling.Sum()

	if 944178772 != w0 {
		t.Errorf("Expected 944178772 as hash for input text")
	}

	if 17 != rolling.Count() {
		t.Errorf("Expected 17 as output for current window")
	}
}

func TestWindowOverflow(t *testing.T) {
	rolling := New()

	rolling = rolling.Write([]byte("abcdef"))
	rolling = rolling.RollOut() // remove a
	rolling = rolling.RollOut() // remove b
	rolling = rolling.RollOut() // remove c
	rolling = rolling.RollOut() // remove d
	rolling = rolling.RollOut() // remove e
	rolling = rolling.RollOut() // remove f
	rolling = rolling.RollOut() // overflow

	if rolling.Count() > 0 {
		t.Errorf("Expected error 'Window size equal 0'")
	}
}

func TestRollIn(t *testing.T) {
	rolling := New()

	w0 := rolling.Write([]byte("ow are you doing")).Sum()
	w1 := rolling.
		RollIn('o').
		RollIn('w').
		RollIn(' ').
		RollIn('a').
		RollIn('r').
		RollIn('e').
		RollIn(' ').
		RollIn('y').
		RollIn('o').
		RollIn('u').
		RollIn(' ').
		RollIn('d').
		RollIn('o').
		RollIn('i').
		RollIn('n').
		RollIn('g').Sum()

	if w0 != w1 {
		t.Errorf("Expected same hash for same input after RolledIn bytes")
	}

}

func TestRollOut(t *testing.T) {
	rolling := New()

	w0 := rolling.Write([]byte("w are you doing")).Sum()
	w1 := rolling.RollIn('h').
		RollIn('o').
		RollIn('w').
		RollIn(' ').
		RollIn('a').
		RollIn('r').
		RollIn('e').
		RollIn(' ').
		RollIn('y').
		RollIn('o').
		RollIn('u').
		RollIn(' ').
		RollIn('d').
		RollIn('o').
		RollIn('i').
		RollIn('n').
		RollIn('g').
		RollOut(). // remove h
		RollOut(). // remove o
		Sum()

	if w0 != w1 {
		t.Errorf("Expected same hash for same text after RolledOut byte")
	}

}
