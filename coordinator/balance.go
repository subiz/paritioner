package coordinator

import (
	"sort"
)

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

func isIn(A []string, s string) bool {
	for _, a := range A {
		if a == s {
			return true
		}
	}
	return false
}

func balance(partitions map[string][]int32, nodes []string) map[string][]int32 {
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
		found := false
		for _, e := range elems {
			if e.id == n {
				found = true
				break
			}
		}

		if !found {
			elems = append(elems, elem{id: n})
		}
	}

	sort.Sort(byLength(elems))

	// count total of partition
	totalPars := 0
	for _, pars := range partitions {
		totalPars += len(pars)
	}

	mod := totalPars % len(nodes)
	numWorkerPars := totalPars / len(nodes)
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
