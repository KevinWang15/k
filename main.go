package main

import (
	"github.com/KevinWang15/k/cmd"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "k",
	}

	rootCmd.AddCommand(cmd.RcCmd)
	rootCmd.AddCommand(cmd.WatchChangesCmd)
	rootCmd.AddCommand(cmd.GetAllClustersCmd)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
