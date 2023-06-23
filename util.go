package echoAnalyzer

import (
	"golang.org/x/tools/go/ssa"
)

func eachInstruction(funcs []*ssa.Function, handler func(instr ssa.Instruction)) {
	for _, fn := range funcs {
		for _, block := range fn.Blocks {
			for _, instr := range block.Instrs {
				handler(instr)
			}
		}
	}
}

func isInterfaceMethodCall(instr ssa.Instruction) bool {
	call, isCall := instr.(*ssa.Call)
	if !isCall {
		return false
	}

	return call.Common().IsInvoke()
}

func isStructMethodCall(instr ssa.Instruction) bool {
	call, isCall := instr.(*ssa.Call)
	if !isCall {
		return false
	}

	if call.Common().IsInvoke() {
    return false
  }

  return call.Common().Signature().Recv() != nil
}
