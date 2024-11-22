package cmd

import (
	"github.com/KevinWang15/k/pkg/rc"
	"github.com/spf13/cobra"
)

var RcCmd = &cobra.Command{
	Use:   "rc",
	Short: "generate rc commands",
	Long:  `generate rc commands. "source <(k rc)" in your .profile`,
	Run: func(cmd *cobra.Command, args []string) {
		rc.Run()
	},
}
