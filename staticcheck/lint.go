/*
================================================================
=  Source code from https://github.com/dominikh/go-tools       =
=  Copyright @ Dominik Honnef (https://github.com/dominikh)    =
================================================================
*/

// Package staticcheck contains a linter for Go source code.
package staticcheck

import (
	"fmt"
	"github.com/Tengfei1010/GCBDetector/staticcheck/checkerutil"
	"go/ast"
	"go/token"
	"go/types"
	//"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/Tengfei1010/GCBDetector/functions"
	"github.com/Tengfei1010/GCBDetector/lint"
	. "github.com/Tengfei1010/GCBDetector/lint/lintdsl"
	"github.com/Tengfei1010/GCBDetector/ssa"
	"github.com/deckarep/golang-set"

	"golang.org/x/tools/go/loader"
)

type runeSlice []rune

func (rs runeSlice) Len() int               { return len(rs) }
func (rs runeSlice) Less(i int, j int) bool { return rs[i] < rs[j] }
func (rs runeSlice) Swap(i int, j int)      { rs[i], rs[j] = rs[j], rs[i] }

type Checker struct {
	CheckGenerated bool
	funcDescs      *functions.Descriptions
	deprecatedObjs map[types.Object]string
}

func NewChecker() *Checker {
	return &Checker{}
}

func (*Checker) Name() string   { return "staticcheck" }
func (*Checker) Prefix() string { return "SA" }

func (c *Checker) Funcs() map[string]lint.Func {
	return map[string]lint.Func{

		"SA2000": c.CheckWaitgroupAdd,
		"SA2001": c.CheckEmptyCriticalSection,
		"SA2002": c.CheckConcurrentTesting,
		"SA2003": c.CheckDeferLock,
		"SA2004": c.CheckUnlockAfterLock,
		"SA2005": c.CheckDoubleLock,
		"SA2006": c.CheckAnonRace,
	}
}

func (c *Checker) filterGenerated(files []*ast.File) []*ast.File {
	if c.CheckGenerated {
		return files
	}
	var out []*ast.File
	for _, f := range files {
		if !IsGenerated(f) {
			out = append(out, f)
		}
	}
	return out
}

func (c *Checker) findDeprecated(prog *lint.Program) {
	var docs []*ast.CommentGroup
	var names []*ast.Ident

	doDocs := func(pkginfo *loader.PackageInfo, names []*ast.Ident, docs []*ast.CommentGroup) {
		var alt string
		for _, doc := range docs {
			if doc == nil {
				continue
			}
			parts := strings.Split(doc.Text(), "\n\n")
			last := parts[len(parts)-1]
			if !strings.HasPrefix(last, "Deprecated: ") {
				continue
			}
			alt = last[len("Deprecated: "):]
			alt = strings.Replace(alt, "\n", " ", -1)
			break
		}
		if alt == "" {
			return
		}

		for _, name := range names {
			obj := pkginfo.ObjectOf(name)
			c.deprecatedObjs[obj] = alt
		}
	}

	for _, pkginfo := range prog.Prog.AllPackages {
		for _, f := range pkginfo.Files {
			fn := func(node ast.Node) bool {
				if node == nil {
					return true
				}
				var ret bool
				switch node := node.(type) {
				case *ast.GenDecl:
					switch node.Tok {
					case token.TYPE, token.CONST, token.VAR:
						docs = append(docs, node.Doc)
						return true
					default:
						return false
					}
				case *ast.FuncDecl:
					docs = append(docs, node.Doc)
					names = []*ast.Ident{node.Name}
					ret = false
				case *ast.TypeSpec:
					docs = append(docs, node.Doc)
					names = []*ast.Ident{node.Name}
					ret = true
				case *ast.ValueSpec:
					docs = append(docs, node.Doc)
					names = node.Names
					ret = false
				case *ast.File:
					return true
				case *ast.StructType:
					for _, field := range node.Fields.List {
						doDocs(pkginfo, field.Names, []*ast.CommentGroup{field.Doc})
					}
					return false
				case *ast.InterfaceType:
					for _, field := range node.Methods.List {
						doDocs(pkginfo, field.Names, []*ast.CommentGroup{field.Doc})
					}
					return false
				default:
					return false
				}
				if len(names) == 0 || len(docs) == 0 {
					return ret
				}
				doDocs(pkginfo, names, docs)

				docs = docs[:0]
				names = nil
				return ret
			}
			ast.Inspect(f, fn)
		}
	}
}

