package util

import (
	"github.com/Tengfei1010/GCBDetector/ssa"
	"fmt"
)

func InstrDominas(InstrA ssa.Instruction, InstrB ssa.Instruction)  bool {

	var aBlock *ssa.BasicBlock
	var bBlock *ssa.BasicBlock

	switch InstrA.(type) {

	case *ssa.Alloc:
		aBlock = InstrA.(*ssa.Alloc).Block()
	case *ssa.BinOp:
		aBlock = InstrA.(*ssa.Go).Block()
	case *ssa.Call:
		aBlock = InstrA.(*ssa.Call).Block()
	case *ssa.ChangeInterface:
		aBlock = InstrA.(*ssa.ChangeInterface).Block()
	case *ssa.ChangeType:
		aBlock = InstrA.(*ssa.ChangeType).Block()
	case *ssa.Convert:
		aBlock = InstrA.(*ssa.Convert).Block()
	case *ssa.DebugRef:
		aBlock = InstrA.(*ssa.DebugRef).Block()
	case *ssa.Defer:
		aBlock = InstrA.(*ssa.Defer).Block()
	case *ssa.Extract:
		aBlock = InstrA.(*ssa.Extract).Block()
	case *ssa.Field:
		aBlock = InstrA.(*ssa.Field).Block()
	case *ssa.FieldAddr:
		aBlock = InstrA.(*ssa.FieldAddr).Block()
	case *ssa.Go:
		aBlock = InstrA.(*ssa.Go).Block()
	case *ssa.If:
		aBlock = InstrA.(*ssa.If).Block()
	case *ssa.Index:
		aBlock = InstrA.(*ssa.Index).Block()
	case *ssa.IndexAddr:
		aBlock = InstrA.(*ssa.IndexAddr).Block()
	case *ssa.Jump:
		aBlock = InstrA.(*ssa.Jump).Block()
	case *ssa.Lookup:
		aBlock = InstrA.(*ssa.Lookup).Block()
	case *ssa.MakeChan:
		aBlock = InstrA.(*ssa.MakeChan).Block()
	case *ssa.MakeClosure:
		aBlock = InstrA.(*ssa.MakeClosure).Block()
	case *ssa.MakeInterface:
		aBlock = InstrA.(*ssa.MakeInterface).Block()
	case *ssa.MakeMap:
		aBlock = InstrA.(*ssa.MakeMap).Block()
	case *ssa.MakeSlice:
		aBlock = InstrA.(*ssa.MakeSlice).Block()
	case *ssa.MapUpdate:
		aBlock = InstrA.(*ssa.MapUpdate).Block()
	case *ssa.Next:
		aBlock = InstrA.(*ssa.Next).Block()
	case *ssa.Panic:
		aBlock = InstrA.(*ssa.Panic).Block()
	case *ssa.Phi:
		aBlock = InstrA.(*ssa.Phi).Block()
	case *ssa.Range:
		aBlock = InstrA.(*ssa.Range).Block()
	case *ssa.Return:
		aBlock = InstrA.(*ssa.Return).Block()
	case *ssa.RunDefers:
		aBlock = InstrA.(*ssa.RunDefers).Block()
	case *ssa.Select:
		aBlock = InstrA.(*ssa.Select).Block()
	case *ssa.Send:
		aBlock = InstrA.(*ssa.Send).Block()
	case *ssa.Slice:
		aBlock = InstrA.(*ssa.Slice).Block()
	case *ssa.Store:
		aBlock = InstrA.(*ssa.Store).Block()
	case *ssa.TypeAssert:
		aBlock = InstrA.(*ssa.TypeAssert).Block()
	case *ssa.UnOp:
		aBlock = InstrA.(*ssa.UnOp).Block()

	default:
		fmt.Println("error........")
	}


	switch InstrB.(type) {

	case *ssa.Alloc:
		bBlock = InstrB.(*ssa.Alloc).Block()
	case *ssa.BinOp:
		bBlock = InstrB.(*ssa.Go).Block()
	case *ssa.Call:
		bBlock = InstrB.(*ssa.Call).Block()
	case *ssa.ChangeInterface:
		bBlock = InstrB.(*ssa.ChangeInterface).Block()
	case *ssa.ChangeType:
		bBlock = InstrB.(*ssa.ChangeType).Block()
	case *ssa.Convert:
		bBlock = InstrB.(*ssa.Convert).Block()
	case *ssa.DebugRef:
		bBlock = InstrB.(*ssa.DebugRef).Block()
	case *ssa.Defer:
		bBlock = InstrB.(*ssa.Defer).Block()
	case *ssa.Extract:
		bBlock = InstrB.(*ssa.Extract).Block()
	case *ssa.Field:
		bBlock = InstrB.(*ssa.Field).Block()
	case *ssa.FieldAddr:
		bBlock = InstrB.(*ssa.FieldAddr).Block()
	case *ssa.Go:
		bBlock = InstrB.(*ssa.Go).Block()
	case *ssa.If:
		bBlock = InstrB.(*ssa.If).Block()
	case *ssa.Index:
		bBlock = InstrB.(*ssa.Index).Block()
	case *ssa.IndexAddr:
		bBlock = InstrB.(*ssa.IndexAddr).Block()
	case *ssa.Jump:
		bBlock = InstrB.(*ssa.Jump).Block()
	case *ssa.Lookup:
		bBlock = InstrB.(*ssa.Lookup).Block()
	case *ssa.MakeChan:
		bBlock = InstrB.(*ssa.MakeChan).Block()
	case *ssa.MakeClosure:
		bBlock = InstrB.(*ssa.MakeClosure).Block()
	case *ssa.MakeInterface:
		bBlock = InstrB.(*ssa.MakeInterface).Block()
	case *ssa.MakeMap:
		bBlock = InstrB.(*ssa.MakeMap).Block()
	case *ssa.MakeSlice:
		bBlock = InstrB.(*ssa.MakeSlice).Block()
	case *ssa.MapUpdate:
		bBlock = InstrB.(*ssa.MapUpdate).Block()
	case *ssa.Next:
		bBlock = InstrB.(*ssa.Next).Block()
	case *ssa.Panic:
		bBlock = InstrB.(*ssa.Panic).Block()
	case *ssa.Phi:
		bBlock = InstrB.(*ssa.Phi).Block()
	case *ssa.Range:
		bBlock = InstrB.(*ssa.Range).Block()
	case *ssa.Return:
		bBlock = InstrB.(*ssa.Return).Block()
	case *ssa.RunDefers:
		bBlock = InstrB.(*ssa.RunDefers).Block()
	case *ssa.Select:
		bBlock = InstrB.(*ssa.Select).Block()
	case *ssa.Send:
		bBlock = InstrB.(*ssa.Send).Block()
	case *ssa.Slice:
		bBlock = InstrB.(*ssa.Slice).Block()
	case *ssa.Store:
		bBlock = InstrB.(*ssa.Store).Block()
	case *ssa.TypeAssert:
		bBlock = InstrB.(*ssa.TypeAssert).Block()
	case *ssa.UnOp:
		bBlock = InstrB.(*ssa.UnOp).Block()

	default:
		fmt.Println("error........")
	}

	if aBlock == bBlock {

		//if aBlock.Dominates(aBlock) {
		//	return true
		//}
		// TODO: if aBlock in a loop !!!

		for _, instr := range aBlock.Instrs {

			if instr.Pos() == InstrA.Pos() {
				return true
			}

			if instr.Pos() == InstrB.Pos() {
				return false
			}
		}

	}

	return aBlock.Dominates(bBlock)

}
