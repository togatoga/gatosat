package main

type Statistics struct {
	RestartCount     uint64
	DecisionCount    uint64
	PropagationCount uint64
	ConflictCount    uint64
}

func NewStatistics() *Statistics {
	return &Statistics{
		RestartCount:     0,
		DecisionCount:    0,
		PropagationCount: 0,
		ConflictCount:    0,
	}
}