func (c *Checker) Init(prog *lint.Program) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		c.funcDescs = functions.NewDescriptions(prog.SSA)
		for _, fn := range prog.AllFunctions {
			if fn.Blocks != nil {
				applyStdlibKnowledge(fn)
				ssa.OptimizeBlocks(fn)
			}
		}
		wg.Done()
	}()

	go func() {
		c.deprecatedObjs = map[types.Object]string{}
		c.findDeprecated(prog)
		wg.Done()
	}()

	wg.Wait()
}

func (c *Checker) isInLoop(b *ssa.BasicBlock) bool {
	sets := c.funcDescs.Get(b.Parent()).Loops
	for _, set := range sets {
		if set[b] {
			return true
		}
	}
	return false
}

func applyStdlibKnowledge(fn *ssa.Function) {
	if len(fn.Blocks) == 0 {
		return
	}

	// comma-ok receiving from a time.Tick channel will never return
	// ok == false, so any branching on the value of ok can be
	// replaced with an unconditional jump. This will primarily match
	// `for range time.Tick(x)` loops, but it can also match
	// user-written code.
	for _, block := range fn.Blocks {
		if len(block.Instrs) < 3 {
			continue
		}
		if len(block.Succs) != 2 {
			continue
		}
		var instrs []*ssa.Instruction
		for i, ins := range block.Instrs {
			if _, ok := ins.(*ssa.DebugRef); ok {
				continue
			}
			instrs = append(instrs, &block.Instrs[i])
		}

		for i, ins := range instrs {
			unop, ok := (*ins).(*ssa.UnOp)
			if !ok || unop.Op != token.ARROW {
				continue
			}
			call, ok := unop.X.(*ssa.Call)
			if !ok {
				continue
			}
			if !IsCallTo(call.Common(), "time.Tick") {
				continue
			}
			ex, ok := (*instrs[i+1]).(*ssa.Extract)
			if !ok || ex.Tuple != unop || ex.Index != 1 {
				continue
			}

			ifstmt, ok := (*instrs[i+2]).(*ssa.If)
			if !ok || ifstmt.Cond != ex {
				continue
			}

			*instrs[i+2] = ssa.NewJump(block)
			succ := block.Succs[1]
			block.Succs = block.Succs[0:1]
			succ.RemovePred(block)
		}
	}
}

func hasType(j *lint.Job, expr ast.Expr, name string) bool {
	T := TypeOf(j, expr)
	return IsType(T, name)
}

func isTestMain(j *lint.Job, node ast.Node) bool {
	decl, ok := node.(*ast.FuncDecl)
	if !ok {
		return false
	}
	if decl.Name.Name != "TestMain" {
		return false
	}
	if len(decl.Type.Params.List) != 1 {
		return false
	}
	arg := decl.Type.Params.List[0]
	if len(arg.Names) != 1 {
		return false
	}
	return IsOfType(j, arg.Type, "*testing.M")
}

func selectorX(sel *ast.SelectorExpr) ast.Node {
	switch x := sel.X.(type) {
	case *ast.SelectorExpr:
		return selectorX(x)
	default:
		return x
	}
}

// cgo produces code like fn(&*_Cvar_kSomeCallbacks) which we don't
// want to flag.
var cgoIdent = regexp.MustCompile(`^_C(func|var)_.+$`)

