package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

func main() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Task\tStatus\tTimestamp\t")
	fmt.Fprintf(w, "%s\t%s\t%s\t\n", "CalculateSum", "Success", time.Now())
	w.Flush()
}
