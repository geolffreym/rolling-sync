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

	rolling.Write([]byte("ow are you doing"))
	w0 := rolling.Sum()

	rolling.Reset()

	rolling.RollIn('h')
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
	rolling.RollOut()
	w1 := rolling.Sum()

	if w0 != w1 {
		t.Errorf("Expected same hash for same text after RolledOut byte")
	}

}

// fmt.Printf("%d\n", len(w))
// w0 := s.w.Sum()

// s.w.Reset()
// s.w.RollIn('h')
// s.w.RollIn('o')
// s.w.RollIn('w')
// s.w.RollIn(' ')
// s.w.RollIn('a')
// s.w.RollIn('r')
// s.w.RollIn('e')
// s.w.RollIn(' ')
// s.w.RollIn('y')
// s.w.RollIn('o')
// s.w.RollIn('u')
// s.w.RollIn(' ')
// s.w.RollIn('d')
// s.w.RollIn('o')
// s.w.RollIn('i')
// s.w.RollIn('n')
// w1 := s.w.Sum()
// s.w.RollOut('h')

// w2 := s.w.Sum()
// s.w.RollIn('g')
// w3 := s.w.Sum()

// w = s.w.Window

// fmt.Printf("%d\n", len(w))
// fmt.Printf("%s\n", w)
// fmt.Printf("%d-%d-%d-%d", w0, w1, w2, w3)
