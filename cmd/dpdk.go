package cmd

import (
	"github.com/spf13/cobra"
)

func dpdkStatus() *cobra.Command {
	return &cobra.Command{
		Use:   "dpdk_status",
		Short: "Set dpdk status,  enable, disable or nil",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
}

func dpdkPort() *cobra.Command {
	return &cobra.Command{
		Use:   "dpdk_port",
		Short: "Set dpdk port",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
}

func dpdkOptions() *cobra.Command {
	return &cobra.Command{
		Use:   "dpdk_options",
		Short: "Set dpdk args",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
}
