package ogle

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// TabWriter is a thin wrapper around `text/tabwriter` that prints any
// interface{} value.
type TabWriter struct {
	w *tabwriter.Writer
}

func NewTabWriter(w io.Writer) *TabWriter {
	return &TabWriter{
		w: tabwriter.NewWriter(w, 0, 0, 1, ' ', 0),
	}
}

func (tw *TabWriter) Println(args ...interface{}) {
	if len(args) < 1 {
		return
	}
	f := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		f = append(f, "%v")
	}
	fmt.Fprintf(tw.w, strings.Join(f, "\t")+"\n", args...)
}

func (tw *TabWriter) Flush() error {
	return tw.w.Flush()
}
