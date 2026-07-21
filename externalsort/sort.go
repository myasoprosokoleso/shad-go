//go:build !solution

package externalsort

import (
	"bufio"
	"container/heap"
	"io"
	"os"
	"slices"
	"strings"
)

type lineReader struct {
	r *bufio.Reader
}

func NewReader(r io.Reader) LineReader {
	return &lineReader{r: bufio.NewReader(r)}
}

func (l *lineReader) ReadLine() (string, error) {
	line, err := l.r.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			if line == "" {
				// call with empty buffer
				return "", io.EOF
			}
			// last line without \n
			return line, nil
		}
		return "", err
	}

	// unlike the bufio.Scanner, ReadString does't strip the delimiter!
	return strings.TrimSuffix(line, "\n"), nil
}

type lineWriter struct {
	w io.Writer
}

func NewWriter(w io.Writer) LineWriter {
	return &lineWriter{w: w}
}

func (l *lineWriter) Write(s string) error {
	_, err := l.w.Write([]byte(s + "\n"))
	return err
}

type entry struct {
	line   string
	reader LineReader
}

type lineReaderHeap []*entry

// const
func (h lineReaderHeap) Len() int           { return len(h) }
func (h lineReaderHeap) Less(i, j int) bool { return h[i].line < h[j].line }
func (h lineReaderHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *lineReaderHeap) Push(x any) {
	*h = append(*h, x.(*entry))
}

func (h *lineReaderHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func Merge(w LineWriter, readers ...LineReader) error {
	buf := make(lineReaderHeap, 0, len(readers))
	h := &buf

	for _, r := range readers {
		line, err := r.ReadLine()
		if err == io.EOF {
			continue
		}
		if err != nil {
			return err
		}

		heap.Push(h, &entry{line: line, reader: r})
	}

	for h.Len() > 0 {
		e := heap.Pop(h).(*entry)
		if err := w.Write(e.line); err != nil {
			return err
		}

		line, err := e.reader.ReadLine()
		if err == io.EOF {
			continue
		}
		if err != nil {
			return err
		}

		heap.Push(h, &entry{line: line, reader: e.reader})
	}

	return nil

}

func Sort(w io.Writer, in ...string) error {
	for _, path := range in {
		f, err := os.Open(path)
		if err != nil {
			return err
		}

		reader := NewReader(f)
		var lines []string
		for {
			line, err := reader.ReadLine()
			if err == io.EOF {
				break
			}
			if err != nil {
				_ = f.Close()
				return err
			}

			lines = append(lines, line)
		}
		// need to close before writing
		if err := f.Close(); err != nil {
			return err
		}
		slices.Sort(lines)

		f, err = os.Create(path)
		if err != nil {
			return err
		}

		writer := NewWriter(f)
		for _, line := range lines {
			if err := writer.Write(line); err != nil {
				_ = f.Close()
				return err
			}
		}
		// need to close, otherwise there may be too many open files
		if err := f.Close(); err != nil {
			return err
		}
	}

	readers := make([]LineReader, 0, len(in))
	opened := make([]*os.File, 0, len(in))
	for _, path := range in {
		f, err := os.Open(path)
		if err != nil {
			for _, o := range opened {
				_ = o.Close()
			}
			return err
		}

		opened = append(opened, f)
		readers = append(readers, NewReader(f))
	}

	defer func() {
		for _, o := range opened {
			_ = o.Close()
		}
	}()

	return Merge(NewWriter(w), readers...)
}
