package assemble

// assembly is a slice of nodes ordered by the nodes
// distance from the end of the target vector
type assembly struct {
	// nodes, ordered by distance from the "end" of the vector
	nodes []node

	// estimated cost of making this assembly
	cost float32

	// total number of synthetic nodes that will be needed to make this
	synths int
}

// add a node to the start of this assembly.
// Update the cost of the assembly to include the link between the new first node and the one after it.
// Store the node's id in the list of node ids.
// Complete if a node has matched up onto itself across the zero-index.
// TODO: incorporate cost estimate of the last node in an assembly
func (a *assembly) add(n node) (newAssembly assembly, created, complete bool) {
	// we've complete an assembly if the node being added is the same as the
	// first one in this assembly
	complete = n.uniqueID == a.nodes[len(a.nodes)-1].uniqueID
	if complete {
		return *a, true, complete
	}

	// add to list of nodes, update cost, and return
	if len(a.nodes) > 0 {
		// calc the number of synthesis fragment to get to next fragment
		synths := a.nodes[0].synthDist(n)

		// stay beneath upper limit
		if a.len()+synths+1 > conf.Fragments.MaxCount {
			return newAssembly, false, false
		}

		return assembly{
			nodes:  append([]node{n}, a.nodes...),
			cost:   n.costTo(a.nodes[0]),
			synths: a.synths + synths,
		}, true, false
	}

	// create the start of this assembly, no other nodes
	return assembly{[]node{n}, 0, 0}, true, false
}

// contains returns if the id of the node has already been seen in this assembly
func (a *assembly) contains(n node) (isContained bool) {
	for _, otherN := range a.nodes {
		// they're the same if they have the same id and start index
		// id isn't enough by itself because there may be multiple with the same
		// entry id in the BLAST db
		if otherN.uniqueID == n.uniqueID {
			return true
		}
	}
	return false
}

// len returns len(nodes) + the synthesis fragment count
func (a *assembly) len() int {
	return len(a.nodes) + a.synths
}