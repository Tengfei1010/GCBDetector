/*
================================================================
=  Source code from https://github.com/dominikh/go-tools       =
=  Copyright @ Dominik Honnef (https://github.com/dominikh)    =
================================================================
*/

package staticcheck

import (
	"testing"

	"github.com/Tengfei1010/GCBDetector/lint"
	"github.com/Tengfei1010/GCBDetector/lint/lintutil"
	"github.com/Tengfei1010/GCBDetector/lint/testutil"
)

func TestAll(t *testing.T) {
	c := NewChecker()
	testutil.TestAll(t, c, "")
}

func BenchmarkStdlib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := NewChecker()
		_, err := lintutil.Lint([]lint.Checker{c}, []string{"std"}, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNetHttp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := NewChecker()
		_, err := lintutil.Lint([]lint.Checker{c}, []string{"net/http"}, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
