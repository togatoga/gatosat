package main

type Solver struct {
	Verbosity    bool
	ClaAllocator *ClauseAllocator
	Clauses      map[ClauseReference]bool
}

func NewSolver() *Solver {
	solver := &Solver{
		Verbosity:    false,
		ClaAllocator: NewClauseAllocator(),
		Clauses:      make(map[ClauseReference]bool),
	}
	return solver
}

func (s *Solver) addClause(lits []Lit) {
	claRef, err := s.ClaAllocator.NewAllocate(lits, false)
	if err != nil {
		panic(err)
	}
	s.Clauses[claRef] = true
}
