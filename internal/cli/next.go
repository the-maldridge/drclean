package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/the-maldridge/drclean/internal/registry"
)

func init() {
	rootCmd.AddCommand(nextCmd)
}

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "summon the next sequential version number",
	Long:  "next finds the next monotonically increasing version number for a given repo.",
	Args:  cobra.ExactArgs(1),
	Run:   nextCmdRun,
}

func nextCmdRun(cmd *cobra.Command, args []string) {
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

	tags, _, err = reg.FindBadTags(tags)
	if err != nil {
		fmt.Printf("Error filtering tags: %s", err)
		os.Exit(1)
	}

	tags = reg.SortTagsFull(tags)
	parts := strings.Split(tags[len(tags)-1], viper.GetString("tag.seperator"))

	// Check if its the first version for today.
	date, _ := time.Parse(viper.GetString("tag.dateformat"), parts[0])
	y1, m1, d1 := date.Date()
	y2, m2, d2 := time.Now().Date()

	// If it is, just return the first revision.
	if y1 != y2 || m1 != m2 || d1 != d2 {
		fmt.Printf("%d%02d%02d%s01\n", y2, m2, d2, viper.GetString("tag.seperator"))
		os.Exit(0)
	}

	rel, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("%s%s%02d\n", strings.Join(parts[:1], ""), viper.GetString("tag.seperator"), rel+1)
}
