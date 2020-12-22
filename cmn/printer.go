package cmn

import (
	"fmt"
	"os"
	"text/tabwriter"
)

type Printer struct {
	w *tabwriter.Writer
}

func NewPrinter() Printer {
	return Printer{tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)}
}

func (p Printer) Printf(format string, args ...interface{}) {
	fmt.Fprintf(p.w, format, args...)
}

func (p Printer) Flush() {
	p.w.Flush()
}
