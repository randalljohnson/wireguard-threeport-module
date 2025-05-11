package wireguard

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// calculateConfigMapHash calculates a hash of all ConfigMap objects in the extraDeploy list
func calculateConfigMapHash(extraDeploy []interface{}) string {
	var configMaps []map[string]interface{}

	// Extract ConfigMap objects from extraDeploy
	for _, obj := range extraDeploy {
		if objMap, ok := obj.(map[string]interface{}); ok {
			if kind, ok := objMap["kind"].(string); ok && kind == "ConfigMap" {
				configMaps = append(configMaps, objMap)
			}
		}
	}

	// Sort configMaps by name to ensure consistent hashing
	// Convert to JSON for hashing
	jsonBytes, err := json.Marshal(configMaps)
	if err != nil {
		return ""
	}

	// Calculate SHA-256 hash
	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:])
}

func getHelmValues() map[string]interface{} {
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
				"wg0.conf": `[Interface]
Address = 10.11.12.1/24 ## TODO: should be a variable, not possible without orchestration layer
ListenPort = 51820
MTU = 1420
DNS = 8.8.8.8

# Example peer configuration (uncomment and modify as needed)
[Peer]
PublicKey = fTo/2gZLB3m7Y7CfIEK5TeZ2R8zERxs5VXB/MtcEyXI=
AllowedIPs = 10.11.12.2/32
`,
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
				"add-nat-routing.sh": `#!/bin/bash
IPT="/sbin/iptables"

IN_FACE="eth0"                   # NIC connected to the internet
WG_FACE="wg0"                    # WG NIC
SUB_NET="10.11.12.0/24"          # WG IPv4 sub/net aka CIDR
WG_PORT="51820"                  # WG udp port

# IPv4 rules #
$IPT -t nat -I POSTROUTING 1 -s $SUB_NET -o $IN_FACE -j MASQUERADE
$IPT -I INPUT 1 -i $WG_FACE -j ACCEPT
$IPT -I FORWARD 1 -i $IN_FACE -o $WG_FACE -j ACCEPT
$IPT -I FORWARD 1 -i $WG_FACE -o $IN_FACE -j ACCEPT
$IPT -I INPUT 1 -i $IN_FACE -p udp --dport $WG_PORT -j ACCEPT
`,
			},
		},
	}

	// Calculate hash of ConfigMaps
	configMapHash := calculateConfigMapHash(extraDeploy)

	return map[string]interface{}{
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
				"name": "wireguard-private-key",
				"configMap": map[string]interface{}{
					"name":        "wireguard-private-key",
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
				"mountPath": "/data/wireguard",
				"readOnly":  true,
				"name":      "wireguard-private-key",
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
wg load /data/wireguard/privatekey
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
						"name":      "wireguard-private-key",
						"mountPath": "/data/wireguard",
					},
					map[string]interface{}{
						"name":      "iptables-script",
						"mountPath": "/data/iptables",
					},
				},
			},
		},
	}
}
