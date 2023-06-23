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
