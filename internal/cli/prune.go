package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/the-maldridge/drclean/internal/registry"
)

func init() {
	rootCmd.AddCommand(pruneCmd)
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "prune removes versions from a repository",
	Long:  "prune uses the criteria of max age of a compliant tag or number of versions to determine how many to keep, then removes versions that do not meet the policy for retained versions.",
	Args:  cobra.ExactArgs(1),
	Run:   pruneCmdRun,
}

func pruneCmdRun(cmd *cobra.Command, args []string) {
	reg, err := registry.New()
	if err != nil {
		fmt.Printf("Error during initialization: %s", err)
		os.Exit(1)
	}

	tags, err := reg.GetTags(args[0])
	if err != nil {
		fmt.Printf("Error retrieving tags: %s", err)
		os.Exit(1)
	}

	tags, bad, err := reg.FindBadTags(tags)
	if err != nil {
		fmt.Printf("Error filtering tags: %s", err)
		os.Exit(1)
	}

	fmt.Println("Bad Tags:")
	for _, bt := range bad {
		fmt.Printf("  %s\n", bt)
	}

	tags = reg.SortTagsByDate(tags)
	keep, toss := reg.KeepTags(tags)

	fmt.Println("Tags to be removed")
	for _, t := range toss {
		fmt.Printf("  %s\n", t)
	}

	fmt.Println("Tags to be retained")
	for _, t := range keep {
		fmt.Printf("  %s\n", t)
	}

	reg.RemoveTags(args[0], toss)
}
