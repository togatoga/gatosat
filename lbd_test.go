package main

import (
	"testing"
)

func TestComputeLBD(t *testing.T) {
	//LBD 2
	lits := []Lit{*NewLit(0, false), *NewLit(1, true), *NewLit(2, true)} // (!x1 v x2 v x3)
	varData := []VarData{*NewVarData(ClaRefUndef, 1), *NewVarData(ClaRefUndef, 1), *NewVarData(ClaRefUndef, 2)}
	solver := NewSolver()
	solver.VarData = varData
	if 2 != solver.ComputeLBD(lits) {
		panic(nil)
	}
}
