package checkerutil

import (
	"fmt"
	"github.com/Tengfei1010/GCBDetector/ssa"
	"go/token"
)

// check if called by go
// match freevars and bindings
// check referrers of freevars and bindings

func GetInstMakeClosure(anonFunc *ssa.Function) (instMakeClosure *ssa.MakeClosure, ok bool) {

	for _, referrer := range *anonFunc.Referrers() {

		if instMakeClosure, ok = referrer.(*ssa.MakeClosure); ok {
			return instMakeClosure, ok
		}
	}

	return nil, false
}

// TODO: many async call other than go
func GetInstGo(instMakeClosure *ssa.MakeClosure) (instGo *ssa.Go, ok bool) {

	for _, referrer := range *instMakeClosure.Referrers() {

		if instGo, ok = referrer.(*ssa.Go); ok {
			return instGo, ok
		}
	}

	return nil, false
}

type SharedVarReferrer struct {
	LoadInsts []*ssa.UnOp
	StoreInsts []*ssa.Store
}

func GetBindingLoadStore(bindings []ssa.Value) map[ssa.Value] SharedVarReferrer {

	result := make(map[ssa.Value] SharedVarReferrer)

	for _, binding := range bindings {

		loadInsts := make([]*ssa.UnOp, 0)
		storeInsts := make([]*ssa.Store, 0)

		for _, referrer := range *binding.Referrers() {

			if instUnOp, ok := referrer.(*ssa.UnOp); ok {
				if instUnOp.Op == token.MUL {
					loadInsts = append(loadInsts, instUnOp)
				}
			} else if instStore, ok := referrer.(*ssa.Store); ok {
				storeInsts = append(storeInsts, instStore)
			}
		}

		sharedVarReferrer := SharedVarReferrer{LoadInsts:loadInsts, StoreInsts:storeInsts}

		result[binding] = sharedVarReferrer
	}

	return result
}

func GetFreeVarLoadStore(freeVars []*ssa.FreeVar) map[*ssa.FreeVar] SharedVarReferrer {

	result := make(map[*ssa.FreeVar] SharedVarReferrer)

	for _, freeVar := range freeVars {

		loadInsts := make([]*ssa.UnOp, 0)
		storeInsts := make([]*ssa.Store, 0)

		for _, referrer := range *freeVar.Referrers() {

			if instUnOp, ok := referrer.(*ssa.UnOp); ok {
				if instUnOp.Op == token.MUL {
					loadInsts = append(loadInsts, instUnOp)
				}
			} else if instStore, ok := referrer.(*ssa.Store); ok {
				storeInsts = append(storeInsts, instStore)
			}
		}

		sharedVarReferrer := SharedVarReferrer{LoadInsts:loadInsts, StoreInsts:storeInsts}

		result[freeVar] = sharedVarReferrer
	}

	return result
}

func MapBindings2FreeVars(bindings []ssa.Value, freeVars []*ssa.FreeVar) map[ssa.Value] *ssa.FreeVar {

	bindings2FreeVars := make(map[ssa.Value] *ssa.FreeVar)

	if len(freeVars) != len(bindings) {
		fmt.Errorf("len(anonFunc.FreeVars) != len(makeClosure.Bindings)")
		return nil
	}

	for i := 0; i < len(freeVars); i++ {
		bindings2FreeVars[bindings[i]] = freeVars[i]
	}

	return bindings2FreeVars
}

type ResultInfo struct {
	AFreeVar *ssa.FreeVar
	FreeVarHasLoadStore int
	BindingHasLoadStoreAfterGo int
}

