package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chaordic-io/venue-cluster/internal"
	"github.com/spf13/cobra"
)

func main() {
	err := os.Setenv("ANSIBLE_HOST_KEY_CHECKING", "False")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if !internal.HasDependencies() {
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "hshstack",
		Short: "sets up the hashistack",
		Long:  `hshstack - setups up Consul & Nomad with ACL & Service Mesh enabled`,
		Run: func(cmd *cobra.Command, args []string) {
			err = cmd.Help()
			if err != nil {
				panic(err)
			}
		},
	}

	rootCmd.AddCommand(sync(), destroy(), observability())

	err = rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func sync() *cobra.Command {
	var configFile string
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "bootstraps and starts a cluster or syncs the cluster to its desired state",
		Long:  `bootstraps and starts a cluster or syncs the cluster to its desired state`,
		Run: func(cmd *cobra.Command, args []string) {
			config, err := internal.LoadConfig(configFile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = internal.Bootstrap(config, configFile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	addFlags(cmd, &configFile)

	return cmd
}

func observability() *cobra.Command {
	var configFile string
	cmd := &cobra.Command{
		Use:   "observability",
		Short: "adds observability to a cluster",
		Long:  `adds observability a cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			config, err := internal.LoadConfig(configFile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = internal.Observability(filepath.Join(config.BaseDir, "inventory"), configFile, config.BaseDir, config.CloudProviderConfig.User)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	addFlags(cmd, &configFile)

	return cmd
}

func destroy() *cobra.Command {
	var configFile string
	yes := ""
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "destroys a cluster",
		Long:  `destroys a cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			text := yes
			if yes != "yes" {
				fmt.Println("Are you sure you want to delete all resources? ('yes', or any other input for no)")
				reader := bufio.NewReader(os.Stdin)
				text, err = reader.ReadString('\n')
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
			if text == "yes\n" || text == "yes" {
				config, err := internal.LoadConfig(configFile)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				err = internal.Destroy(filepath.Join(config.BaseDir, "inventory"), config.BaseDir, config.CloudProviderConfig.User)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				return
			}
			fmt.Println("Delete cancelled")
		},
	}
	addFlags(cmd, &configFile)

	return cmd
}

func addFlags(cmd *cobra.Command, file *string) {
	cmd.Flags().StringVarP(file, "config.file", "f", "", "configuration file")

	err := cmd.MarkFlagRequired("config.file")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
