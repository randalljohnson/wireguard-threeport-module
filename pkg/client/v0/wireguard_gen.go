// generated by 'threeport-sdk gen' - do not edit

package v0

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	v0 "github.com/randalljohnson/wireguard-threeport-module/pkg/api/v0"
	tpclient_lib "github.com/threeport/threeport/pkg/client/lib/v0"
	tputil "github.com/threeport/threeport/pkg/util/v0"
	"net/http"
)

// GetWireguardDefinitions fetches all wireguard definitions.
// TODO: implement pagination
func GetWireguardDefinitions(apiClient *http.Client, apiAddr string) (*[]v0.WireguardDefinition, error) {
	var wireguardDefinitions []v0.WireguardDefinition

	response, err := tpclient_lib.GetResponse(
		apiClient,
		fmt.Sprintf("%s%s", apiAddr, v0.PathWireguardDefinitions),
		http.MethodGet,
		new(bytes.Buffer),
		map[string]string{},
		http.StatusOK,
	)
	if err != nil {
		return &wireguardDefinitions, fmt.Errorf("call to threeport API returned unexpected response: %w", err)
	}

	jsonData, err := json.Marshal(response.Data)
	if err != nil {
		return &wireguardDefinitions, fmt.Errorf("failed to marshal response data from threeport API: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&wireguardDefinitions); err != nil {
		return nil, fmt.Errorf("failed to decode object in response data from threeport API: %w", err)
	}

	return &wireguardDefinitions, nil
}

// GetWireguardDefinitionByID fetches a wireguard definition by ID.
func GetWireguardDefinitionByID(apiClient *http.Client, apiAddr string, id uint) (*v0.WireguardDefinition, error) {
	var wireguardDefinition v0.WireguardDefinition

	response, err := tpclient_lib.GetResponse(
		apiClient,
		fmt.Sprintf("%s%s/%d", apiAddr, v0.PathWireguardDefinitions, id),
		http.MethodGet,
		new(bytes.Buffer),
		map[string]string{},
		http.StatusOK,
	)
	if err != nil {
		return &wireguardDefinition, fmt.Errorf("call to threeport API returned unexpected response: %w", err)
	}

	jsonData, err := json.Marshal(response.Data[0])
	if err != nil {
		return &wireguardDefinition, fmt.Errorf("failed to marshal response data from threeport API: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&wireguardDefinition); err != nil {
		return nil, fmt.Errorf("failed to decode object in response data from threeport API: %w", err)
	}

	return &wireguardDefinition, nil
}

// GetWireguardDefinitionsByQueryString fetches wireguard definitions by provided query string.
func GetWireguardDefinitionsByQueryString(apiClient *http.Client, apiAddr string, queryString string) (*[]v0.WireguardDefinition, error) {
	var wireguardDefinitions []v0.WireguardDefinition

	response, err := tpclient_lib.GetResponse(
		apiClient,
		fmt.Sprintf("%s%s?%s", apiAddr, v0.PathWireguardDefinitions, queryString),
		http.MethodGet,
		new(bytes.Buffer),
		map[string]string{},
		http.StatusOK,
	)
	if err != nil {
		return &wireguardDefinitions, fmt.Errorf("call to threeport API returned unexpected response: %w", err)
	}

	jsonData, err := json.Marshal(response.Data)
	if err != nil {
		return &wireguardDefinitions, fmt.Errorf("failed to marshal response data from threeport API: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&wireguardDefinitions); err != nil {
		return nil, fmt.Errorf("failed to decode object in response data from threeport API: %w", err)
	}

	return &wireguardDefinitions, nil
}

// GetWireguardDefinitionByName fetches a wireguard definition by name.
func GetWireguardDefinitionByName(apiClient *http.Client, apiAddr, name string) (*v0.WireguardDefinition, error) {
	var wireguardDefinitions []v0.WireguardDefinition

	response, err := tpclient_lib.GetResponse(
		apiClient,
		fmt.Sprintf("%s%s?name=%s", apiAddr, v0.PathWireguardDefinitions, name),
		http.MethodGet,
		new(bytes.Buffer),
		map[string]string{},
		http.StatusOK,
	)
	if err != nil {
		return &v0.WireguardDefinition{}, fmt.Errorf("call to threeport API returned unexpected response: %w", err)
	}

	jsonData, err := json.Marshal(response.Data)
	if err != nil {
		return &v0.WireguardDefinition{}, fmt.Errorf("failed to marshal response data from threeport API: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&wireguardDefinitions); err != nil {
		return nil, fmt.Errorf("failed to decode object in response data from threeport API: %w", err)
	}

	switch {
	case len(wireguardDefinitions) < 1:
		return &v0.WireguardDefinition{}, errors.New(fmt.Sprintf("no wireguard definition with name %s", name))
	case len(wireguardDefinitions) > 1:
		return &v0.WireguardDefinition{}, errors.New(fmt.Sprintf("more than one wireguard definition with name %s returned", name))
	}

	return &wireguardDefinitions[0], nil
}

// CreateWireguardDefinition creates a new wireguard definition.
func CreateWireguardDefinition(apiClient *http.Client, apiAddr string, wireguardDefinition *v0.WireguardDefinition) (*v0.WireguardDefinition, error) {
	tpclient_lib.ReplaceAssociatedObjectsWithNil(wireguardDefinition)
	jsonWireguardDefinition, err := tputil.MarshalObject(wireguardDefinition)
	if err != nil {
		return wireguardDefinition, fmt.Errorf("failed to marshal provided object to JSON: %w", err)
	}

	response, err := tpclient_lib.GetResponse(
		apiClient,
		fmt.Sprintf("%s%s", apiAddr, v0.PathWireguardDefinitions),
		http.MethodPost,
		bytes.NewBuffer(jsonWireguardDefinition),
		map[string]string{},
		http.StatusCreated,
	)
	if err != nil {
		return wireguardDefinition, fmt.Errorf("call to threeport API returned unexpected response: %w", err)
	}

	jsonData, err := json.Marshal(response.Data[0])
	if err != nil {
		return wireguardDefinition, fmt.Errorf("failed to marshal response data from threeport API: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&wireguardDefinition); err != nil {
		return nil, fmt.Errorf("failed to decode object in response data from threeport API: %w", err)
	}

	return wireguardDefinition, nil
}

// UpdateWireguardDefinition updates a wireguard definition.
func UpdateWireguardDefinition(apiClient *http.Client, apiAddr string, wireguardDefinition *v0.WireguardDefinition) (*v0.WireguardDefinition, error) {
	tpclient_lib.ReplaceAssociatedObjectsWithNil(wireguardDefinition)
	// capture the object ID, make a copy of the object, then remove fields that
	// cannot be updated in the API
	wireguardDefinitionID := *wireguardDefinition.ID
	payloadWireguardDefinition := *wireguardDefinition
	payloadWireguardDefinition.ID = nil
	payloadWireguardDefinition.CreatedAt = nil
	payloadWireguardDefinition.UpdatedAt = nil

	jsonWireguardDefinition, err := tputil.MarshalObject(payloadWireguardDefinition)
	if err != nil {
		return wireguardDefinition, fmt.Errorf("failed to marshal provided object to JSON: %w", err)
	}

	response, err := tpclient_lib.GetResponse(
		apiClient,
		fmt.Sprintf("%s%s/%d", apiAddr, v0.PathWireguardDefinitions, wireguardDefinitionID),
		http.MethodPatch,
		bytes.NewBuffer(jsonWireguardDefinition),
		map[string]string{},
		http.StatusOK,
	)
	if err != nil {
		return wireguardDefinition, fmt.Errorf("call to threeport API returned unexpected response: %w", err)
	}

	jsonData, err := json.Marshal(response.Data[0])
	if err != nil {
		return wireguardDefinition, fmt.Errorf("failed to marshal response data from threeport API: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&payloadWireguardDefinition); err != nil {
		return nil, fmt.Errorf("failed to decode object in response data from threeport API: %w", err)
	}

	payloadWireguardDefinition.ID = &wireguardDefinitionID
	return &payloadWireguardDefinition, nil
}

// DeleteWireguardDefinition deletes a wireguard definition by ID.
func DeleteWireguardDefinition(apiClient *http.Client, apiAddr string, id uint) (*v0.WireguardDefinition, error) {
	var wireguardDefinition v0.WireguardDefinition

	response, err := tpclient_lib.GetResponse(
		apiClient,
		fmt.Sprintf("%s%s/%d", apiAddr, v0.PathWireguardDefinitions, id),
		http.MethodDelete,
		new(bytes.Buffer),
		map[string]string{},
		http.StatusOK,
	)
	if err != nil {
		return &wireguardDefinition, fmt.Errorf("call to threeport API returned unexpected response: %w", err)
	}

	jsonData, err := json.Marshal(response.Data[0])
	if err != nil {
		return &wireguardDefinition, fmt.Errorf("failed to marshal response data from threeport API: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&wireguardDefinition); err != nil {
		return nil, fmt.Errorf("failed to decode object in response data from threeport API: %w", err)
	}

	return &wireguardDefinition, nil
}

// GetWireguardInstances fetches all wireguard instances.
// TODO: implement pagination
func GetWireguardInstances(apiClient *http.Client, apiAddr string) (*[]v0.WireguardInstance, error) {
	var wireguardInstances []v0.WireguardInstance

	response, err := tpclient_lib.GetResponse(
		apiClient,
		fmt.Sprintf("%s%s", apiAddr, v0.PathWireguardInstances),
		http.MethodGet,
		new(bytes.Buffer),
		map[string]string{},
		http.StatusOK,
	)
	if err != nil {
		return &wireguardInstances, fmt.Errorf("call to threeport API returned unexpected response: %w", err)
	}

	jsonData, err := json.Marshal(response.Data)
	if err != nil {
		return &wireguardInstances, fmt.Errorf("failed to marshal response data from threeport API: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&wireguardInstances); err != nil {
		return nil, fmt.Errorf("failed to decode object in response data from threeport API: %w", err)
	}

	return &wireguardInstances, nil
}

// GetWireguardInstanceByID fetches a wireguard instance by ID.
func GetWireguardInstanceByID(apiClient *http.Client, apiAddr string, id uint) (*v0.WireguardInstance, error) {
	var wireguardInstance v0.WireguardInstance

	response, err := tpclient_lib.GetResponse(
		apiClient,
		fmt.Sprintf("%s%s/%d", apiAddr, v0.PathWireguardInstances, id),
		http.MethodGet,
		new(bytes.Buffer),
		map[string]string{},
		http.StatusOK,
	)
	if err != nil {
		return &wireguardInstance, fmt.Errorf("call to threeport API returned unexpected response: %w", err)
	}

	jsonData, err := json.Marshal(response.Data[0])
	if err != nil {
		return &wireguardInstance, fmt.Errorf("failed to marshal response data from threeport API: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&wireguardInstance); err != nil {
		return nil, fmt.Errorf("failed to decode object in response data from threeport API: %w", err)
	}

	return &wireguardInstance, nil
}

// GetWireguardInstancesByQueryString fetches wireguard instances by provided query string.
func GetWireguardInstancesByQueryString(apiClient *http.Client, apiAddr string, queryString string) (*[]v0.WireguardInstance, error) {
	var wireguardInstances []v0.WireguardInstance

	response, err := tpclient_lib.GetResponse(
		apiClient,
		fmt.Sprintf("%s%s?%s", apiAddr, v0.PathWireguardInstances, queryString),
		http.MethodGet,
		new(bytes.Buffer),
		map[string]string{},
		http.StatusOK,
	)
	if err != nil {
		return &wireguardInstances, fmt.Errorf("call to threeport API returned unexpected response: %w", err)
	}

	jsonData, err := json.Marshal(response.Data)
	if err != nil {
		return &wireguardInstances, fmt.Errorf("failed to marshal response data from threeport API: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&wireguardInstances); err != nil {
		return nil, fmt.Errorf("failed to decode object in response data from threeport API: %w", err)
	}

	return &wireguardInstances, nil
}

// GetWireguardInstanceByName fetches a wireguard instance by name.
func GetWireguardInstanceByName(apiClient *http.Client, apiAddr, name string) (*v0.WireguardInstance, error) {
	var wireguardInstances []v0.WireguardInstance

	response, err := tpclient_lib.GetResponse(
		apiClient,
		fmt.Sprintf("%s%s?name=%s", apiAddr, v0.PathWireguardInstances, name),
		http.MethodGet,
		new(bytes.Buffer),
		map[string]string{},
		http.StatusOK,
	)
	if err != nil {
		return &v0.WireguardInstance{}, fmt.Errorf("call to threeport API returned unexpected response: %w", err)
	}

	jsonData, err := json.Marshal(response.Data)
	if err != nil {
		return &v0.WireguardInstance{}, fmt.Errorf("failed to marshal response data from threeport API: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&wireguardInstances); err != nil {
		return nil, fmt.Errorf("failed to decode object in response data from threeport API: %w", err)
	}

	switch {
	case len(wireguardInstances) < 1:
		return &v0.WireguardInstance{}, errors.New(fmt.Sprintf("no wireguard instance with name %s", name))
	case len(wireguardInstances) > 1:
		return &v0.WireguardInstance{}, errors.New(fmt.Sprintf("more than one wireguard instance with name %s returned", name))
	}

	return &wireguardInstances[0], nil
}

// CreateWireguardInstance creates a new wireguard instance.
func CreateWireguardInstance(apiClient *http.Client, apiAddr string, wireguardInstance *v0.WireguardInstance) (*v0.WireguardInstance, error) {
	tpclient_lib.ReplaceAssociatedObjectsWithNil(wireguardInstance)
	jsonWireguardInstance, err := tputil.MarshalObject(wireguardInstance)
	if err != nil {
		return wireguardInstance, fmt.Errorf("failed to marshal provided object to JSON: %w", err)
	}

	response, err := tpclient_lib.GetResponse(
		apiClient,
		fmt.Sprintf("%s%s", apiAddr, v0.PathWireguardInstances),
		http.MethodPost,
		bytes.NewBuffer(jsonWireguardInstance),
		map[string]string{},
		http.StatusCreated,
	)
	if err != nil {
		return wireguardInstance, fmt.Errorf("call to threeport API returned unexpected response: %w", err)
	}

	jsonData, err := json.Marshal(response.Data[0])
	if err != nil {
		return wireguardInstance, fmt.Errorf("failed to marshal response data from threeport API: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&wireguardInstance); err != nil {
		return nil, fmt.Errorf("failed to decode object in response data from threeport API: %w", err)
	}

	return wireguardInstance, nil
}

// UpdateWireguardInstance updates a wireguard instance.
func UpdateWireguardInstance(apiClient *http.Client, apiAddr string, wireguardInstance *v0.WireguardInstance) (*v0.WireguardInstance, error) {
	tpclient_lib.ReplaceAssociatedObjectsWithNil(wireguardInstance)
	// capture the object ID, make a copy of the object, then remove fields that
	// cannot be updated in the API
	wireguardInstanceID := *wireguardInstance.ID
	payloadWireguardInstance := *wireguardInstance
	payloadWireguardInstance.ID = nil
	payloadWireguardInstance.CreatedAt = nil
	payloadWireguardInstance.UpdatedAt = nil

	jsonWireguardInstance, err := tputil.MarshalObject(payloadWireguardInstance)
	if err != nil {
		return wireguardInstance, fmt.Errorf("failed to marshal provided object to JSON: %w", err)
	}

	response, err := tpclient_lib.GetResponse(
		apiClient,
		fmt.Sprintf("%s%s/%d", apiAddr, v0.PathWireguardInstances, wireguardInstanceID),
		http.MethodPatch,
		bytes.NewBuffer(jsonWireguardInstance),
		map[string]string{},
		http.StatusOK,
	)
	if err != nil {
		return wireguardInstance, fmt.Errorf("call to threeport API returned unexpected response: %w", err)
	}

	jsonData, err := json.Marshal(response.Data[0])
	if err != nil {
		return wireguardInstance, fmt.Errorf("failed to marshal response data from threeport API: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&payloadWireguardInstance); err != nil {
		return nil, fmt.Errorf("failed to decode object in response data from threeport API: %w", err)
	}

	payloadWireguardInstance.ID = &wireguardInstanceID
	return &payloadWireguardInstance, nil
}

// DeleteWireguardInstance deletes a wireguard instance by ID.
func DeleteWireguardInstance(apiClient *http.Client, apiAddr string, id uint) (*v0.WireguardInstance, error) {
	var wireguardInstance v0.WireguardInstance

	response, err := tpclient_lib.GetResponse(
		apiClient,
		fmt.Sprintf("%s%s/%d", apiAddr, v0.PathWireguardInstances, id),
		http.MethodDelete,
		new(bytes.Buffer),
		map[string]string{},
		http.StatusOK,
	)
	if err != nil {
		return &wireguardInstance, fmt.Errorf("call to threeport API returned unexpected response: %w", err)
	}

	jsonData, err := json.Marshal(response.Data[0])
	if err != nil {
		return &wireguardInstance, fmt.Errorf("failed to marshal response data from threeport API: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&wireguardInstance); err != nil {
		return nil, fmt.Errorf("failed to decode object in response data from threeport API: %w", err)
	}

	return &wireguardInstance, nil
}
