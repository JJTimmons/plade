// Package traverse is for traversing the target matches and
// creating a list of possible assemblies for making the target
// sequence
package traverse

import (
	"math"
	"sort"

	"github.com/jjtimmons/decvec/config"
	"github.com/jjtimmons/decvec/internal/frag"
)

// node is a single node within the DP tree for building up
// the vector from smaller building fragments
type node struct {
	// start of this node on the target vector (which has been 3x'ed for BLAST)
	start int

	// end of this node on the target vector
	end int
}

// Traverse the matches on the target fragment
// and build up a list of possible assemblies
func Traverse(f *frag.Fragment) {
	c := config.NewConfig()
	firstBP := len(f.Seq)

	// map fragment Matches to nodes
	var nodes []node
	for _, m := range f.Matches {
		nodes = append(nodes, node{
			start: m.Start,
			end:   m.End,
		})
	}

	// get each fragments minimum number of fragments in an assembly
	dists := distanceToEnd(nodes, len(f.Seq), c.Synthesis.MaxLength)

	// remove all fragments that are greater than the target distance
	var filteredNodes []node
	for m, dist := range dists {
		// if it overlaps with the start of the search range
		// 	compare it against the max-distance from the end
		// else
		// 	check if it's max-distance - 1 from end and remove
		// 	it if it's not
		if m.start < firstBP {
			if dist < c.Fragments.MaxCount {
				filteredNodes = append(filteredNodes, m)
			}
		} else {
			if dist+1 < c.Fragments.MaxCount {
				filteredNodes = append(filteredNodes, m)
			}
		}
	}
}

// distanceToEnd is for calculating each fragment's minimum distance to
// a vector "endpoint". used to weed out fragments that are too far
// away from the end of the target
//
// for each building fragment, find the minimum number of fragments
// needed to get from it to a "last-bp" (2x the length of the target
// fragment's sequence length)
//
// maxSynth is the maximum length of a synthesized stretch of DNA
func distanceToEnd(nodes []node, seqL int, maxSynth int) map[node]int {
	// last bp within range being scanned
	lastBP := 2 * seqL
	// map from a fragment to the minimum number of fragments
	// needed in an assembly including it
	dists := make(map[node]int)

	// sort by start
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].start < nodes[j].start
	})

	// add a "sink" to ensure there's a building fragment just past
	// the end of the scanned range
	sink := node{
		start: lastBP + 1,
		end:   lastBP + 1,
	}
	nodes = append(nodes, sink)

	// number of fragments distance for this fragment
	// through to the target bp (2x the seqLength)
	//
	// if dist has been calculated already
	// 	return that
	// else if this fragment ends after the lastBP,
	// 	return 1 and store 1 to the cache for this fragment
	// else
	//	return the min-distance from this to all fragments
	// 	this could reasonably be assembled with
	var distFor func(int, node) int
	distFor = func(i int, f node) int {
		if dist, cached := dists[f]; cached {
			return dist
		}

		if f.end >= lastBP {
			dists[f] = 1
			return 1
		}

		// calculate how many synthesis fragments are necessary to get
		// from this fragment to each fragment after it
		var synthCount []int
		for _, n := range nodes[i+1:] {
			synthsToNext := 0
			distToNext := n.start - f.end
			if distToNext > 0 {
				synthsToNext = int(math.Ceil(float64(distToNext) / float64(maxSynth)))
			}
			synthCount = append(synthCount, synthsToNext)
		}

		// find the minimum distance among these options, accounting
		// for the fact that we need to each synthesis is an additional (unseen) fragment
		minDistNext := 1 + synthCount[0] + distFor(i+1, nodes[i+1])
		for j, m := range nodes[i+1:] {
			distToFrag := 1 + synthCount[j] + distFor(i+j, m)
			if distToFrag < minDistNext {
				minDistNext = distToFrag
			}
		}

		// store in cache in case this is referenced later
		dists[f] = minDistNext
		return minDistNext
	}

	// fill cache
	distFor(0, nodes[0])

	// delete the sink fragment
	delete(dists, sink)

	return dists
}
