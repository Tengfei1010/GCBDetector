package util

import (
	"github.com/Tengfei1010/GCBDetector/ssa"
)

func InstrDominates(InstrA ssa.Instruction, InstrB ssa.Instruction)  bool {

	aBlock := InstrA.Block()
	bBlock := InstrB.Block()

	if aBlock == bBlock {

		//if aBlock.Dominates(aBlock) {
		//	return true
		//}
		// TODO: if aBlock in a loop !!!

		//if
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
