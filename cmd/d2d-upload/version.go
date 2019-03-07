package main

import (
	"fmt"
	"io"
	"runtime"
)

var (
	Version   string = "latest"
	GitCommit string = "latest"
	Built     string = "just now"
)

func PrintVersion(w io.Writer) {
	fmt.Fprintf(w, "d2d-upload\n")
	fmt.Fprintf(w, "  Version:      %s\n", Version)
	fmt.Fprintf(w, "  Go version:   %s\n", runtime.Version())
	fmt.Fprintf(w, "  Git commit:   %s\n", GitCommit)
	fmt.Fprintf(w, "  Built:        %s\n", Built)
}
