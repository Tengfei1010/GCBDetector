/*
================================================================
=  Source code from https://github.com/dominikh/go-tools       =
=  Copyright @ Dominik Honnef (https://github.com/dominikh)    =
================================================================
*/

// staticcheck detects a myriad of bugs and inefficiencies in your
// code.
package main

import (
	"github.com/Tengfei1010/GCBDetector/lint/lintutil"
	"github.com/Tengfei1010/GCBDetector/staticcheck"
	"os"
)

func main() {
	//path := []string {"/home/tensor/Develop/Go/src/honnef.co/go/tools/staticcheck/testdata/CheckDeferLock.go"}
	fs := lintutil.FlagSet("staticcheck")
	gen := fs.Bool("generated", false, "Check generated code")
	fs.Parse(os.Args[1:])
	//fs.Parse(path)
	c := staticcheck.NewChecker()
	c.CheckGenerated = *gen
	cfg := lintutil.CheckerConfig{
		Checker:     c,
		ExitNonZero: true,
	}
	lintutil.ProcessFlagSet([]lintutil.CheckerConfig{cfg}, fs)
}
