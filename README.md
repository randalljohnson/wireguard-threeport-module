# Wireguard Threeport Module

A Threeport module for managing Wireguard VPN configurations in Kubernetes clusters.

## Prerequisites

- A running Threeport control plane
- Kubernetes cluster with load balancer support

## Installation

1. Build the module:
```bash
mage build:allImagesRelease
mage build:plugin
```

2. Install the module to your Threeport control plane:
```bash
./bin/wireguard install
```

## Usage

### Creating a Wireguard Configuration

1. Create a configuration file (e.g., `wireguard-config.yaml`):
```yaml
Wireguard:
  Name: wg-test
```

2. Create the Wireguard instance:
```bash
./bin/wireguard create wireguard --config=wireguard-config.yaml
```

### Getting Wireguard Configuration

To retrieve the Wireguard client configuration:
```bash
./bin/wireguard get wireguard-conf --name=wg-test
```

This will output a Wireguard client configuration that can be used to connect to the VPN.

IMPORTANT: This command will print your client configuration's private key to stdout.
Consider redirecting output to a file if you wish to avoid this.

Download your relevant Wireguard client [here](https://www.wireguard.com/install/) and use the above file to configure it.

### Deleting a Wireguard Configuration

To remove a Wireguard configuration:
```bash
./bin/wireguard delete wireguard --config=wireguard-config.yaml
```

## Development

For development purposes, you can enable debug mode during installation:
```bash
./bin/wireguard install --debug
```