package main

//ComputeLBD computes the literal block distance for lits and returns a value
func (s *Solver) ComputeLBD(lits []Lit) int {
	//TODO
	// We can replace it with faster one
	// There is a performance problem to compute lbd by using map
	permDiff := map[int]bool{}
	for i := 0; i < len(lits); i++ {
		permDiff[s.Level(lits[i].Var())] = true
	}
	return len(permDiff)
}

//LBD returns a value of the Literal block distance for a clause
func (c *Clause) LBD() int {
	return c.header.Lbd
}

//SetLBD sets the literal block distance for a clause
func (c *Clause) SetLBD(lbd int) {
	c.header.Lbd = lbd
}
