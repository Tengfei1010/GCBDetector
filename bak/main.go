package main

import (
	"go/build"
	"go/token"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/ssa/ssautil"
	"golang.org/x/tools/go/ssa"
	"bytes"
	"fmt"

	s "strings"
	"path/filepath"
	"os"
	"io/ioutil"
)

//// anony function local share variable data race detection
//func aflvDetector(function *ssa.Function) {
//
//	anonyFs := function.AnonFuncs
//
//	if len(anonyFs) <= 0 {
//		return
//	}
//
//	fmt.Println(function.Name())
//	isWriteInAf := false
//	for _, fa := range anonyFs {
//		freeVars := fa.FreeVars
//		if len(freeVars) > 0 {
//			for _, fv := range freeVars {
//				for _, user := range *fv.Referrers() {
//					switch instr := user.(type) {
//
//					case *ssa.Store:
//						isWriteInAf = true
//						fmt.Println("write in anonyfunction....", instr)
//					}
//				}
//			}
//		}
//	}
//
//	if isWriteInAf {
//		// TODO:: behind define , if there is read or write, it will race
//	} else {
//		// TODO: behind define, if there is write, it will race
//	}
//
//}


func constainsAtomic64 (iiStr string) bool {
	atomic64Strs := [8]string {"AddInt64", "AddUint64",
	"LoadInt64", "LoadUint64", "StoreInt64",
	"StoreUint64", "SwapInt64", "SwapUint64"}

	for _, str := range atomic64Strs {
		if s.Contains(iiStr, str) {
			return true
		}
	}

	return false
}

// atomic uint64 is not safe in 32bits system
// https://golang.org/pkg/sync/atomic/#pkg-note-BUG
func isAtomic64Race(functions []*ssa.Function)  {
	// On x86-32, the 64-bit functions use instructions
	// unavailable before the Pentium MMX.
	//arch := runtime.GOARCH
	//
	//if arch == "amd64" {
	//	return
	//}
	// TODO:: we think current system is 32bits
	if len(functions) == 0 {
		return
	}

	for _, f := range functions {
		for _, bb := range f.Blocks {
			for _, ii := range bb.Instrs {
				iiStr := ii.String()
				if constainsAtomic64(iiStr){
					fmt.Println("In 32bits OS, it may be not safe using Atomic Uint64.")
					fmt.Println("There is potential data race in", ii.Parent().Name(), "function.")
					fmt.Println(ii.Parent().Prog.Fset.File(ii.Pos()).Line(ii.Pos()))
				}
			}
		}
	}
}

type members []ssa.Member

// collect all functions in main package
func collectAnonyFuncs(f *ssa.Function) []*ssa.Function {
	if len(f.AnonFuncs) == 0 {
		return nil
	}

	var allAnonF [] *ssa.Function
	for _, anonf := range f.AnonFuncs {
		allAnonF = append(allAnonF, anonf)
		allAnonF = append(allAnonF, collectAnonyFuncs(anonf)...)
	}

	return allAnonF
}

// toSSA converts go source to SSA
func toSSA(source string, fileName, packageName string, debug bool) ([]byte, error) {
	// adopted from saa package example
	conf := loader.Config{
		Build: &build.Default,
	}

	file, err := conf.ParseFile(fileName, source)
	if err != nil {
		return nil, err
	}

	conf.CreateFromFiles("main.go", file)

	prog, err := conf.Load()
	if err != nil {
		return nil, err
	}

	ssaProg := ssautil.CreateProgram(prog, ssa.NaiveForm|ssa.BuildSerially)
	//ssaProg.Build()

	mainPkg := ssaProg.Package(prog.InitialPackages()[0].Pkg)
	out := new(bytes.Buffer)
	mainPkg.SetDebugMode(debug)
	mainPkg.WriteTo(out)
	mainPkg.Build()

	// grab just the functions
	funcs := members([]ssa.Member{})

	// all functions include anony functions
	var allFuncs []*ssa.Function

	for _, obj := range mainPkg.Members {
		if obj.Token() == token.FUNC {
			funcs = append(funcs, obj)
		}
	}

	for _, f := range funcs {
		mainPkg.Func(f.Name()).WriteTo(out)
		currentF := mainPkg.Func(f.Name())
		allFuncs = append(allFuncs, currentF)
		allFuncs = append(allFuncs, collectAnonyFuncs(currentF)...)
	}

	//
	isAtomic64Race(allFuncs)

	return out.Bytes(), nil
}

// collect all go file
func visitDir(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
		}

		if filepath.Ext(path) == ".go" {
			*files = append(*files, path)
		}
		return nil
	}
}


func main() {
	var files []string
	currentDir := "/home/kevin/GoStudy/src/bugs/datarace"
	err := filepath.Walk(currentDir, visitDir(&files))
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		fmt.Println("current file is ", file)
		fmt.Println("-------------------------------")
		dat, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		source := string(dat)
		ssa_out, err := toSSA(source, "main.go", "main", false)
		if true {
			fmt.Println(string(ssa_out))
		}
		fmt.Println("-------------------------------")
	}
}
