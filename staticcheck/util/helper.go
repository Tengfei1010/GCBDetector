package util

import (
	"github.com/Tengfei1010/GCBDetector/ssa"
)

func Contains(visitedBlocks[]*ssa.BasicBlock, block *ssa.BasicBlock) bool{
	for _, b := range visitedBlocks {
		if block == b {
			return true
		}
	}
	return false
}

func InstrIndexInBlock(instr ssa.Instruction) int{
	aBlock := instr.Block()

	for i, ins := range aBlock.Instrs {
		if ins == instr {
			return i
		}
	}
	return 0
}

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
