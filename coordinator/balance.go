package main

import "sort"

type elem struct {
	id      string
	pars    []int32
	numPars int
}

type byLength []elem

func (s byLength) Len() int      { return len(s) }
func (s byLength) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byLength) Less(i, j int) bool {
	if len(s[i].pars) == len(s[j].pars) {
		return s[i].id < s[j].id
	}
	return len(s[i].pars) > len(s[j].pars)
}

// isIn tells whether slice A has an element equal s
func isIn(A []string, s string) bool {
	for _, a := range A {
		if a == s {
			return true
		}
	}
	return false
}

func balance(numPar int32, partitions map[string][]int32, nodes []string) map[string][]int32 {
	if len(nodes) == 0 {
		return nil
	}

	elems := make([]elem, len(partitions))
	i := 0
	for id, pars := range partitions {
		elems[i] = elem{id: id, pars: pars}
		i++
	}

	for _, n := range nodes {
		_, found := partitions[n]
		if !found {
			elems = append(elems, elem{id: n})
		}
	}

	sort.Sort(byLength(elems))

	allpars := make(map[int32]bool)
	for i := int32(0); i < numPar; i++ {
		allpars[i] = true
	}

	// TODO: handle malform pars: 2 worker handle 1 par or out of range par
	// find all missing partition
	for _, pars := range partitions {
		for _, p := range pars {
			delete(allpars, p)
		}
	}
	// we know that elems has at least 1 element
	// use this loop instead of `for p := range allpars {` because we want
	// to preseve increment order
	for i := int32(0); i < numPar; i++ {
		if allpars[i] {
			elems[0].pars = append(elems[0].pars, i)
		}
	}

	mod := int(numPar) % len(nodes)
	numWorkerPars := int(numPar) / len(nodes)
	i = 0
	for k, e := range elems {
		if !isIn(nodes, e.id) {
			continue
		}
		e.numPars = numWorkerPars
		if i < mod {
			e.numPars++
		}
		i++
		elems[k] = e
	}

	redurantPars := make([]int32, 0)
	for k, e := range elems {
		if len(e.pars) > e.numPars { // have redurant job
			redurantPars = append(redurantPars, e.pars[e.numPars:]...)
			e.pars = e.pars[:e.numPars]
			elems[k] = e
		}
	}

	for k, e := range elems {
		if len(e.pars) < e.numPars { // have redurant job
			lenepars := len(e.pars)
			e.pars = append(e.pars, redurantPars[:e.numPars-lenepars]...)

			redurantPars = redurantPars[e.numPars-lenepars:]
			elems[k] = e
		}
	}

	partitions = make(map[string][]int32)
	for _, e := range elems {
		if len(e.pars) > 0 {
			partitions[e.id] = e.pars
		}
	}
	return partitions
}
