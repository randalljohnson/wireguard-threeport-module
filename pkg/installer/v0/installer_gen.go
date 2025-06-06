// generated by 'threeport-sdk gen' - do not edit

package v0

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	api_v0 "github.com/randalljohnson/wireguard-threeport-module/pkg/api/v0"
	tp_api "github.com/threeport/threeport/pkg/api/v0"
	tp_auth "github.com/threeport/threeport/pkg/auth/v0"
	tp_client "github.com/threeport/threeport/pkg/client/v0"
	kube "github.com/threeport/threeport/pkg/kube/v0"
	tp_installer "github.com/threeport/threeport/pkg/threeport-installer/v0"
	util "github.com/threeport/threeport/pkg/util/v0"
	errors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	dynamic "k8s.io/client-go/dynamic"
	"net/http"
)

const (
	ReleaseImageRepo          = "docker.io/rj9317/wireguard-threeport-module"
	DevImageRepo              = "localhost:5001"
	DbInitFilename            = "db.sql"
	DbInitLocation            = "/etc/threeport/db-create"
	defaultNamespace          = "threeport-wireguard"
	defaultThreeportNamespace = "threeport-control-plane"
	apiServerDeployName       = "threeport-wireguard-api-server"
	moduleName                = "randalljohnson.us/wireguard-module-api"
	caSecretName              = "wireguard-controller-ca"
	certSecretName            = "wireguard-controller-cert"
)

// Installer contains the values needed for a module installation.
type Installer struct {
	// dynamice interface client for Kubernetes API
	KubeClient dynamic.Interface

	// Kubernetes API REST mapper
	KubeRestMapper *meta.RESTMapper

	// The Kubernetes namespace to install the module components in.
	ModuleNamespace string

	// The Kubernetes namespace the Threeport control plane is installed in.
	ThreeportNamespace string

	// The container image repository to pull module's API server and
	// controller/s' container images from.
	ControlPlaneImageRepo string

	// The container image tag to use for module's API server and
	// controller/s' container image.
	ControlPlaneImageTag string

	// If true, auth is enabled on Threeport API.
	AuthEnabled bool
}

// NewInstaller returns a wireguard module installer with default values.
func NewInstaller(
	kubeClient dynamic.Interface,
	restMapper *meta.RESTMapper,
) *Installer {
	defaultInstaller := Installer{
		KubeClient:         kubeClient,
		KubeRestMapper:     restMapper,
		ModuleNamespace:    defaultNamespace,
		ThreeportNamespace: defaultThreeportNamespace,
	}

	return &defaultInstaller
}

