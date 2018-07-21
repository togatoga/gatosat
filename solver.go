package main

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/k0kubun/pp"
)

type Solver struct {
	Verbosity        bool
	ClaAllocator     *ClauseAllocator         //The allocator for clause
	Clauses          map[ClauseReference]bool //List of problem clauses.
	LearntClauses    map[ClauseReference]bool //List of learnt clauses.
	Watches          map[Lit][]*Watcher       //'watches[lit]' is a list of constraints watching 'lit' (will go there if literal becomes true).
	Assigns          []LitBool                //The current assignments.
	Qhead            int                      //Head of queue (as index into the trail -- no more explicit propagation queue in MiniSat).
	Trail            []Lit                    //Assignment stack; stores all assigments made in the order the were made.
	TrailLim         []int                    //Separator indices for different decision levels in 'trail'.
	NextVar          Var                      //Next variable to be created.
	Decision         []bool                   // A priority queue of variables ordered with respect to the variable activity.
	VarData          []VarData                //Stores reason and level for each variable.
	VarOrder         *Heap                    // A priority queue of variables ordered with respect to the variable activity.
	OK               bool                     //If FALSE, the constraints are already unsatisfiable. No part of the solver state may be used!
	VarIncreaseRatio float64                  // Amount to bump next variable with.
	VarDecayRatio    float64
	Seen             []bool      //The seen variable for clause learning
	Model            []LitBool   // If problem is satisfiable, this vector contains the model (if any).
	Statistics       *Statistics //Statistics
}

func NewSolver(c *cli.Context) *Solver {
	return &Solver{
		Verbosity:        c.Bool("verbosity"),
		ClaAllocator:     NewClauseAllocator(),
		Clauses:          make(map[ClauseReference]bool),
		LearntClauses:    make(map[ClauseReference]bool),
		Watches:          make(map[Lit][]*Watcher),
		Qhead:            0,
		NextVar:          0,
		VarOrder:         NewHeap(),
		OK:               true,
		VarIncreaseRatio: 1.0,
		VarDecayRatio:    0.95,
		Statistics:       NewStatistics(),
	}
}

func (s *Solver) NewVar() Var {
	v := s.NextVar
	s.NextVar++
	s.Assigns = append(s.Assigns, LitBoolUndef)
	s.VarData = append(s.VarData, *NewVarData(ClaRefUndef, 0))
	s.Seen = append(s.Seen, false)
	s.Decision = append(s.Decision, true)
	s.SetDecisionVar(v, true)
	return v
}

func (s *Solver) ValueVar(p Var) LitBool {
	return s.Assigns[p]
}

func (s *Solver) ValueLit(p Lit) LitBool {
	if s.Assigns[p.Var()] == LitBoolUndef {
		return LitBoolUndef
	} else if s.Assigns[p.Var()] == LitBoolTrue {
		if !p.Sign() {
			return LitBoolTrue
		}
	} else if s.Assigns[p.Var()] == LitBoolFalse {
		if p.Sign() {
			return LitBoolTrue
		}
	}
	return LitBoolFalse
}

func (s *Solver) varDecayActivity() {
	s.VarIncreaseRatio *= (1 / s.VarDecayRatio)
}

func (s *Solver) varBumpActitivy(v Var) {
	s.varBumpActitivyByInc(v, s.VarIncreaseRatio)
}

func (s *Solver) varBumpActitivyByInc(v Var, inc float64) {
	s.VarOrder.activity[v] += inc
	if s.VarOrder.Activity(v) > 1e100 {
		//Rscale:
		for i := 0; i < s.NumVars(); i++ {
			s.VarOrder.activity[i] *= 1e-100
		}
		s.VarIncreaseRatio *= 1e-100
	}
	if s.VarOrder.InHeap(v) {
		s.VarOrder.Update(v)
	}
}

func (s *Solver) NumVars() int {
	return int(s.NextVar)
}

func (s *Solver) NumClauses() uint64 {
	return s.Statistics.NumClauses
}

