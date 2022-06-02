package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/chaordic-io/venue-cluster/internal"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "hshstack",
		Short: "sets up the hashistack",
		Long:  `hshstack - setups up Consul & Nomad with ACL & Service Mesh enabled`,
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				panic(err)
			}
		},
	}

	rootCmd.AddCommand(bootstrap(), destroy())

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func bootstrap() *cobra.Command {
	var inventoryFile string
	var user string
	var dcName string
	cmd := &cobra.Command{
		Use:   "up",
		Short: "bootstraps and starts a cluster",
		Long:  `bootstraps and starts a cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			err := internal.Bootstrap(inventoryFile, dcName, user)
			if err != nil {

				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	addFlags(cmd, &inventoryFile, &user)
	cmd.Flags().StringVarP(&dcName, "datacentre", "d", "", "name of data center")

	err := cmd.MarkFlagRequired("datacentre")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return cmd
}

func destroy() *cobra.Command {
	var inventoryFile string
	var user string
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
				internal.Destroy(inventoryFile, user)
				return
			}
			fmt.Println("Delete cancelled")
		},
	}
	addFlags(cmd, &inventoryFile, &user)

	return cmd
}

func addFlags(cmd *cobra.Command, inventoryFile, user *string) {
	cmd.Flags().StringVarP(inventoryFile, "inventory", "i", "", "inventory file")
	cmd.Flags().StringVarP(user, "user", "u", "", "User to run as against servers")

	err := cmd.MarkFlagRequired("inventory")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = cmd.MarkFlagRequired("user")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
