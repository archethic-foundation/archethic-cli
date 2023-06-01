package cli

import (
	"fmt"

	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	archethic "github.com/archethic-foundation/libgo"
	"github.com/spf13/cobra"
)

func GetDeleteServiceFromKeychainCmd() *cobra.Command {
	deleteServiceFromKeychainCmd := &cobra.Command{
		Use:   "delete-service-from-keychain",
		Short: "Delete service from keychain",
		Run: func(cmd *cobra.Command, args []string) {
			accessSeed, _ := cmd.Flags().GetString("access-seed")
			serviceName, _ := cmd.Flags().GetString("service-name")

			fmt.Println("Deleting service from keychain...")
			accessSeedBytes, err := archethic.MaybeConvertToHex(accessSeed)
			if err != nil {
				fmt.Println(err)
				return
			}
			feedback, err := tuiutils.RemoveServiceFromKeychain(accessSeedBytes, endpoint.String(), serviceName)
			if err != nil {
				fmt.Println(err)
				return
			} else {
				fmt.Println(feedback)
			}
		},
	}
	deleteServiceFromKeychainCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	deleteServiceFromKeychainCmd.Flags().String("access-seed", "", "Access Seed")
	deleteServiceFromKeychainCmd.Flags().String("service-name", "", "Service Name")
	return deleteServiceFromKeychainCmd
}
