package main

import (
	"fmt"
	"math"
)

type ClauseReference uint32

const ClaRefUndef ClauseReference = math.MaxUint32

//ClauseAllocator is a allocator for the clause
//NOTE we need to improve the performance of alloc/free in the future
type ClauseAllocator struct {
	Qhead      ClauseReference // Qhead is the head of the Clauses
	Clauses    []*Clause       // Clauses is the list of Clause pointer
	WastedSize int             // WastedSize is the number of the removed literal
}

func NewClauseAllocator() *ClauseAllocator {
	return &ClauseAllocator{Qhead: 0, Clauses: []*Clause{}}
}

//NewAllocate allocates a new clause and returns a reference for a clause
func (c *ClauseAllocator) NewAllocate(lits []Lit, learnt bool) (ClauseReference, error) {
	cref := c.Qhead
	if cref >= ClaRefUndef {
		panic(fmt.Errorf("The overflow for a clause allocator happnes"))
	}
	c.Clauses = append(c.Clauses, NewClause(lits, false, learnt))
	c.Qhead++
	return cref, nil
}

//GetClause returns a pointer for a clause
//check whether the reference is invalid or not
func (c *ClauseAllocator) GetClause(claRef ClauseReference) (clause *Clause) {
	claRefInt := int(claRef)
	if claRefInt >= len(c.Clauses) {
		panic(fmt.Errorf("The clause is not allocated: ref = %d size = %d", claRef, len(c.Clauses)))
	}
	cla := c.Clauses[claRef]
	if cla.IsRemoved() {
		panic(fmt.Errorf("This clause is already removed: ref = %d", claRef))
	}
	return cla
}

//FreeClause deletes the clause if the clause is allocated
//NOTE we must not call FreeClause before the removal of the clause
func (c *ClauseAllocator) FreeClause(claRef ClauseReference) {
	claRefInt := int(claRef)
	if claRefInt >= len(c.Clauses) {
		panic(fmt.Errorf("The clause is not allocated: ref = %d size = %d", claRef, len(c.Clauses)))
	}
	cla := c.Clauses[claRef]
	if !cla.IsRemoved() {
		panic(fmt.Errorf("This clause is not removed: ref = %d", claRef))
	}
	c.WastedSize += cla.Size()
}
