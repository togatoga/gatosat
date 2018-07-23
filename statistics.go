package main

type Statistics struct {
	RestartCount       uint64
	DecisionCount      uint64
	PropagationCount   uint64
	ConflictCount      uint64
	NumLearnts         uint64
	NumUnitLearnts     uint64
	NumBinaryLearnts   uint64
	NumClauses         uint64
	ReduceDBCount      uint64
	RemovedClauseCount uint64
}

func NewStatistics() *Statistics {
	return &Statistics{
		RestartCount:       0,
		DecisionCount:      0,
		PropagationCount:   0,
		ConflictCount:      0,
		NumLearnts:         0,
		NumUnitLearnts:     0,
		NumBinaryLearnts:   0,
		NumClauses:         0,
		ReduceDBCount:      0,
		RemovedClauseCount: 0,
	}
}

func (s *Solver) NumClauses() uint64 {
	return s.Statistics.NumClauses
}
