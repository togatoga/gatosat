package main

import (
	"fmt"
)

const (
	ExistMark   uint = iota
	DeletedMark uint = iota
)

//Header is the structure for additional information for a clause
type Header struct {
	Mark     uint // The Marks represents whether the clause already is deleted or not
	Learnt   bool // The Learnt represents whether the clause is a learnt clause or not
	HasExtra bool // TODO
	Lbd      int  // The value of the Literal Block Distance
	Size     int  // The Size represents the number of the clause
}

//Clause is the structure for core information for a clause
type Clause struct {
	header Header  // The header represents
	Data   []Lit   // The Data is the list of the literal
	Act    float32 // The Act is the clause activity. when we need to delete clauses, we use it
}

//NewClause returns a pointer of a new clause
func NewClause(ps []Lit, useExtra, learnt bool) *Clause {
	var c Clause
	c.header.Mark = ExistMark
	c.header.Learnt = learnt
	c.header.HasExtra = useExtra
	c.header.Size = len(ps)
	c.header.Lbd = 0
	c.Data = make([]Lit, len(ps))

	copy(c.Data, ps)
	c.Act = 0

	return &c
}

//Size returns the size of a clause
func (c *Clause) Size() int {
	return c.header.Size
}

//Learnt returns a boolean indicating whether a clause is learnt or not
func (c *Clause) Learnt() bool {
	return c.header.Learnt
}

//HasExtra returns a boolean indicating whether a HasExtra is true or not
func (c *Clause) HasExtra() bool {
	return c.header.HasExtra
}

//SetMark sets a mark
func (c *Clause) SetMark(mark uint) {
	c.header.Mark = mark
}

//Mark returns a mark
func (c *Clause) Mark() uint {
	return c.header.Mark
}

//At returns a i idx lit for a clause
func (c *Clause) At(i int) Lit {
	return c.Data[i]
}

//Pop decreases the size of clause by 1
func (c *Clause) Pop() {
	if c.Size() == 0 {
		panic(fmt.Errorf("Pop empty clause"))
	}
	c.header.Size -= 1
}

//Last returns a last Lit for a clause
func (c *Clause) Last() Lit {
	return c.Data[c.Size()-1]
}

//Activity returns an activity for a clause
func (c *Clause) Activity() float32 {
	return c.Act
}

//IsRemoved returns boolean whether the clause is removed or not
func (c *Clause) IsRemoved() bool {
	return c.header.Mark == DeletedMark
}

func (s *Solver) removeSatisfied(data *[]ClauseReference) {
	copiedIdx := 0
	for lastIdx := 0; lastIdx < len(*data); lastIdx++ {
		c := s.ClaAllocator.GetClause((*data)[lastIdx])
		if s.satisfied(c) {
			s.removeClause((*data)[lastIdx])
		} else {
			//Trim Clause
			if !(s.ValueLit(c.At(0)) == LitBoolUndef && s.ValueLit(c.At(1)) == LitBoolUndef) {
				panic(fmt.Errorf("The 0th and 1th of clause value is not LitBoolUndef: v1: %d = %d v2: %d = %d", c.At(0), s.ValueLit(c.At(0)), c.At(1), s.ValueLit(c.At(1))))
			}
			for k := 2; k < c.Size(); k++ {
				if s.ValueLit(c.At(k)) == LitBoolFalse {
					c.Data[k] = c.Last()
					k--
					c.Pop()
				}
			}
			(*data)[copiedIdx] = (*data)[lastIdx]
			copiedIdx++
		}
	}
	(*data) = (*data)[:copiedIdx]
}

func (s *Solver) detachClause(cr ClauseReference) {
	c := s.ClaAllocator.GetClause(cr)
	if c.Size() <= 1 {
		panic(fmt.Errorf("The size of clause is less than 2: %d", c.Size()))
	}
	firstLit := c.At(0)
	secondLit := c.At(1)
	RemoveWatcher(s.Watches, firstLit.Flip(), NewWatcher(cr, secondLit))
	RemoveWatcher(s.Watches, secondLit.Flip(), NewWatcher(cr, firstLit))
	if c.Learnt() {
		s.Statistics.NumLearnts--
	} else {
		s.Statistics.NumClauses--
	}
}

func (s *Solver) locked(c *Clause) bool {
	firstLit := c.At(0)
	if s.ValueLit(firstLit) == LitBoolTrue && s.Reason(firstLit.Var()) != ClaRefUndef {
		return true
	}
	return false
}

func (s *Solver) satisfied(c *Clause) bool {
	for i := 0; i < c.Size(); i++ {
		if s.ValueLit(c.At(i)) == LitBoolTrue {
			return true
		}
	}
	return false
}

func (s *Solver) removeClause(cr ClauseReference) {
	c := s.ClaAllocator.GetClause(cr)
	s.detachClause(cr)
	firstLit := c.At(0)
	if s.locked(c) {
		s.VarData[firstLit.Var()].Reason = ClaRefUndef
	}
	c.SetMark(DeletedMark)
	s.ClaAllocator.FreeClause(cr)
}

func (s *Solver) attachClause(claRef ClauseReference) (err error) {
	clause := s.ClaAllocator.GetClause(claRef)

	if clause.Size() < 2 {
		return fmt.Errorf("The size of clause is less than 2 %v", clause)
	}

	firstLit := clause.At(0)
	secondLit := clause.At(1)
	s.Watches.Append(firstLit.Flip(), NewWatcher(claRef, secondLit))
	s.Watches.Append(secondLit.Flip(), NewWatcher(claRef, firstLit))

	if clause.Learnt() {
		s.Statistics.NumLearnts++
	} else {
		s.Statistics.NumClauses++
	}
	return nil
}