func (s *Solver) UncheckedEnqueue(p Lit, from ClauseReference) {
	if s.ValueLit(p) != LitBoolUndef {
		panic(fmt.Sprintf("The assign is not LiteralUndef: ValueLit(%d) = %v", p, s.ValueLit(p)))
	}
	if !p.Sign() {
		s.Assigns[p.Var()] = LitBoolTrue
	} else {
		s.Assigns[p.Var()] = LitBoolFalse
	}
	s.VarData[p.Var()] = *NewVarData(from, s.decisionLevel())
	s.Trail = append(s.Trail, p)
}

func (s *Solver) Propagate() ClauseReference {
	confl := ClaRefUndef

	for s.Qhead < len(s.Trail) {
		p := s.Trail[s.Qhead]
		s.Qhead++
		lastIdx := 0
		copiedIdx := 0
		for lastIdx < len(s.Watches[p]) {
			watcher := s.Watches[p][lastIdx]
			blocker := watcher.blocker

			s.Statistics.PropagationCount++
			// Try to avoid inspecting the clause.
			if s.ValueLit(blocker) == LitBoolTrue {
				s.Watches[p][copiedIdx] = s.Watches[p][lastIdx]
				lastIdx++
				copiedIdx++
				continue
			}

			// Make sure the false literal is data[1]
			cr := watcher.claRef
			clause, err := s.ClaAllocator.GetClause(cr)
			if err != nil {
				panic(err)
			}
			falseLit := p.Flip()
			if clause.At(0) == falseLit {
				clause.Data[0], clause.Data[1] = clause.Data[1], falseLit
			}
			if clause.At(1) != falseLit {
				panic(fmt.Errorf("The 1th literal is not falseLit: %v %v", clause.At(1), falseLit))
			}
			lastIdx++

			// If 0th watch is true, then clause is already satisfied
			firstLiteral := clause.At(0)
			w := NewWatcher(cr, firstLiteral)
			if firstLiteral != blocker && s.ValueLit(firstLiteral) == LitBoolTrue {
				s.Watches[p][copiedIdx] = w
				copiedIdx++
				continue
			}

			// Look for new watch:
			for i := 2; i < clause.Size(); i++ {
				//Find the candidate for watching
				if s.ValueLit(clause.At(i)) != LitBoolFalse {
					clause.Data[1], clause.Data[i] = clause.Data[i], falseLit
					x := clause.At(1)
					s.Watches[x.Flip()] = append(s.Watches[x.Flip()], w)
					goto NextClause
				}
			}
			// Did not find watch -- clause is unit under assignment:
			s.Watches[p][copiedIdx] = w
			copiedIdx++
			if s.ValueLit(firstLiteral) == LitBoolFalse {
				confl = cr
				s.Qhead = len(s.Trail)
				//Copy the remaining watches:
				for lastIdx < len(s.Watches[p]) {
					s.Watches[p][copiedIdx] = s.Watches[p][lastIdx]
					lastIdx++
					copiedIdx++
				}
			} else {
				s.UncheckedEnqueue(firstLiteral, cr)
			}
		NextClause:
		}
		s.Watches[p] = s.Watches[p][:copiedIdx]
	}

	return confl
}

func (s *Solver) CancelUntil(level int) {
	if s.decisionLevel() > level {
		for c := len(s.Trail) - 1; c >= s.TrailLim[level]; c-- {
			x := s.Trail[c].Var()
			s.Assigns[x] = LitBoolUndef
			//TODO Phase Saving
			s.InsertVarOrder(x)
		}
		s.Qhead = s.TrailLim[level]
		s.Trail = s.Trail[:s.Qhead]
		s.TrailLim = s.TrailLim[:level]
	}
}