// InstallWireguardModule installs the controller and API for the wireguard module.
func (i *Installer) InstallWireguardModule() error {
	// create namespace
	var namespace = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": i.ModuleNamespace,
			},
		},
	}

	if _, err := kube.CreateOrUpdateResource(
		namespace,
		i.KubeClient,
		*i.KubeRestMapper,
	); err != nil {
		return fmt.Errorf("failed to create/update wireguard module namespace: %w", err)
	}

	// copy secrets into module namespace
	if err := copySecret(
		i.KubeClient,
		*i.KubeRestMapper,
		"db-root-cert",
		i.ThreeportNamespace,
		i.ModuleNamespace,
	); err != nil {
		return fmt.Errorf("failed to copy secret: %w", err)
	}

	if err := copySecret(
		i.KubeClient,
		*i.KubeRestMapper,
		"db-threeport-cert",
		i.ThreeportNamespace,
		i.ModuleNamespace,
	); err != nil {
		return fmt.Errorf("failed to copy secret: %w", err)
	}

	if err := copySecret(
		i.KubeClient,
		*i.KubeRestMapper,
		"encryption-key",
		i.ThreeportNamespace,
		i.ModuleNamespace,
	); err != nil {
		return fmt.Errorf("failed to copy secret: %w", err)
	}

	if err := copySecret(
		i.KubeClient,
		*i.KubeRestMapper,
		"controller-config",
		i.ThreeportNamespace,
		i.ModuleNamespace,
	); err != nil {
		return fmt.Errorf("failed to copy secret: %w", err)
	}

	if err := copySecret(
		i.KubeClient,
		*i.KubeRestMapper,
		"db-config",
		i.ThreeportNamespace,
		i.ModuleNamespace,
	); err != nil {
		return fmt.Errorf("failed to copy secret: %w", err)
	}

	// create configmap used to initialize API database
	var dbCreateConfig = &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"data": map[string]interface{}{
			"db.sql": "CREATE USER IF NOT EXISTS threeport;\nCREATE DATABASE IF NOT EXISTS threeport_wireguard_api encoding='utf-8';\nGRANT ALL ON DATABASE threeport_wireguard_api TO threeport;",
		},
		"kind": "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      "db-create",
			"namespace": i.ModuleNamespace,
		},
	}}

	if _, err := kube.CreateOrUpdateResource(dbCreateConfig, i.KubeClient, *i.KubeRestMapper); err != nil {
		return fmt.Errorf("failed to create/update wireguard DB initialization configmap: %w", err)
	}

	// install wireguard API server deployment
	apiArgs := []interface{}{"-auto-migrate=true"}
	if !i.AuthEnabled {
		apiArgs = append(apiArgs, "-auth-enabled=false")
	}
	var wireguardApiDeploy = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      apiServerDeployName,
				"namespace": i.ModuleNamespace,
			},
			"spec": map[string]interface{}{
				"replicas": 1,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app.kubernetes.io/name": apiServerDeployName,
					},
				},
				"strategy": map[string]interface{}{
					"rollingUpdate": map[string]interface{}{
						"maxSurge":       "25%",
						"maxUnavailable": "25%",
					},
					"type": "RollingUpdate",
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": nil,
						"labels": map[string]interface{}{
							"app.kubernetes.io/name": apiServerDeployName,
						},
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"args": apiArgs,
								"command": []interface{}{
									"/rest-api",
								},
								"envFrom": []interface{}{
									map[string]interface{}{
										"secretRef": map[string]interface{}{
											"name": "encryption-key",
										},
									},
								},
								"image": fmt.Sprintf(
									"%s/threeport-wireguard-rest-api:%s",
									i.ControlPlaneImageRepo,
									i.ControlPlaneImageTag,
								),
								"imagePullPolicy": "IfNotPresent",
								"name":            "api-server",
								"ports": []interface{}{
									map[string]interface{}{
										"containerPort": 1323,
										"name":          "api",
										"protocol":      "TCP",
									},
								},
								"readinessProbe": map[string]interface{}{
									"failureThreshold": 1,
									"httpGet": map[string]interface{}{
										"path":   "/readyz",
										"port":   8081,
										"scheme": "HTTP",
									},
									"initialDelaySeconds": 1,
									"periodSeconds":       2,
									"successThreshold":    1,
									"timeoutSeconds":      1,
								},
								"volumeMounts": []interface{}{
									map[string]interface{}{
										"mountPath": "/etc/threeport/",
										"name":      "db-config",
									},
									map[string]interface{}{
										"mountPath": "/etc/threeport/db-certs",
										"name":      "db-threeport-cert",
									},
								},
							},
						},
						"initContainers": []interface{}{
							map[string]interface{}{
								"command": []interface{}{
									"bash",
									"-c",
									fmt.Sprintf("cockroach sql --certs-dir=/etc/threeport/db-certs --host crdb.%s.svc.cluster.local --port 26257 -f /etc/threeport/db-create/db.sql", i.ThreeportNamespace),
								},
								"image":           "cockroachdb/cockroach:v23.1.14",
								"imagePullPolicy": "IfNotPresent",
								"name":            "db-init",
								"volumeMounts": []interface{}{
									map[string]interface{}{
										"mountPath": "/etc/threeport/db-create",
										"name":      "db-create",
									},
									map[string]interface{}{
										"mountPath": "/etc/threeport/db-certs",
										"name":      "db-root-cert",
									},
								},
							},
							map[string]interface{}{
								"args": []interface{}{
									"-env-file=/etc/threeport/env",
									"up",
								},
								"command": []interface{}{
									"/database-migrator",
								},
								"image": fmt.Sprintf(
									"%s/threeport-wireguard-database-migrator:%s",
									i.ControlPlaneImageRepo,
									i.ControlPlaneImageTag,
								),
								"imagePullPolicy": "IfNotPresent",
								"name":            "database-migrator",
								"volumeMounts": []interface{}{
									map[string]interface{}{
										"mountPath": "/etc/threeport/",
										"name":      "db-config",
									},
									map[string]interface{}{
										"mountPath": "/etc/threeport/db-certs",
										"name":      "db-threeport-cert",
									},
								},
							},
						},
						"restartPolicy":                 "Always",
						"terminationGracePeriodSeconds": 30,
						"volumes": []interface{}{
							map[string]interface{}{
								"name": "db-root-cert",
								"secret": map[string]interface{}{
									"defaultMode": 420,
									"secretName":  "db-root-cert",
								},
							},
							map[string]interface{}{
								"name": "db-threeport-cert",
								"secret": map[string]interface{}{
									"defaultMode": 420,
									"secretName":  "db-threeport-cert",
								},
							},
							map[string]interface{}{
								"name": "db-config",
								"secret": map[string]interface{}{
									"defaultMode": 420,
									"secretName":  "db-config",
								},
							},
							map[string]interface{}{
								"configMap": map[string]interface{}{
									"defaultMode": 420,
									"name":        "db-create",
								},
								"name": "db-create",
							},
						},
					},
				},
			},
		},
	}

	if _, err := kube.CreateOrUpdateResource(wireguardApiDeploy, i.KubeClient, *i.KubeRestMapper); err != nil {
		return fmt.Errorf("failed to create/update wireguard API deployment: %w", err)
	}

	// install wireguard API server service
	var wireguardApiService = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"app.kubernetes.io/name": apiServerDeployName,
				},
				"name":      apiServerDeployName,
				"namespace": i.ModuleNamespace,
			},
			"spec": map[string]interface{}{
				"ports": []interface{}{
					map[string]interface{}{
						"name":       "http",
						"port":       80,
						"protocol":   "TCP",
						"targetPort": 1323,
					},
				},
				"selector": map[string]interface{}{
					"app.kubernetes.io/name": apiServerDeployName,
				},
			},
		},
	}
	if _, err := kube.CreateOrUpdateResource(wireguardApiService, i.KubeClient, *i.KubeRestMapper); err != nil {
		return fmt.Errorf("failed to create/updated wireguard API service: %w", err)
	}

	// install wireguard controller/s
	controllerVolumes := []interface{}{}
	controllerVolumeMounts := []interface{}{}
	if i.AuthEnabled {
		// if auth is enabled, get the Threeport API server CA cert and key from
		// the Kubernetes cluster.
		caCert, caKey, err := i.getApiCa()
		if err != nil {
			return fmt.Errorf("failed to retrieve Threeport API CA cert and key: %w", err)
		}

		// load the cert and key
		x509CaCert, rsaCaKey, err := loadApiCa(caCert, caKey)
		if err != nil {
			return fmt.Errorf("failed to load Threeport API CA cert and key: %w", err)
		}

		// generate a cert and key for the controller that needs to connect to
		// the Threeport API
		clientCert, clientKey, err := tp_auth.GenerateCertificate(x509CaCert, rsaCaKey, "localhost")
		if err != nil {
			return fmt.Errorf("failed to generate client cert and key for wireguard controller: %w", err)
		}

		// create secrets for controller to load credentials from
		if err := i.createAuthCertSecrets(string(caCert), clientCert, clientKey); err != nil {
			return fmt.Errorf("failed to create client auth certs for wireguard controller: %w", err)
		}

		// add the volumes and volume mounts for deployment manifest
		controllerVolumes = getVolumes()
		controllerVolumeMounts = getVolumeMounts()
	}

	// set auth enabled flag if auth not enabled (default is true)
	controllerArgs := []interface{}{}
	if !i.AuthEnabled {
		controllerArgs = append(controllerArgs, "-auth-enabled=false")
	}

	var WireguardControllerDeploy = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "threeport-wireguard-wireguard-controller",
				"namespace": i.ModuleNamespace,
			},
			"spec": map[string]interface{}{
				"replicas": 1,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app.kubernetes.io/name": "threeport-wireguard-wireguard-controller",
					},
				},
				"strategy": map[string]interface{}{
					"rollingUpdate": map[string]interface{}{
						"maxSurge":       "25%",
						"maxUnavailable": "25%",
					},
					"type": "RollingUpdate",
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app.kubernetes.io/name": "threeport-wireguard-wireguard-controller",
						},
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"args": controllerArgs,
								"command": []interface{}{
									"/wireguard-controller",
								},
								"envFrom": []interface{}{
									map[string]interface{}{
										"secretRef": map[string]interface{}{
											"name": "controller-config",
										},
									},
									map[string]interface{}{
										"secretRef": map[string]interface{}{
											"name": "encryption-key",
										},
									},
								},
								"image": fmt.Sprintf(
									"%s/threeport-wireguard-wireguard-controller:%s",
									i.ControlPlaneImageRepo,
									i.ControlPlaneImageTag,
								),
								"imagePullPolicy": "IfNotPresent",
								"name":            "wireguard-wireguard-controller",
								"readinessProbe": map[string]interface{}{
									"failureThreshold": 1,
									"httpGet": map[string]interface{}{
										"path":   "/readyz",
										"port":   8081,
										"scheme": "HTTP",
									},
									"initialDelaySeconds": 1,
									"periodSeconds":       2,
									"successThreshold":    1,
									"timeoutSeconds":      1,
								},
								"volumeMounts": controllerVolumeMounts,
							},
						},
						"restartPolicy":                 "Always",
						"terminationGracePeriodSeconds": 30,
						"volumes":                       controllerVolumes,
					},
				},
			},
		},
	}

	if _, err := kube.CreateOrUpdateResource(WireguardControllerDeploy, i.KubeClient, *i.KubeRestMapper); err != nil {
		return fmt.Errorf("failed to create/update wireguard controller deployment: %w", err)
	}

	return nil
}

