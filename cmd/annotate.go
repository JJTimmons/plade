package cmd

import (
	"github.com/jjtimmons/defrag/internal/defrag"
	"github.com/spf13/cobra"
)

// annotateCmd is for finding features or enzymes by their name.
var annotateCmd = &cobra.Command{
	Use:                        "annotate [seq]",
	Run:                        defrag.Annotate,
	Short:                      "Annotate a vector sequence against the feature db",
	SuggestionsMinimumDistance: 3,
	Long:                       ``,
}

// set flags
func init() {
	annotateCmd.Flags().StringP("in", "i", "", "input file name")
	annotateCmd.Flags().StringP("out", "o", "", "output file name")
	annotateCmd.Flags().StringP("exclude", "x", "", "keywords for excluding features")
	annotateCmd.Flags().IntP("identity", "t", 100, "match %-identity threshold (see 'blastn -help')")
	annotateCmd.Flags().BoolP("enclosed", "c", false, "annotate with features enclosed in others")

	rootCmd.AddCommand(annotateCmd)
}