func consts(val ssa.Value, out []*ssa.Const, visitedPhis map[string]bool) ([]*ssa.Const, bool) {
	if visitedPhis == nil {
		visitedPhis = map[string]bool{}
	}
	var ok bool
	switch val := val.(type) {
	case *ssa.Phi:
		if visitedPhis[val.Name()] {
			break
		}
		visitedPhis[val.Name()] = true
		vals := val.Operands(nil)
		for _, phival := range vals {
			out, ok = consts(*phival, out, visitedPhis)
			if !ok {
				return nil, false
			}
		}
	case *ssa.Const:
		out = append(out, val)
	case *ssa.Convert:
		out, ok = consts(val.X, out, visitedPhis)
		if !ok {
			return nil, false
		}
	default:
		return nil, false
	}
	if len(out) < 2 {
		return out, true
	}
	uniq := []*ssa.Const{out[0]}
	for _, val := range out[1:] {
		if val.Value == uniq[len(uniq)-1].Value {
			continue
		}
		uniq = append(uniq, val)
	}
	return uniq, true
}

func objectName(obj types.Object) string {
	if obj == nil {
		return "<nil>"
	}
	var name string
	if obj.Pkg() != nil && obj.Pkg().Scope().Lookup(obj.Name()) == obj {
		var s string
		s = obj.Pkg().Path()
		if s != "" {
			name += s + "."
		}
	}
	name += obj.Name()
	return name
}

func isName(j *lint.Job, expr ast.Expr, name string) bool {
	var obj types.Object
	switch expr := expr.(type) {
	case *ast.Ident:
		obj = ObjectOf(j, expr)
	case *ast.SelectorExpr:
		obj = ObjectOf(j, expr.Sel)
	}
	return objectName(obj) == name
}

func hasSideEffects(node ast.Node) bool {
	dynamic := false
	ast.Inspect(node, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.CallExpr:
			dynamic = true
			return false
		case *ast.UnaryExpr:
			if node.Op == token.ARROW {
				dynamic = true
				return false
			}
		}
		return true
	})
	return dynamic
}

func unwrapFunction(val ssa.Value) *ssa.Function {
	switch val := val.(type) {
	case *ssa.Function:
		return val
	case *ssa.MakeClosure:
		return val.Fn.(*ssa.Function)
	default:
		return nil
	}
}

func shortCallName(call *ssa.CallCommon) string {
	if call.IsInvoke() {
		return ""
	}
	switch v := call.Value.(type) {
	case *ssa.Function:
		fn, ok := v.Object().(*types.Func)
		if !ok {
			return ""
		}
		return fn.Name()
	case *ssa.Builtin:
		return v.Name()
	}
	return ""
}

func hasCallTo(block *ssa.BasicBlock, name string) bool {
	for _, ins := range block.Instrs {
		call, ok := ins.(*ssa.Call)
		if !ok {
			continue
		}
		if IsCallTo(call.Common(), name) {
			return true
		}
	}
	return false
}

func loopedRegexp(name string) CallCheck {
	return func(call *Call) {
		if len(extractConsts(call.Args[0].Value.Value)) == 0 {
			return
		}
		if !call.Checker.isInLoop(call.Instr.Block()) {
			return
		}
		call.Invalid(fmt.Sprintf("calling %s in a loop has poor performance, consider using regexp.Compile", name))
	}
}

func buildTagsIdentical(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	s1s := make([]string, len(s1))
	copy(s1s, s1)
	sort.Strings(s1s)
	s2s := make([]string, len(s2))
	copy(s2s, s2)
	sort.Strings(s2s)
	for i, s := range s1s {
		if s != s2s[i] {
			return false
		}
	}
	return true
}

func isCallToLock(callCommon *ssa.CallCommon) bool {

	if IsCallTo(callCommon, "(*sync.Mutex).Lock") ||
		IsCallTo(callCommon, "(*sync.RWMutex).RLock") ||
		IsCallTo(callCommon, "(*sync.RWMutex).Lock") {
			return true
	}

	return false
}

func  isCallToUnlock(callCommon *ssa.CallCommon) bool {
	if IsCallTo(callCommon, "(*sync.Mutex).Unlock") ||
		IsCallTo(callCommon, "(*sync.RWMutex).RUnlock") ||
		IsCallTo(callCommon, "(*sync.RWMutex).UnLock") {
		return true
	}

	return false

}

