package main

type Var int

type Lit struct {
	X int
}

func NewLit(x Var, sign bool) Lit {
	var p Lit
	y := 2 * x
	if sign == true {
		y++
	}
	p.X = int(y)
	return p
}

func (l *Lit) Sign() bool {
	if l.X&1 != 0 {
		return true
	}
	return false
}
