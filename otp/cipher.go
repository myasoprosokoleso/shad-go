//go:build !solution

package otp

import (
	"io"
)

type cipherReader struct {
	r    io.Reader
	prng io.Reader
}

type cipherWriter struct {
	w    io.Writer
	prng io.Reader
}

func NewReader(r io.Reader, prng io.Reader) io.Reader {
	return &cipherReader{r: r, prng: prng}
}

func NewWriter(w io.Writer, prng io.Reader) io.Writer {
	return &cipherWriter{w: w, prng: prng}
}

// Read should always return n == 0, if len(p) == 0. It may return a
// non-nil error if some error condition is known, such as EOF.
//
// An instance of this general case is that a Reader returning
// a non-zero number of bytes at the end of the input stream may
// return either err == EOF or err == nil. The next Read should
// return 0, EOF.
//
// Implementations of Read are discouraged from returning a
// zero byte count with a nil error, except when len(p) == 0.
// Callers should treat a return of 0 and nil as indicating that
// nothing happened; in particular it does not indicate EOF.
func (c *cipherReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	prngBuf := make([]byte, n)
	// ignoring prng errors according to the README
	_, _ = io.ReadFull(c.prng, prngBuf)

	for i := range n {
		p[i] ^= prngBuf[i]
	}

	return n, err
}

// Write must return a non-nil error if it returns n < len(p).
func (c *cipherWriter) Write(p []byte) (int, error) {
	prngBuf := make([]byte, len(p))
	// ignoring prng errors according to the README
	_, _ = io.ReadFull(c.prng, prngBuf)

	resBuf := make([]byte, len(p))
	copy(resBuf, p)
	for i := range resBuf {
		resBuf[i] ^= prngBuf[i]
	}

	n, err := c.w.Write(resBuf)
	if err != nil {
		return n, err
	}

	if n < len(p) {
		return n, io.ErrShortWrite
	}

	return n, nil
}
