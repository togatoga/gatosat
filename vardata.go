package main

type VarData struct {
	Reason ClauseReference
	Level  int
}

func NewVarData(claRef ClauseReference, level int) *VarData {
	return &VarData{
		Reason: claRef,
		Level:  level,
	}
}

func (s *Solver) Reason(x Var) ClauseReference {
	return s.VarData[x].Reason
}

func (s *Solver) Level(x Var) int {
	return s.VarData[x].Level
}
