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
