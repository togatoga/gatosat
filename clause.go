package main

type Header struct {
	Mark     uint
	Learnt   bool
	HasExtra bool
	Size     int
}

type Clause struct {
	header Header
	Data   []Lit
	Act    float32
}

func NewClause(ps []Lit, useExtra, learnt bool) Clause {
	var c Clause
	c.header.Mark = 0
	c.header.Learnt = learnt
	c.header.HasExtra = useExtra
	c.header.Size = len(ps)

	for i := 0; i < len(ps); i++ {
		c.Data = append(c.Data, ps[i])
	}
	c.Act = 0

	return c
}

func (c *Clause) Size() int {
	return c.header.Size
}

func (c *Clause) Learnt() bool {
	return c.header.Learnt
}

func (c *Clause) HasExtra() bool {
	return c.header.HasExtra
}

func (c *Clause) Mark() uint {
	return c.header.Mark
}

func (c *Clause) At(i int) Lit {
	return c.Data[i]
}

func (c *Clause) Last(i int) Lit {
	return c.Data[c.Size()-1]
}

func (c *Clause) Activity() float32 {
	return c.Act
}