func (c *Checker) CheckWaitgroupAdd(j *lint.Job) {
	fn := func(node ast.Node) bool {
		g, ok := node.(*ast.GoStmt)
		if !ok {
			return true
		}
		fun, ok := g.Call.Fun.(*ast.FuncLit)
		if !ok {
			return true
		}
		if len(fun.Body.List) == 0 {
			return true
		}
		stmt, ok := fun.Body.List[0].(*ast.ExprStmt)
		if !ok {
			return true
		}
		call, ok := stmt.X.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		fn, ok := ObjectOf(j, sel.Sel).(*types.Func)
		if !ok {
			return true
		}
		if fn.FullName() == "(*sync.WaitGroup).Add" {
			j.Errorf(sel, "should call %s before starting the goroutine to avoid a race",
				Render(j, stmt))
		}
		return true
	}
	for _, f := range j.Program.Files {
		ast.Inspect(f, fn)
	}
}

func (c *Checker) CheckEmptyCriticalSection(j *lint.Job) {
	// Initially it might seem like this check would be easier to
	// implement in SSA. After all, we're only checking for two
	// consecutive method calls. In reality, however, there may be any
	// number of other instructions between the lock and unlock, while
	// still constituting an empty critical section. For example,
	// given `m.x().Lock(); m.x().Unlock()`, there will be a call to
	// x(). In the AST-based approach, this has a tiny potential for a
	// false positive (the second call to x might be doing work that
	// is protected by the mutex). In an SSA-based approach, however,
	// it would miss a lot of real bugs.

	mutexParams := func(s ast.Stmt) (x ast.Expr, funcName string, ok bool) {
		expr, ok := s.(*ast.ExprStmt)
		if !ok {
			return nil, "", false
		}
		call, ok := expr.X.(*ast.CallExpr)
		if !ok {
			return nil, "", false
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return nil, "", false
		}

		fn, ok := ObjectOf(j, sel.Sel).(*types.Func)
		if !ok {
			return nil, "", false
		}
		sig := fn.Type().(*types.Signature)
		if sig.Params().Len() != 0 || sig.Results().Len() != 0 {
			return nil, "", false
		}

		return sel.X, fn.Name(), true
	}

	fn := func(node ast.Node) bool {
		block, ok := node.(*ast.BlockStmt)
		if !ok {
			return true
		}
		if len(block.List) < 2 {
			return true
		}
		for i := range block.List[:len(block.List)-1] {
			sel1, method1, ok1 := mutexParams(block.List[i])
			sel2, method2, ok2 := mutexParams(block.List[i+1])

			if !ok1 || !ok2 || Render(j, sel1) != Render(j, sel2) {
				continue
			}
			if (method1 == "Lock" && method2 == "Unlock") ||
				(method1 == "RLock" && method2 == "RUnlock") {
				j.Errorf(block.List[i+1], "empty critical section")
			}
		}
		return true
	}
	for _, f := range j.Program.Files {
		ast.Inspect(f, fn)
	}
}

func (c *Checker) CheckConcurrentTesting(j *lint.Job) {
	for _, ssafn := range j.Program.InitialFunctions {
		for _, block := range ssafn.Blocks {
			for _, ins := range block.Instrs {
				gostmt, ok := ins.(*ssa.Go)
				if !ok {
					continue
				}
				var fn *ssa.Function
				switch val := gostmt.Call.Value.(type) {
				case *ssa.Function:
					fn = val
				case *ssa.MakeClosure:
					fn = val.Fn.(*ssa.Function)
				default:
					continue
				}
				if fn.Blocks == nil {
					continue
				}
				for _, block := range fn.Blocks {
					for _, ins := range block.Instrs {
						call, ok := ins.(*ssa.Call)
						if !ok {
							continue
						}
						if call.Call.IsInvoke() {
							continue
						}
						callee := call.Call.StaticCallee()
						if callee == nil {
							continue
						}
						recv := callee.Signature.Recv()
						if recv == nil {
							continue
						}
						if !IsType(recv.Type(), "*testing.common") {
							continue
						}
						fn, ok := call.Call.StaticCallee().Object().(*types.Func)
						if !ok {
							continue
						}
						name := fn.Name()
						switch name {
						case "FailNow", "Fatal", "Fatalf", "SkipNow", "Skip", "Skipf":
						default:
							continue
						}
						j.Errorf(gostmt, "the goroutine calls T.%s, which must be called in the same goroutine as the test", name)
					}
				}
			}
		}
	}
}

