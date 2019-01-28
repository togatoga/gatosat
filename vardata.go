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

//Reason returns a ClauseReference for
func (s *Solver) Reason(x Var) ClauseReference {
	return s.VarData[x].Reason
}

//Level returns a decision level for a Var
func (s *Solver) Level(x Var) int {
	return s.VarData[x].Level
}
