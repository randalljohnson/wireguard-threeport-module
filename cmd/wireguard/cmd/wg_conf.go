package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/randalljohnson/wireguard-threeport-module/internal/wireguard"
	wg_client "github.com/randalljohnson/wireguard-threeport-module/pkg/client/v0"
	cobra "github.com/spf13/cobra"
	tptctl_cmd "github.com/threeport/threeport/cmd/tptctl/cmd"
	cli "github.com/threeport/threeport/pkg/cli/v0"
	tptctl_config "github.com/threeport/threeport/pkg/config/v0"
	kube "github.com/threeport/threeport/pkg/kube/v0"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var getWireguardConfigVersion string
var wireguardInstanceName string

// GetWireguardConfigCmd represents the wireguard-definition command
var GetWireguardConfigCmd = &cobra.Command{
	Example: "  tptctl wireguard get wireguard-config",
	Long:    "Get wireguard configuration from the system.",
	PreRun:  CommandPreRunFunc,
	Run: func(cmd *cobra.Command, args []string) {

		apiClient, _, apiEndpoint, requestedControlPlane := tptctl_cmd.GetClientContext(cmd)

		// get encryption key from threeport config
		threeportConfig, requestedControlPlane, err := tptctl_config.GetThreeportConfig(cliArgs.ControlPlaneName)
		if err != nil {
			cli.Error("failed to get threeport config: %w", err)
			os.Exit(1)
		}

		// get encryption key
		encryptionKey, err := threeportConfig.GetThreeportEncryptionKey(requestedControlPlane)
		if err != nil {
			cli.Error("failed to get encryption key from threeport config: %w", err)
			os.Exit(1)
		}

		switch getWireguardConfigVersion {
		case "v0":
			wireguardConfig, err := getWireguardConfigVersionV0(apiClient, apiEndpoint, encryptionKey)
			if err != nil {
				cli.Error("failed to retrieve wireguard configuration", err)
				os.Exit(1)
			}

			fmt.Println(wireguardConfig)
		default:
			cli.Error("", errors.New("unrecognized object version"))
			os.Exit(1)
		}
	},
	Short:        "Get wireguard configuration from the system",
	SilenceUsage: true,
	Use:          "wireguard-config",
}

func init() {
	GetCmd.AddCommand(GetWireguardConfigCmd)

	GetWireguardConfigCmd.Flags().StringVarP(
		&getWireguardConfigVersion,
		"version", "v", "v0", "Version of wireguard configuration object to retrieve. One of: [v0]",
	)
	GetWireguardConfigCmd.Flags().StringVarP(
		&wireguardInstanceName,
		"name", "n", "", "Name of wireguard instance to retrieve configuration for.",
	)
	GetWireguardConfigCmd.MarkFlagRequired("name")
}

// getWireguardConfigVersionV0 retrieves the wireguard configuration from the system
func getWireguardConfigVersionV0(
	apiClient *http.Client,
	apiEndpoint string,
	encryptionKey string,
) (string, error) {

	// get wireguard instance
	wireguardInstance, err := wg_client.GetWireguardInstanceByName(
		apiClient,
		apiEndpoint,
		wireguardInstanceName,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get wireguard instance: %w", err)
	}

	// get wireguard service
	wireguardService, err := wireguard.GetWireguardService(
		apiClient,
		apiEndpoint,
		encryptionKey,
		wireguardInstance,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get wireguard service: %w", err)
	}

	// get wireguard loadbalancer ip address
	ingressSlice, found, err := unstructured.NestedSlice(
		wireguardService.Object,
		"status",
		"loadBalancer",
		"ingress",
	)
	switch {
	case err != nil:
		return "", fmt.Errorf("failed to get wireguard loadbalancer ip address: %w", err)
	case !found:
		return "", fmt.Errorf("failed to find wireguard loadbalancer ip address")
	}

	// get public-facing ip address
	var publicFacingIpAddress string
	for _, ingress := range ingressSlice {
		// convert ingress to map
		ingressMap, ok := ingress.(map[string]interface{})
		if !ok {
			continue
		}

		// get ip address
		ip, found, err := unstructured.NestedString(ingressMap, "ip")
		if err != nil || !found {
			continue
		}

		// parse and validate ip address
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			continue
		}

		// return if not private
		if !parsedIP.IsPrivate() {
			publicFacingIpAddress = ip
			break
		}
	}

	// get wireguard private key secrets
	// get wireguard kubernetes runtime instance
	kubernetesRuntimeInstance, err := wireguard.GetWireguardKubernetesRuntimeInstance(
		apiClient,
		apiEndpoint,
		wireguardInstance,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get wireguard kubernetes runtime instance: %w", err)
	}

	// get wireguard secrets
	kubeClient, _, err := kube.GetClient(
		kubernetesRuntimeInstance,
		false,
		apiClient,
		apiEndpoint,
		encryptionKey,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get kubernetes client: %w", err)
	}

	// get wireguard server secret
	serverSecret, err := getWireguardSecret(
		kubeClient,
		"wireguard-server-keys",
		*wireguardInstance.Name,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get wireguard server secret: %w", err)
	}

	// get wireguard client secret
	clientSecret, err := getWireguardSecret(
		kubeClient,
		"wireguard-client-keys",
		*wireguardInstance.Name,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get wireguard client secret: %w", err)
	}

	// get wireguard server public key
	serverPublicKey, err := getWireguardSecretData(serverSecret, "publicKey")
	if err != nil {
		return "", fmt.Errorf("failed to get wireguard server private and public keys: %w", err)
	}

	// get wireguard client private key
	clientPrivateKey, err := getWireguardSecretData(clientSecret, "privateKey")
	if err != nil {
		return "", fmt.Errorf("failed to get wireguard client private and public keys: %w", err)
	}

	return fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = 10.0.0.2/24
DNS = 8.8.8.8

[Peer]
PublicKey = %s
AllowedIPs = 0.0.0.0/0
Endpoint = %s:51820
PersistentKeepalive = 25`,
		clientPrivateKey,
		serverPublicKey,
		publicFacingIpAddress,
	), nil
}

// getWireguardSecret returns a secret with a given name
func getWireguardSecret(
	kubeClient dynamic.Interface,
	secretName string,
	secretNamespace string,
) (unstructured.Unstructured, error) {
	// list wireguard secrets
	secretsList, err := kubeClient.Resource(
		schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "secrets",
		},
	).Namespace(secretNamespace).List(
		context.Background(),
		metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.name=%s", secretName),
		},
	)
	switch {
	case err != nil:
		return unstructured.Unstructured{}, fmt.Errorf("failed to list secrets: %w", err)
	case len(secretsList.Items) == 0:
		return unstructured.Unstructured{}, fmt.Errorf("failed to find wireguard secrets")
	case len(secretsList.Items) > 1:
		return unstructured.Unstructured{}, fmt.Errorf("found multiple wireguard secrets")
	}

	// return secret
	return secretsList.Items[0], nil
}

// getWireguardSecretData returns the private and public keys from a wireguard secret
func getWireguardSecretData(
	secret unstructured.Unstructured,
	key string,
) (string, error) {
	data, found, err := unstructured.NestedMap(secret.Object, "data")
	if err != nil || !found {
		return "", fmt.Errorf("failed to get secret data: %w", err)
	}

	valueBase64, ok := data[key].(string)
	if !ok {
		return "", fmt.Errorf("privateKey not found in secret")
	}

	value, err := base64.StdEncoding.DecodeString(valueBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode privateKey: %w", err)
	}

	return string(value), nil
}
