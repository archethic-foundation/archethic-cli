package cli

import (
	"encoding/json"
	"fmt"

	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	"github.com/spf13/cobra"
)

func GetKeychainCmd() *cobra.Command {
	getKeychainCmd := &cobra.Command{
		Use:   "get-keychain",
		Short: "Get keychain",
		Run: func(cmd *cobra.Command, args []string) {
			err := validateRequiredFlags(cmd.Flags(), "ssh", "ssh-path", "access-seed", "mnemonic")
			cobra.CheckErr(err)
			accessSeedBytes, err := tuiutils.GetSeedBytes(cmd.Flags(), "ssh", "ssh-path", "access-seed", "mnemonic")
			cobra.CheckErr(err)
			keychain, err := tuiutils.AccessKeychain(endpoint.String(), accessSeedBytes)
			cobra.CheckErr(err)
			jsonServices, err := json.Marshal(keychain.Services)
			cobra.CheckErr(err)
			fmt.Printf("%s\n", jsonServices)
		},
	}
	getKeychainCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	getKeychainCmd.Flags().String("access-seed", "", "Access Seed")
	getKeychainCmd.Flags().Bool("ssh", false, "Enable SSH key mode")
	getKeychainCmd.Flags().String("ssh-path", GetFirstSshKeyDefaultPath(), "Path to ssh key")
	getKeychainCmd.Flags().Bool("mnemonic", false, "Enable mnemonic words for seed")
	getKeychainCmd.MarkFlagsMutuallyExclusive("access-seed", "ssh")
	getKeychainCmd.MarkFlagsMutuallyExclusive("access-seed", "ssh-path")
	getKeychainCmd.MarkFlagsMutuallyExclusive("mnemonic", "ssh")
	getKeychainCmd.MarkFlagsMutuallyExclusive("mnemonic", "ssh-path")
	getKeychainCmd.MarkFlagsMutuallyExclusive("mnemonic", "access-seed")
	return getKeychainCmd
}
