package sync

import (
	"bufio"
	"fmt"
)

type matchKind uint8

const (
	MATCH_KIND_LITERAL matchKind = iota
	MATCH_KIND_COPY
)

type Match struct {
	kind   matchKind
	pos    uint64
	len    uint64
	output *bufio.Writer
	lit    []byte
}

func intSize(d uint64) uint8 {
	switch {
	case d == uint64(uint8(d)):
		return 1
	case d == uint64(uint16(d)):
		return 2
	case d == uint64(uint32(d)):
		return 4
	default:
		return 8
	}
}

func (m *Match) flush() error {
	if m.len == 0 {
		return nil
	}

	posSize := intSize(m.pos)
	lenSize := intSize(m.len)

	switch m.kind {
	case MATCH_KIND_COPY:
		m.output.WriteString("COPY_START=")
		m.output.WriteString(fmt.Sprintf("%d:%d", m.pos, posSize))
		m.output.WriteString(fmt.Sprintf("%d:%d", m.len, lenSize))
		m.output.WriteString("\n")
	case MATCH_KIND_LITERAL:

		m.output.WriteString("LITERAL_START=")
		m.output.WriteString(fmt.Sprintf("%d:%d", m.len, lenSize))
		m.output.Write(m.lit)
		m.output.WriteString("\n")
		m.lit = []byte{}
	}
	m.pos = 0
	m.len = 0
	return nil
}

func (m *Match) add(kind matchKind, pos, len uint64) error {
	if len != 0 && m.kind != kind {
		err := m.flush()
		if err != nil {
			return err
		}
	}

	m.kind = kind
	switch kind {
	case MATCH_KIND_LITERAL:
		m.lit = append(m.lit, byte(pos))
		m.len += 1
	case MATCH_KIND_COPY:
		m.lit = []byte{}
		if m.pos+m.len != pos {
			err := m.flush()
			if err != nil {
				return err
			}
			m.pos = pos
			m.len = len
		} else {
			m.len += len
		}
	}
	return nil
}
