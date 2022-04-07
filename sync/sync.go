package sync

import (
	"crypto/md5"
	"hash/adler32"
)

type Sync struct {
	blockSize int
}

func New(size int) *Sync {
	return &Sync{
		blockSize: size,
	}
}

func (s *Sync) Signature(block []byte) (uint32, []byte) {
	adler := adler32.New()
	adler.Write(block)

	md5 := md5.New()
	md5.Write(block)
	return adler.Sum32(), md5.Sum(nil)
}
