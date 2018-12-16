package checkerutil

import (
	"github.com/Tengfei1010/GCBDetector/ssa"
)


func IsPotentiallyReachableInst(lhs ssa.Instruction, rhs ssa.Instruction, blockReachability BlockReachability) (bool, ok bool) {

	if lhs.Block() == rhs.Block() {

		indexLhs := -1
		indexRhs := -1

		for index, II := range lhs.Block().Instrs {

			if II == lhs {
				indexLhs = index
			} else if II == rhs {
				indexRhs = index
			}

			if indexLhs != -1 && indexRhs != -1 {
				return indexLhs < indexRhs, true
			}
		}
		return false, false

	} else {
		return IsPotentiallyReachableBlocks(lhs.Block(), rhs.Block(), blockReachability), true
	}
}

func IsPotentiallyReachableBlocks(lhs *ssa.BasicBlock, rhs *ssa.BasicBlock, blockReachability BlockReachability) bool {
	return blockReachability.Reachability[blockReachability.BlockNum * lhs.Index + rhs.Index]
}

type BlockReachability struct {
	Reachability []bool
	BlockNum int
}

func MapReachableBlocks(aFunc *ssa.Function) BlockReachability {

	blockNum := len(aFunc.Blocks)

	reachable := make([]bool, blockNum*blockNum)

	for _, BB := range aFunc.Blocks {

		for _, succ := range BB.Succs {
			reachable[blockNum* BB.Index + succ.Index] = true
		}
	}

	for k := 0; k < blockNum; k++ {

		for i := 0; i < blockNum; i++ {

			for j := 0; j < blockNum; j++ {

				if reachable[blockNum* i + j] == false {
					if reachable[blockNum*i+k] == true && reachable[blockNum*k+j] == true {
						reachable[blockNum*i+j] = true
					}
				}
			}
		}

	}

	return BlockReachability{Reachability: reachable, BlockNum: blockNum}
}
