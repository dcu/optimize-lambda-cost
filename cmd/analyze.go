package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dcu/optimize-lambda-cost/analyze"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
	"github.com/spf13/cobra"
)

var (
	awsProfile string
	since      string
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze <function-name>",
	Short: "Analyzes the costs of a Lambda function",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Println(cmd.Usage())
			return nil
		}

		startTime := time.Now().Add(-30 * time.Minute)
		if since != "" {
			w := when.New(nil)
			w.Add(en.All...)
			w.Add(common.All...)

			r, err := w.Parse(since, time.Now())
			if err != nil {
				return err
			}

			startTime = r.Time
		}

		log.Println("Start fetching logs for", args[0], "starting on", startTime)

		analyzer := analyze.New(awsProfile)
		buckets, err := analyzer.FetchBuckets(args[0], startTime, 5*time.Minute)
		if err != nil {
			log.Println("error fetching logs:", err.Error())
			return err
		}

		for _, bucket := range buckets {
			bucket.Print(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(&awsProfile, "profile", "p", "", "AWS to use to authenticate")
	analyzeCmd.Flags().StringVarP(&since, "since", "s", "", "Time to start to fetch")
}
