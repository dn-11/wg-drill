package cmd

import (
	"fmt"
	"wg-natdrill/util"

	"github.com/spf13/cobra"
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install natdrill plugin",
	Long:  `Install natdrill plugin to wireguard tools (detects platform and init system).`,
	Run: func(cmd *cobra.Command, args []string) {
		platform := util.DetectPlatform()
		initSys := util.DetectInitSystem()
		fmt.Printf("Platform: %s\n", platform)
		fmt.Printf("Init: %s\n", initSys)
		// 后续可根据 platform/initSys 执行不同的安装逻辑
	},
}

func init() {
	RootCmd.AddCommand(InstallCmd)
}
