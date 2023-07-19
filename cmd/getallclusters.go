package cmd

import (
	"fmt"

	"github.com/KevinWang15/k/pkg/utils"
	"github.com/spf13/cobra"
)

var GetAllClustersCmd = &cobra.Command{
	Use:   "get-all-clusters",
	Short: "Return a list of all clusters",
	Run: func(cmd *cobra.Command, args []string) {
		for _, cluster := range utils.GetConfig().Clusters {
			fmt.Println(cluster.Name)
		}
	},
}