// copySecret copies a secret from one namespace to another.  The function
// returns without error if the secret already exists in the target namespace.
func copySecret(
	dynamicClient dynamic.Interface,
	restMapper meta.RESTMapper,
	secretName string,
	sourceNamespace string,
	targetNamespace string,
) error {
	secretGVR := schema.GroupVersionResource{
		Group:    "",
		Resource: "secrets",
		Version:  "v1",
	}
	secretGK := schema.GroupKind{
		Group: "",
		Kind:  "Secret",
	}

	mapping, err := restMapper.RESTMapping(secretGK, secretGVR.Version)
	if err != nil {
		return fmt.Errorf("failed to get RESTMapping for Secret resource: %w", err)
	}

	targetSecretResource := dynamicClient.Resource(mapping.Resource).Namespace(targetNamespace)
	_, err = targetSecretResource.Get(context.TODO(), secretName, metav1.GetOptions{})
	if err == nil {
		// secret already exists, return nil
		return nil
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf(
			"failed to check if Secret '%s' exists in namespace '%s': %w",
			secretName,
			targetNamespace,
			err,
		)
	}

	secretResource := dynamicClient.Resource(mapping.Resource).Namespace(sourceNamespace)
	secret, err := secretResource.Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf(
			"failed to get Secret '%s' from namespace '%s': %w",
			secretName,
			sourceNamespace,
			err,
		)
	}

	secret.SetNamespace(targetNamespace)
	secret.SetResourceVersion("")
	secret.SetUID("")
	secret.SetSelfLink("")
	secret.SetCreationTimestamp(metav1.Time{})
	secret.SetManagedFields(nil)

	_, err = targetSecretResource.Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create/update Secret in namespace '%s': %w", targetNamespace, err)
	}

	return nil
}

