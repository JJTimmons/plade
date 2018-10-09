package blast

import (
	"sort"

	"github.com/jjtimmons/decvec/config"
	"github.com/jjtimmons/decvec/internal/dvec"
)

// filter "propertizes" the matches from BLAST
//
// propertizing fragment matches means removing those that are completely
// self-contained in other fragments: the larger of the available fragments
// will be the better one, since it covers a greater region and will almost
// always be preferable to the smaller one
//
// also remove small fragments here, that are too small to be useful during
// assembly
func filter(matches []dvec.Match) (properized []dvec.Match) {
	c := config.NewConfig()

	// remove fragments that are shorter the minimum cut off size
	var largeEnough []dvec.Match
	for _, m := range matches {
		if m.Length() < c.Fragments.MinMatch {
			largeEnough = append(largeEnough, m)
		}
	}

	// sort largeEnough by their start index
	// for fragments with equivelant starting indexes, put the larger one first
	sort.Slice(largeEnough, func(i, j int) bool {
		if largeEnough[i].Start != largeEnough[j].Start {
			return largeEnough[i].Start < largeEnough[j].Start
		}
		return largeEnough[i].Length() > largeEnough[j].Length()
	})

	// only include those that aren't encompassed by the one before it
	for _, m := range largeEnough {
		lastMatch := len(properized) - 1
		if lastMatch < 0 || m.End > properized[lastMatch].End {
			properized = append(properized, m)
		}
	}

	return properized
}
