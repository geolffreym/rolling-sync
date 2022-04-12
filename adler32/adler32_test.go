package adler32

import (
	"testing"
)

func TestWriteSum(t *testing.T) {
	rolling := &Adler32{}

	rolling.Write([]byte("how are you doing"))
	w0 := rolling.Sum()

	if 944178772 != w0 {
		t.Errorf("Expected 944178772 as hash for input text")
	}

	if 17 != rolling.Count() {
		t.Errorf("Expected 17 as output for current window")
	}
}

func TestRollIn(t *testing.T) {
	rolling := &Adler32{}

	rolling.Write([]byte("ow are you doing"))
	w0 := rolling.Sum()

	rolling.Reset()

	rolling.RollIn('o')
	rolling.RollIn('w')
	rolling.RollIn(' ')
	rolling.RollIn('a')
	rolling.RollIn('r')
	rolling.RollIn('e')
	rolling.RollIn(' ')
	rolling.RollIn('y')
	rolling.RollIn('o')
	rolling.RollIn('u')
	rolling.RollIn(' ')
	rolling.RollIn('d')
	rolling.RollIn('o')
	rolling.RollIn('i')
	rolling.RollIn('n')
	rolling.RollIn('g')
	w1 := rolling.Sum()

	if w0 != w1 {
		t.Errorf("Expected same hash for same input after RolledIn bytes")
	}

}

func TestRollOut(t *testing.T) {
	rolling := &Adler32{}

	rolling.Write([]byte("w are you doing"))
	w0 := rolling.Sum()

	rolling.Reset()

	rolling.RollIn('h') // remove this
	rolling.RollIn('o') // remove this
	rolling.RollIn('w')
	rolling.RollIn(' ')
	rolling.RollIn('a')
	rolling.RollIn('r')
	rolling.RollIn('e')
	rolling.RollIn(' ')
	rolling.RollIn('y')
	rolling.RollIn('o')
	rolling.RollIn('u')
	rolling.RollIn(' ')
	rolling.RollIn('d')
	rolling.RollIn('o')
	rolling.RollIn('i')
	rolling.RollIn('n')
	rolling.RollIn('g')
	rolling.RollOut()
	rolling.RollOut()
	w1 := rolling.Sum()

	if w0 != w1 {
		t.Errorf("Expected same hash for same text after RolledOut byte")
	}

}