// RegisterWireguardModule calls the Threeport API to register the module
// API so that module object requests are proxied to the Wireguard module
// API.
func (i *Installer) RegisterWireguardModule(
	apiClient *http.Client,
	apiAddr string,
) error {
	// check to see if module is already registered
	var existingModApi *tp_api.ModuleApi
	existingModApi, _ = tp_client.GetModuleApiByName(apiClient, apiAddr, moduleName)
	if existingModApi.ID == nil {
		// register the module in the Threeport API
		moduleApi := tp_api.ModuleApi{
			Endpoint: util.Ptr(fmt.Sprintf("%s.%s.svc.cluster.local", apiServerDeployName, defaultNamespace)),
			Name:     util.Ptr(moduleName),
		}
		createdModApi, err := tp_client.CreateModuleApi(apiClient, apiAddr, &moduleApi)
		if err != nil {
			return fmt.Errorf("failed to create module API object in Threeport API: %w", err)
		}
		existingModApi = createdModApi
	}

	// add all the paths to the registered module if they don't already exist
	allRoutePaths := []string{
		api_v0.PathWireguardDefinitionVersions,
		api_v0.PathWireguardDefinitions,
		api_v0.PathWireguardInstanceVersions,
		api_v0.PathWireguardInstances,
	}
	for _, path := range allRoutePaths {
		// check to see if route path exists
		query := fmt.Sprintf("path=%s&moduleapiid=%d", path, *existingModApi.ID)
		existingRoutes, err := tp_client.GetModuleApiRoutesByQueryString(apiClient, apiAddr, query)
		if err != nil {
			return fmt.Errorf("failed to check for existing route path %s: %w", path, err)
		}
		if len(*existingRoutes) == 0 {
			// route path doesn't exist - create it
			route := tp_api.ModuleApiRoute{
				ModuleApiID: existingModApi.ID,
				Path:        &path,
			}
			_, err := tp_client.CreateModuleApiRoute(apiClient, apiAddr, &route)
			if err != nil {
				return fmt.Errorf("failed to create route with path %s in Threeport API: %w", path, err)
			}
		}
	}

	return nil
}

