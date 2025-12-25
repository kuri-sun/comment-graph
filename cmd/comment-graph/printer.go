package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

type printer struct {
	warn   *color.Color
	err    *color.Color
	ok     *color.Color
	header *color.Color
}

func newPrinter() printer {
	color.NoColor = !shouldUseColor()
	return printer{
		warn:   color.New(color.FgYellow, color.Bold),
		err:    color.New(color.FgRed, color.Bold),
		ok:     color.New(color.FgGreen, color.Bold),
		header: color.New(color.Bold),
	}
}

func shouldUseColor() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	return true
}

func (p printer) infof(format string, args ...any) {
	fmt.Printf("  "+format+"\n", args...)
}

func (p printer) warnLine(msg string) {
	fmt.Fprintln(os.Stderr, p.warn.Sprint("[warn]"), msg)
}

func (p printer) errLine(msg string) {
	fmt.Fprintln(os.Stderr, p.err.Sprint("[err]"), msg)
}

func (p printer) okLine(msg string) {
	fmt.Println(p.ok.Sprint("[ok]"), msg)
}

func (p printer) section(msg string) {
	fmt.Println(p.header.Sprint("> " + msg))
}

func (p printer) resultLine(ok bool) {
	status := "failed"
	statusColored := p.err.Sprint(status)
	if ok {
		status = "succeeded"
		statusColored = p.ok.Sprint(status)
	}
	fmt.Println("  result :", statusColored)
}
