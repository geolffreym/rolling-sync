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

	rolling = rolling.RollIn('o')
	rolling = rolling.RollIn('w')
	rolling = rolling.RollIn(' ')
	rolling = rolling.RollIn('a')
	rolling = rolling.RollIn('r')
	rolling = rolling.RollIn('e')
	rolling = rolling.RollIn(' ')
	rolling = rolling.RollIn('y')
	rolling = rolling.RollIn('o')
	rolling = rolling.RollIn('u')
	rolling = rolling.RollIn(' ')
	rolling = rolling.RollIn('d')
	rolling = rolling.RollIn('o')
	rolling = rolling.RollIn('i')
	rolling = rolling.RollIn('n')
	rolling = rolling.RollIn('g')
	w1 := rolling.Sum()

	if w0 != w1 {
		t.Errorf("Expected same hash for same input after RolledIn bytes")
	}

}

func TestRollOut(t *testing.T) {
	rolling := New()

	w0 := rolling.Write([]byte("w are you doing")).Sum()

	rolling = rolling.RollIn('h')
	rolling = rolling.RollIn('o')
	rolling = rolling.RollIn('w')
	rolling = rolling.RollIn(' ')
	rolling = rolling.RollIn('a')
	rolling = rolling.RollIn('r')
	rolling = rolling.RollIn('e')
	rolling = rolling.RollIn(' ')
	rolling = rolling.RollIn('y')
	rolling = rolling.RollIn('o')
	rolling = rolling.RollIn('u')
	rolling = rolling.RollIn(' ')
	rolling = rolling.RollIn('d')
	rolling = rolling.RollIn('o')
	rolling = rolling.RollIn('i')
	rolling = rolling.RollIn('n')
	rolling = rolling.RollIn('g')
	rolling = rolling.RollOut() // remove h
	rolling = rolling.RollOut() // remove o
	w1 := rolling.Sum()

	if w0 != w1 {
		t.Errorf("Expected same hash for same text after RolledOut byte")
	}

}
