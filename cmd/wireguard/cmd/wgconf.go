package cmd

import (
	"errors"
	"net/http"
	"os"

	cobra "github.com/spf13/cobra"
	tptctl_cmd "github.com/threeport/threeport/cmd/tptctl/cmd"
	cli "github.com/threeport/threeport/pkg/cli/v0"
)

var getWireguardConfVersion string

// GetWireguardConfCmd represents the wireguard-definition command
var GetWireguardConfCmd = &cobra.Command{
	Example: "  tptctl wireguard get wireguard-conf",
	Long:    "Get wireguard configuration from the system.",
	PreRun:  CommandPreRunFunc,
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, _, apiEndpoint, requestedControlPlane := tptctl_cmd.GetClientContext(cmd)

		switch getWireguardDefinitionVersion {
		case "v0":
		default:
			cli.Error("", errors.New("unrecognized object version"))
			os.Exit(1)
		}
	},
	Short:        "Get wireguard configuration from the system",
	SilenceUsage: true,
	Use:          "wireguard-conf",
}

func init() {
	GetCmd.AddCommand(GetWireguardConfCmd)

	GetWireguardConfCmd.Flags().StringVarP(
		&getWireguardConfVersion,
		"version", "v", "v0", "Version of wireguard configuration object to retrieve. One of: [v0]",
	)
}

// getWireguardConfVersionV0 retrieves the wireguard configuration from the system
func getWireguardConfVersionV0(
	apiClient *http.Client,
	apiEndpoint string,
) (string, error) {
	return "", nil
}
