package defrag

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jjtimmons/defrag/config"
)

// node is a single node within the DP tree for building up
// the vector from smaller building fragments
type node struct {
	// id of the node's source in the database (will be used to avoid off-targets within it)
	id string

	// seq of the fragment
	seq string

	// unique-id of a match, the  start index % seq-length + id
	// (unique identified that catches nodes that cross the zero-index)
	uniqueID string

	// start of this node on the target vector (which has been 3x'ed for BLAST)
	start int

	// end of this node on the target vector
	end int

	// assemblies that span from this node to the end of the vector
	assemblies []assembly

	// assembly configuration
	conf *config.Config

	// the cost of the node. eg: fragments from Addgene cost $65
	cost float64
}

// new creates a node from a Match
func new(m Match, seqL int, conf *config.Config) node {
	cost := 0.0
	if strings.Contains(m.Entry, "addgene") {
		cost = conf.AddGeneVectorCost
	}

	return node{
		id:       m.Entry,
		seq:      strings.ToUpper(m.Seq),
		uniqueID: strconv.Itoa(m.Start%seqL) + m.Entry,
		start:    m.Start,
		end:      m.End,
		conf:     conf,
		cost:     cost,
	}
}

// fragment converts a node into a fragment
func (n *node) fragment() Fragment {
	return Fragment{
		ID:    n.id,
		Seq:   n.seq,
		Entry: n.id,
		Type:  PCR,
	}
}

// distTo returns the distance between the start of this node and the end of the other.
// assumes that this node starts before the other
// will return a negative number if this node overlaps with the other and positive otherwise
func (n *node) distTo(other node) (bpDist int) {
	return other.start - n.end
}

// synthDist returns the number of synthesized fragments that would need to be created
// between one node and another if the two were to be joined, with no existing
// fragments/nodes in-between, in an assembly
func (n *node) synthDist(other node) (synthCount int) {
	dist := n.distTo(other)

	if dist <= 5 {
		// if the dist is <5, we can try and PCR our way there
		return 0
	}

	// if it's > 5, we have to synthesize to get there (arbitatry, TODO: move to settings)
	// can't be negative before the ceil
	floatDist := math.Max(1.0, float64(dist))

	// split up the distance between them by the max synthesized fragment size
	return int(math.Ceil(floatDist / float64(n.conf.Synthesis.MaxLength)))
}

// costTo calculates the $ amount needed to get from this fragment
// to the other node passed, either by PCR or synthesis
//
// PCR is, of course, preferred if we can add homology between this and
// the other fragment with PCR homology arms alone
//
// Otherwise we find the total synthesis distance between this and
// the other fragment and divide that by the cost per bp of synthesized DNA
//
// Account for the cost of the other node (eg if the node has to be ordered from
// Addgene there's a cost in getting it)
func (n *node) costTo(other node) (cost float64) {
	dist := n.distTo(other)

	if dist <= 5 {
		if dist < -(n.conf.Fragments.MinHomology) {
			// there's already enough overlap between this node and the one being tested
			// estimating two primers, 20bp each
			return 40*n.conf.PCR.BPCost + n.cost
		}

		// we have to create some additional primer sequence to reach the next fragment
		// guessing 40bp plus half MinHomology on each primer
		return float64(40+n.conf.Fragments.MinHomology)*n.conf.PCR.BPCost + n.cost
	}

	// we need to create a new synthetic fragment to get from this fragment to the next
	// to account for both the bps between them as well as the additional bps we need to add
	// for homology between the two
	fragLength := n.conf.Fragments.MinHomology + dist

	return n.conf.SynthCost(fragLength) + n.cost
}

// reach returns a slice of node indexes that overlap with, or are the first synth_count nodes
// away from this one within a slice of ordered nodes
func (n *node) reach(nodes []node, i, synthCount int) (reachable []int) {
	reachable = []int{}

	// accumulate the nodes that overlap with this one
	for true {
		i++

		// we've run out of nodes
		if i >= len(nodes) {
			return reachable
		}

		// these nodes overlap by enough for assembly without PCR
		if n.distTo(nodes[i]) <= -(n.conf.Fragments.MinHomology) {
			reachable = append(reachable, i)
		} else if synthCount > 0 {
			// there's not enough existing overlap, but we can synthesize to it
			synthCount--
			reachable = append(reachable, i)
		} else {
			break
		}
	}

	return
}

// synthTo returns synthetic fragments to get this node to the next.
// It creates a slice of building fragments that have homology against
// one another and are within the upper and lower synthesis bounds.
// seq is the vector's sequence. We need it to build up the target
// vector's sequence
func (n *node) synthTo(next node, seq string) (synthedFrags []Fragment) {
	// check whether we need to make synthetic fragments to get
	// to the next fragment in the assembly
	fragC := n.synthDist(next)
	if fragC == 0 {
		return nil
	}

	// make the slice
	synthedFrags = []Fragment{}

	// length of each synthesized fragment
	fragL := n.distTo(next) / fragC
	if n.conf.Synthesis.MinLength > fragL {
		// need to synthesize at least Synthesis.MinLength bps
		fragL = n.conf.Synthesis.MinLength
	}

	// account for homology on either end of each synthetic fragment
	fragL += n.conf.Fragments.MinHomology * 2
	seq += seq // double to account for sequence across the zero-index

	// slide along the range of sequence to create synthetic fragments for
	// and create one at each point, each w/ MinHomology for the fragment
	// before it and after it
	for fragIndex := 0; fragIndex < int(fragC); fragIndex++ {
		start := n.end - n.conf.Fragments.MinHomology // start w/ homology
		start += fragIndex * fragL                    // slide along the range to cover
		end := start + fragL + n.conf.Fragments.MinHomology

		sFrag := Fragment{
			ID:   fmt.Sprintf("%s-synthetic-%d", n.id, fragIndex+1),
			Seq:  seq[start:end],
			Type: Synthetic,
			Cost: n.conf.SynthCost(len(n.seq)) + n.cost,
		}
		synthedFrags = append(synthedFrags, sFrag)
	}

	return
}