func (c *Checker) CheckDeferLock(j *lint.Job) {

	for _, ssafn := range j.Program.InitialFunctions {
		for _, block := range ssafn.Blocks {
			instrs := FilterDebug(block.Instrs)
			if len(instrs) < 2 {
				continue
			}
			for i, ins := range instrs[:len(instrs)-1] {
				call, ok := ins.(*ssa.Call)
				if !ok {
					continue
				}
				if !isCallToLock(call.Common()) {
					continue
				}

				nins, ok := instrs[i+1].(*ssa.Defer)
				if !ok {
					continue
				}
				if !isCallToLock(nins.Common()) {
					continue
				}
				if call.Common().Args[0] != nins.Call.Args[0] {
					continue
				}
				name := shortCallName(call.Common())
				alt := ""
				switch name {
				case "Lock":
					alt = "Unlock"
				case "RLock":
					alt = "RUnlock"
				}
				j.Errorf(nins, "deferring %s right after having locked already; did you mean to defer %s?", name, alt)
			}
		}
	}
}

func (c *Checker) CheckUnlockAfterLock(j *lint.Job) {

	for _, ssafn := range j.Program.InitialFunctions {
		for _, block := range ssafn.Blocks {

			instrs := FilterDebug(block.Instrs)

			if len(instrs) < 2 {
				continue
			}

			for i, ins := range instrs[:len(instrs)-1] {
				call, ok := ins.(*ssa.Call)
				if !ok {
					continue
				}
				if !isCallToLock(call.Common()) {
					continue
				}
				nins, ok := instrs[i+1].(*ssa.Call)
				if !ok {
					continue
				}
				if !isCallToUnlock(nins.Common()) {
					continue
				}
				if call.Common().Args[0] != nins.Call.Args[0] {
					continue
				}
				name := shortCallName(call.Common())
				alt := ""
				switch name {
				case "Lock":
					alt = "Unlock"
				case "RLock":
					alt = "RUnlock"
				}
				j.Errorf(nins, "Unlock %s right after locking; did you mean to defer %s?", name, alt)
			}
		}
	}
}

func (c *Checker) CheckDoubleLock(j *lint.Job) {

	for _, ssafn := range j.Program.InitialFunctions {
		lockSet := mapset.NewSet()
		for _, block := range ssafn.Blocks {
			instrs := FilterDebug(block.Instrs)

			if len(instrs) < 2 {
				continue
			}

			for _, ins := range instrs[:len(instrs)-1] {

				call, ok := ins.(*ssa.Call)

				if !ok {
					continue
				}

				if isCallToLock(call.Common()) {
					// if call is lock, save to the set; else if call is unlock remove the lock from the set
					ok = lockSet.Add(call.Common().Args[0])
					if !ok {
						// add error, it has already acquired the lock
						// TODO: update error message
						name := shortCallName(call.Common())
						j.Errorf(call, "Acquiring %s right after having locked already", name)
					}
				} else if isCallToUnlock(call.Common()) {

					lockSet.Remove(call.Common().Args[0])

				} else {
					continue
				}
			}
		}
	}
}

func (c *Checker) CheckAnonRace(j *lint.Job) {

	for _, ssafn := range j.Program.InitialFunctions {

		if strings.HasSuffix(ssafn.String(), ".init") {
			continue
		}

		blockReachability := checkerutil.MapReachableBlocks(ssafn)

		if result, ok := checkerutil.HasAnonRace(ssafn.AnonFuncs, blockReachability); ok {
			fmt.Println(result)
		}
	}

}
