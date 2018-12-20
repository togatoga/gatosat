package main

//Watcher is the struct to detect conflicts
type Watcher struct {
	claRef  ClauseReference //claRef is a reference for a clause
	blocker Lit             //blocker is a checker variable whether a clause is conflicted or not
}

//NewWatcher returns a pointer of Watcher
func NewWatcher(cla ClauseReference, p Lit) *Watcher {
	return &Watcher{
		claRef:  cla,
		blocker: p,
	}
}

//Equal returns a boolean indicating a clause reference is equal
func (w *Watcher) Equal(wr Watcher) bool {
	if w.claRef == wr.claRef {
		return true
	}
	return false
}

//Watches is a struct for watchers
type Watches struct {
	watches [][]*Watcher
}

//NewWatches returns a pointer of Watches
func NewWatches() *Watches {
	return &Watches{}
}

//Init append a new empty watcher if the size of watches is greater than a variable
func (w *Watches) Init(v Var) {
	size := 2*int(v) + 1
	for len(w.watches) <= size {
		w.watches = append(w.watches, []*Watcher{})
	}
}

//Lookup returns a pointer of literal's watches
func (w *Watches) Lookup(x Lit) *[]*Watcher {
	idx := LitToInt(x)
	return &(w.watches[idx])
}

//Append appends a new watcher to watches
func (w *Watches) Append(x Lit, watcher *Watcher) {
	idx := LitToInt(x)
	w.watches[idx] = append(w.watches[idx], watcher)
}

//RemoveWatcher removes a watcher which has literal x from watches
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
	if startCopyIdx == -1 {
		panic("Wacher is not found")
	}

	//Copy the rest of watcher exclude the value of startCopyIdx
	for copiedIdx := startCopyIdx; copiedIdx < len(*ws)-1; copiedIdx++ {
		(*ws)[copiedIdx] = (*ws)[copiedIdx+1]
	}
	//pop
	*ws = (*ws)[:len(*ws)-1]
}
