package cmd

import (
	"bytes"
	"io"
	"sync"
)

const (
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan  = "\033[36m"
	colorReset = "\033[0m"
)

type colorWriter struct {
	w   io.Writer
	buf bytes.Buffer
	mu  sync.Mutex
}

func newColorWriter(w io.Writer) *colorWriter {
	return &colorWriter{w: w}
}

func (cw *colorWriter) Write(p []byte) (int, error) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	cw.buf.Write(p)

	for {
		line, err := cw.buf.ReadBytes('\n')
		if err != nil {
			cw.buf.Write(line)
			break
		}

		line = cw.colorize(line)
		if _, err := cw.w.Write(line); err != nil {
			return len(p), err
		}
	}

	return len(p), nil
}

func (cw *colorWriter) colorize(line []byte) []byte {
	var color string
	switch {
	case bytes.Contains(line, []byte("level=ERROR")):
		color = colorRed
	case bytes.Contains(line, []byte("level=WARN")):
		color = colorYellow
	case bytes.Contains(line, []byte("level=DEBUG")):
		color = colorCyan
	case bytes.Contains(line, []byte("level=INFO")):
		color = colorGreen
	default:
		return line
	}

	result := make([]byte, 0, len(line)+len(color)+len(colorReset))
	result = append(result, []byte(color)...)
	result = append(result, line...)
	result = append(result, []byte(colorReset)...)
	return result
}
