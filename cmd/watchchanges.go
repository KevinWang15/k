package cmd

import (
	"github.com/KevinWang15/k/pkg/watchchanges"
	"github.com/spf13/cobra"
)

var WatchChangesCmd = &cobra.Command{
	Use:   "watch-changes",
	Short: "watch-changes (internal command)",
	Run: func(cmd *cobra.Command, args []string) {
		watchchanges.Run()
	},
}