func (s *Solver) pickBranchLit() Lit {
	// Activity based decision
	nextVar := VarUndef
	for nextVar == VarUndef || s.ValueVar(nextVar) != LitBoolUndef || !s.Decision[nextVar] {
		if s.VarOrder.Empty() {
			nextVar = VarUndef
			break
		}
		nextVar = s.VarOrder.RemoveMin()
	}
	if nextVar == VarUndef {
		return Lit{X: LitUndef}
	}

	return *NewLit(nextVar, true)
}

func (s *Solver) newDecisionLevel() {
	s.TrailLim = append(s.TrailLim, len(s.Trail))
}

func (s *Solver) decisionLevel() int {
	return len(s.TrailLim)
}

func (s *Solver) addClause(lits []Lit) bool {
	if s.decisionLevel() != 0 {
		panic(fmt.Errorf("The decision level is not zero: %d", s.decisionLevel()))
	}
	if !s.OK {
		return false
	}
	// Check if clause is satisfied and remove false/duplicate literals:
	p := Lit{X: LitUndef}
	copiedIdx := 0
	for i := 0; i < len(lits); i++ {
		if s.ValueLit(lits[i]) == LitBoolTrue || lits[i].Equal(p.Flip()) {
			return true
		} else if s.ValueLit(lits[i]) != LitBoolFalse && lits[i].NotEqual(p) {
			lits[copiedIdx], p = lits[i], lits[i]
			copiedIdx++
		}
	}
	lits = lits[:copiedIdx]
	// What clause is empty means that the problem is unsatisfiable
	if len(lits) == 0 {
		s.OK = false
	} else if len(lits) == 1 {
		s.UncheckedEnqueue(lits[0], ClaRefUndef)
		//Found conflict
		if confl := s.Propagate(); confl != ClaRefUndef {
			s.OK = false
		}
	} else {
		claRef, err := s.ClaAllocator.NewAllocate(lits, false)
		if err != nil {
			panic(err)
		}
		s.Clauses[claRef] = true
		err = s.attachClause(claRef)
		if err != nil {
			panic(err)
		}
	}
	return true
}

func (s *Solver) attachClause(claRef ClauseReference) (err error) {
	clause, err := s.ClaAllocator.GetClause(claRef)
	if err != nil {
		return err
	}
	if clause.Size() < 2 {
		return fmt.Errorf("The size of clause is less than 2 %v", clause)
	}

	firstLit := clause.At(0)
	secondLit := clause.At(1)

	s.Watches[firstLit.Flip()] = append(s.Watches[firstLit.Flip()], NewWatcher(claRef, secondLit))
	s.Watches[secondLit.Flip()] = append(s.Watches[secondLit.Flip()], NewWatcher(claRef, firstLit))
	if clause.Learnt() {
		s.Statistics.NumLearnts++
	} else {
		s.Statistics.NumClauses++
	}
	return nil
}

func (s *Solver) Solve() LitBool {
	if !s.OK {
		return LitBoolFalse
	}
	status := LitBoolUndef
	for true {
		status = s.Search(100 * int(s.Statistics.RestartCount+1))
		if status != LitBoolUndef {
			break
		}
		s.Statistics.RestartCount++
	}
	if status == LitBoolTrue {
		for i := 0; i < s.NumVars(); i++ {
			s.Model = append(s.Model, s.ValueVar(Var(i)))
		}
	} else if status == LitBoolFalse {
		s.OK = false
	}
	return status
}

func (s *Solver) Reason(x Var) ClauseReference {
	return s.VarData[x].Reason
}

func (s *Solver) Level(x Var) int {
	return s.VarData[x].Level
}

func (s *Solver) SetDecisionVar(x Var, eligible bool) {
	s.Decision[int(x)] = eligible
	s.InsertVarOrder(x)
}

func (s *Solver) InsertVarOrder(x Var) {
	if !s.VarOrder.InHeap(x) && s.Decision[x] {
		s.VarOrder.PushBack(x)
	}
}

