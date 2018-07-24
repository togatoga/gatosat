package main

type Watcher struct {
	claRef  ClauseReference
	blocker Lit
}

func NewWatcher(cla ClauseReference, p Lit) *Watcher {
	return &Watcher{
		claRef:  cla,
		blocker: p,
	}
}

func (w *Watcher) Equal(wr Watcher) bool {
	if w.claRef == wr.claRef {
		return true
	}
	return false
}

type Watches struct {
	watches [][]*Watcher
}

func NewWatches() *Watches {
	return &Watches{}
}

func (w *Watches) Init(v Var) {
	size := 2*int(v) + 1
	if len(w.watches) <= size {
		w.watches = make([][]*Watcher, size+1)
	}
}

func (w *Watches) Lookup(x Lit) *[]*Watcher {
	idx := LitToInt(x)
	return &(w.watches[idx])
}

func (w *Watches) Append(x Lit, watcher *Watcher) {
	idx := LitToInt(x)
	w.watches[idx] = append(w.watches[idx], watcher)
}

func RemoveWatcher(watches *Watches, x Lit, watcher *Watcher) {
	startCopyIdx := -1
	//Find the index of watcher
	ws := watches.Lookup(x)
	for i := 0; i < len(*ws); i++ {
		if (*ws)[i].Equal(*watcher) {
			startCopyIdx = i
			break
		}
	}
	//Copy the rest of watcher exclude the value of startCopyIdx
	for copiedIdx := startCopyIdx; copiedIdx < len(*ws)-1; copiedIdx++ {
		(*ws)[copiedIdx] = (*ws)[copiedIdx+1]
	}
	//pop
	*ws = (*ws)[:len(*ws)-1]
}
