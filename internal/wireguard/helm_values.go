package wireguard

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

// getHelmValues returns the helm values for the wireguard deployment
func getHelmValues() (map[string]interface{}, error) {
	serverPrivateKey, serverPublicKey, err := generateWireguardKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to generate wireguard server keys: %v", err)
	}

	clientPrivateKey, clientPublicKey, err := generateWireguardKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to generate wireguard client keys: %v", err)
	}

	subnet := "10.0.0.0/24"
	serverIP := "10.0.0.1/24"
	clientIP := "10.0.0.2/32"
	wireguardPort := "51820"

	wg0conf := `[Interface]
Address = %s
ListenPort = %s
MTU = 1420
DNS = 8.8.8.8

# Example peer configuration (uncomment and modify as needed)
[Peer]
PublicKey = %s
AllowedIPs = %s`
	formattedWg0Conf := fmt.Sprintf(wg0conf, serverIP, wireguardPort, clientPublicKey, clientIP)

	iptablesScript := `#!/bin/bash
IPT="/sbin/iptables"

IN_FACE="eth0"                   # NIC connected to the internet
WG_FACE="wg0"                    # WG NIC
SUB_NET="%s"                     # WG IPv4 subnet
WG_PORT="%s"                     # WG udp port

# IPv4 rules #
$IPT -t nat -I POSTROUTING 1 -s $SUB_NET -o $IN_FACE -j MASQUERADE
$IPT -I INPUT 1 -i $WG_FACE -j ACCEPT
$IPT -I FORWARD 1 -i $IN_FACE -o $WG_FACE -j ACCEPT
$IPT -I FORWARD 1 -i $WG_FACE -o $IN_FACE -j ACCEPT
$IPT -I INPUT 1 -i $IN_FACE -p udp --dport $WG_PORT -j ACCEPT`
	formattedIpTablesScript := fmt.Sprintf(iptablesScript, subnet, wireguardPort)

	extraDeploy := []interface{}{
		map[string]interface{}{
			"kind": "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "wg0-config",
				"labels": map[string]interface{}{
					"app": "wg-portal",
				},
			},
			"data": map[string]interface{}{
				"wg0.conf": formattedWg0Conf,
			},
			"apiVersion": "v1",
		},
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "iptables-script",
			},
			"data": map[string]interface{}{
				"add-nat-routing.sh": formattedIpTablesScript,
			},
		},
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name": "wireguard-server-keys",
			},
			"stringData": map[string]interface{}{
				"privateKey": serverPrivateKey,
				"publicKey":  serverPublicKey,
			},
		},
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name": "wireguard-client-keys",
			},
			"stringData": map[string]interface{}{
				"privateKey": clientPrivateKey,
				"publicKey":  clientPublicKey,
			},
		},
	}

	// calculate hash of ConfigMaps, which will trigger a pod restart
	// on changes made to wireguard config
	configMapHash := calculateConfigMapHash(extraDeploy)

	values := map[string]interface{}{
		"service": map[string]interface{}{
			"wireguard": map[string]interface{}{
				"annotations": map[string]interface{}{
					"oci.oraclecloud.com/load-balancer-type": "nlb",
				},
			},
		},
		"podAnnotations": map[string]interface{}{
			"checksum/configmaps": configMapHash,
		},
		"extraDeploy": extraDeploy,
		"volumes": []interface{}{
			map[string]interface{}{
				"name": "wg-config",
				"configMap": map[string]interface{}{
					"name": "wg0-config",
				},
			},
			map[string]interface{}{
				"name": "wireguard-server-keys",
				"secret": map[string]interface{}{
					"secretName":  "wireguard-server-keys",
					"defaultMode": 0400,
				},
			},
			map[string]interface{}{
				"name": "iptables-script",
				"configMap": map[string]interface{}{
					"name":        "iptables-script",
					"defaultMode": 493,
				},
			},
		},
		"volumeMounts": []interface{}{
			map[string]interface{}{
				"mountPath": "/data/wireguard",
				"readOnly":  true,
				"name":      "wg-config",
			},
			map[string]interface{}{
				"mountPath": "/data/wireguard-server-keys",
				"readOnly":  true,
				"name":      "wireguard-server-keys",
			},
		},
		"initContainers": []interface{}{
			map[string]interface{}{
				"command": []interface{}{
					"sh",
					"-c",
					`sysctl -w net.ipv4.conf.all.forwarding=1 &&
sh -c /data/iptables/add-nat-routing.sh &&
wg-quick up /data/wireguard/wg0.conf &&
wg set wg0 private-key /data/wireguard-server-keys/privateKey
`,
				},
				"image":           "ghcr.io/h44z/wg-portal:v2",
				"imagePullPolicy": "IfNotPresent",
				"name":            "network-init",
				"securityContext": map[string]interface{}{
					"capabilities": map[string]interface{}{
						"add": []interface{}{
							"NET_ADMIN",
						},
					},
					"privileged": true,
				},
				"volumeMounts": []interface{}{
					map[string]interface{}{
						"name":      "wg-config",
						"mountPath": "/data/wireguard",
					},
					map[string]interface{}{
						"name":      "wireguard-server-keys",
						"mountPath": "/data/wireguard-server-keys",
					},
					map[string]interface{}{
						"name":      "iptables-script",
						"mountPath": "/data/iptables",
					},
				},
			},
		},
	}

	return values, nil
}

// calculateConfigMapHash calculates a hash of all ConfigMap objects in the extraDeploy list
func calculateConfigMapHash(extraDeploy []interface{}) string {
	var configMaps []map[string]interface{}

	// extract ConfigMap objects from extraDeploy
	for _, obj := range extraDeploy {
		if objMap, ok := obj.(map[string]interface{}); ok {
			if kind, ok := objMap["kind"].(string); ok && kind == "ConfigMap" {
				configMaps = append(configMaps, objMap)
			}
		}
	}

	// convert to JSON for hashing
	jsonBytes, err := json.Marshal(configMaps)
	if err != nil {
		return ""
	}

	// calculate SHA-256 hash
	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:])
}

// generateWireguardKeys returns a Wireguard key pair
func generateWireguardKeys() (string, string, error) {
	// generate a 32-byte private key using crypto/rand
	privateKey := make([]byte, 32)
	_, err := rand.Read(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %v", err)
	}

	// clamp private key for curve25519 (WireGuard requirement)
	privateKey[0] &= 248  // 1. clear lowest 3 bits of first byte
	privateKey[31] &= 127 // 2. clear highest bit of last byte
	privateKey[31] |= 64  // 3. set second highest bit of last byte

	// derive public key from the private key using curve25519
	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, (*[32]byte)(privateKey))

	// encode keys in base64
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKey)
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey[:])

	return privateKeyBase64, publicKeyBase64, nil
}