func (s *Solver) Analyze(confl ClauseReference) (learntClause []Lit, backTrackLevel int) {

	p := Lit{X: LitUndef}
	pathConflict := 0
	idx := len(s.Trail) - 1

	learntClause = append(learntClause, p) // (leave room for the asserting literal)
	for {

		if confl == ClaRefUndef {
			pp.Println(s.VarData[p.Var()], p.Var(), s.decisionLevel(), s.ValueLit(p), pathConflict)
			panic("The conflict doesn't point any regisions")
		}
		conflCla, err := s.ClaAllocator.GetClause(confl)
		if err != nil {
			panic(err)
		}
		var startIndex int
		if p.X == LitUndef {
			startIndex = 0
		} else {
			startIndex = 1
		}
		for i := startIndex; i < conflCla.Size(); i++ {
			q := conflCla.At(i)
			if !s.Seen[q.Var()] && s.Level(q.Var()) > 0 {
				s.varBumpActitivy(q.Var())
				s.Seen[q.Var()] = true
				if s.Level(q.Var()) > s.decisionLevel() {
					panic("The decision level of var is greater than or equal to 1")
				}
				if s.Level(q.Var()) == s.decisionLevel() {
					pathConflict++
				} else {
					learntClause = append(learntClause, q)
				}
			}
		}
		// Select next clause to look at:
		update := true
		for update {
			p = s.Trail[idx]
			update = !s.Seen[p.Var()]
			idx--
		}

		confl = s.Reason(p.Var())
		s.Seen[p.Var()] = false
		pathConflict--
		if pathConflict <= 0 {
			break
		}
	}
	learntClause[0] = p.Flip()

	//TODO
	//Simplify conflict clause:

	analyzeToClear := make([]Lit, len(learntClause))
	copy(analyzeToClear, learntClause)

	if len(learntClause) == 1 {
		backTrackLevel = 0
	} else {
		maxIdx := 1
		// Find the first literal assigned at the next-highest level:
		for i := 2; i < len(learntClause); i++ {
			if s.Level(learntClause[i].Var()) > s.Level(learntClause[maxIdx].Var()) {
				maxIdx = i
			}
		}

		backTrackLevel = s.Level(learntClause[maxIdx].Var())
		// Swap-in this literal at index 1:
		learntClause[maxIdx], learntClause[1] = learntClause[1], learntClause[maxIdx]
	}

	for _, lit := range analyzeToClear {
		s.Seen[lit.Var()] = false
	}

	return learntClause, backTrackLevel
}

func (s *Solver) Search(maxConflictCount int) LitBool {
	if !s.OK {
		panic("s.OK is false")
	}

	conflictCount := 0

	for {
		confl := s.Propagate()
		if confl != ClaRefUndef {
			//Conflict
			s.Statistics.ConflictCount++
			conflictCount++

			//If the decision level is 0, the problem is unsatisfiable.
			if s.decisionLevel() == 0 {
				return LitBoolFalse
			}

			learntClause, backTrackLevel := s.Analyze(confl)
			s.CancelUntil(backTrackLevel)

			if len(learntClause) == 1 {
				s.UncheckedEnqueue(learntClause[0], ClaRefUndef)
			} else {
				claRef, err := s.ClaAllocator.NewAllocate(learntClause, true)
				if err != nil {
					panic(err)
				}
				s.LearntClauses[claRef] = true
				err = s.attachClause(claRef)
				if err != nil {
					panic(err)
				}
				s.UncheckedEnqueue(learntClause[0], claRef)
			}

			s.varDecayActivity()

		} else {
			//NO CONFLICT
			if maxConflictCount >= 0 && conflictCount > maxConflictCount {
				//Restart
				s.CancelUntil(0)
				return LitBoolUndef
			}
			nextLit := Lit{X: LitUndef}

			if nextLit.X == LitUndef {
				s.Statistics.DecisionCount++
				nextLit = s.pickBranchLit()
				if nextLit.X == LitUndef {
					// Model found:
					return LitBoolTrue
				}
			}
			s.newDecisionLevel()
			s.UncheckedEnqueue(nextLit, ClaRefUndef)
		}
	}
}