func HasBindingLoadStoreAfterGo(bindingLoadStore map[ssa.Value] SharedVarReferrer, instGo *ssa.Go, blockReachability BlockReachability) map[ssa.Value] int {
	// 0: Nope; 1: Load; 2: Store; 3: Load&Store
	result := make(map[ssa.Value] int)

	for binding, loadAndStore := range bindingLoadStore {

		for _, instLoad := range loadAndStore.LoadInsts {

			if loadAfterGo, ok := IsPotentiallyReachableInst(instGo, instLoad, blockReachability); ok {
				//fmt.Println("LOAD", instGo, instLoad, loadAfterGo)
				if loadAfterGo {
					result[binding] = 1
					break
				}
			}
		}

		for _, instStore := range loadAndStore.StoreInsts {
			if storeAfterGo, ok := IsPotentiallyReachableInst(instGo, instStore, blockReachability); ok {
				//fmt.Println("STORE", instGo, instStore, storeAfterGo)
				if storeAfterGo {
					result[binding] += 2
					break
				}
			}
		}
	}

	return result
}

func HasFreeVarLoadStore(freeVarLoadStore map[*ssa.FreeVar] SharedVarReferrer) map[*ssa.FreeVar] int {
	// 0: Nope; 1: Load; 2: Store; 3: Load&Store
	result := make(map[*ssa.FreeVar] int)

	for freeVar, loadAndStore := range freeVarLoadStore {

		if len(loadAndStore.LoadInsts) > 0 {
			result[freeVar] = 1
		}

		if len(loadAndStore.StoreInsts) > 0 {
			result[freeVar] += 2
		}
	}

	return result
}

func GetLoadStoreInfo(anonFuncs []*ssa.Function, blockReachability BlockReachability) map[ssa.Value] map[*ssa.Function] ResultInfo {

	// main Read after go Store
	// main Store after go Read
	// go1 go2 either Stores

	results := make(map[ssa.Value] map[*ssa.Function] ResultInfo)

	for _, anonFunc := range anonFuncs {

		if instMakeClosure, ok := GetInstMakeClosure(anonFunc); ok {
			if instGo, ok := GetInstGo(instMakeClosure); ok {

				binding2FreeVars := MapBindings2FreeVars(instMakeClosure.Bindings, anonFunc.FreeVars)

				bindingLoadStore := GetBindingLoadStore(instMakeClosure.Bindings)
				freeVarLoadStore := GetFreeVarLoadStore(anonFunc.FreeVars)
				hbls := HasBindingLoadStoreAfterGo(bindingLoadStore, instGo, blockReachability)
				hfvls := HasFreeVarLoadStore(freeVarLoadStore)

				for binding, bindingLoadStore := range hbls {

					if _, ok := results[binding]; !ok {
						results[binding] = make(map[*ssa.Function] ResultInfo)
					}

					freeVar := binding2FreeVars[binding]
					freeVarLoadStore := hfvls[freeVar]

					resultInfo := ResultInfo{AFreeVar: freeVar, FreeVarHasLoadStore: freeVarLoadStore, BindingHasLoadStoreAfterGo: bindingLoadStore}
					results[binding][anonFunc] = resultInfo
				}
			}
		}
	}

	return results
}

func HasAnonRace(anonFuncs []*ssa.Function, blockReachability BlockReachability) (map[ssa.Value] map[*ssa.Function] ResultInfo, bool) {

	raceVars := make(map[ssa.Value] map[*ssa.Function] ResultInfo)

	hasAnonRace := false

	loadStoreInfo := GetLoadStoreInfo(anonFuncs, blockReachability)

	for binding, bindingInfo := range loadStoreInfo {

		// More than 1 go routine shares a var and at least one stores it
		if len(bindingInfo) > 1 {

			for _, funcInfo := range bindingInfo {

				if funcInfo.FreeVarHasLoadStore >= 2 {

					raceVars[binding] = bindingInfo

					hasAnonRace = true
				}
			}
		}

		for _, funcInfo := range bindingInfo {

			// As long as there is one Store in either Main or Go routine
			if funcInfo.BindingHasLoadStoreAfterGo >= 2 || funcInfo.FreeVarHasLoadStore >= 2 {

				raceVars[binding] = bindingInfo

				hasAnonRace = true
			}
		}
	}

	return loadStoreInfo, hasAnonRace
}