// getApiCa gets the Threeport API CA cert secret from the Kubernetes cluster
// and returns the base 64 decoded string value for the CA cert and key.
func (i *Installer) getApiCa() ([]byte, []byte, error) {
	// get secret resource
	apiCaSecret, err := kube.GetResource(
		"core",
		"v1",
		"Secret",
		i.ThreeportNamespace,
		tp_installer.ThreeportApiCaSecret,
		i.KubeClient,
		*i.KubeRestMapper,
	)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("failed to get Threeport API CA secret from Kubernetes cluster: %w", err)
	}

	// retrieve 'data' field
	data, found, err := unstructured.NestedMap(apiCaSecret.Object, "data")
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("failed to retrieve 'data' field: %w", err)
	}
	if !found {
		return []byte{}, []byte{}, fmt.Errorf("'data' field not found in the secret")
	}

	// extract and decode tls.crt
	tlsCrtBase64, found := data["tls.crt"].(string)
	if !found {
		return []byte{}, []byte{}, fmt.Errorf("'tls.crt' not found in the secret data")
	}
	tlsCrtBytes, err := base64.StdEncoding.DecodeString(tlsCrtBase64)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("failed to decode 'tls.crt': %w", err)
	}

	// extract and decode tls.key
	tlsKeyBase64, found := data["tls.key"].(string)
	if !found {
		return []byte{}, []byte{}, fmt.Errorf("'tls.key' not found in the secret data")
	}
	tlsKeyBytes, err := base64.StdEncoding.DecodeString(tlsKeyBase64)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("failed to decode 'tls.key': %w", err)
	}

	return tlsCrtBytes, tlsKeyBytes, nil
}

// loadApiCa takes the PEM encoded CA cert and key as strings and returns the
// x509.Certificate and rsa.PrivateKey objects.
func loadApiCa(caCertPem, caKeyPem []byte) (*x509.Certificate, *rsa.PrivateKey, error) {
	// decode PEM to extract the certificate
	block, _ := pem.Decode(caCertPem)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, nil, fmt.Errorf("failed to decode CA certificate PEM")
	}

	// Parse the certificate
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// decode PEM to extract the private key
	block, _ = pem.Decode(caKeyPem)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, nil, fmt.Errorf("failed to decode CA private key PEM")
	}

	// Parse the RSA private key
	caPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA private key: %w", err)
	}

	return caCert, caPrivateKey, nil
}

// createAuthCertSecrets creates the Kubernetes secrets needed for a controller
// to connect to the Threeport API.
func (i *Installer) createAuthCertSecrets(caCert, clientCert, clientKey string) error {
	var caCertSecret = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":      caSecretName,
				"namespace": i.ModuleNamespace,
			},
			"stringData": map[string]interface{}{
				"tls.crt": caCert,
			},
		},
	}
	if _, err := kube.CreateOrUpdateResource(caCertSecret, i.KubeClient, *i.KubeRestMapper); err != nil {
		return fmt.Errorf("failed to create/update wireguard CA cert secret: %w", err)
	}

	var clientCertSecret = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":      certSecretName,
				"namespace": i.ModuleNamespace,
			},
			"stringData": map[string]interface{}{
				"tls.crt": clientCert,
				"tls.key": clientKey,
			},
		},
	}
	if _, err := kube.CreateOrUpdateResource(clientCertSecret, i.KubeClient, *i.KubeRestMapper); err != nil {
		return fmt.Errorf("failed to create/update wireguard client cert secret: %w", err)
	}

	return nil
}

// getVolumes returns the volumes for the CA and client certs needed for a
// controller to authenticate to the Threeport API.
func getVolumes() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"name": caSecretName,
			"secret": map[string]interface{}{
				"secretName": caSecretName,
			},
		},
		map[string]interface{}{
			"name": certSecretName,
			"secret": map[string]interface{}{
				"secretName": certSecretName,
			},
		},
	}
}

// getVolumeMounts returns the volume mounts for the CA and client certs needed
// for a controller to authenticate to the Threeport API.
func getVolumeMounts() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"mountPath": "/etc/threeport/ca",
			"name":      caSecretName,
		},
		map[string]interface{}{
			"mountPath": "/etc/threeport/cert",
			"name":      certSecretName,
		},
	}
}
