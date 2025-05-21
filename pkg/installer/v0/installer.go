// package v0

// import (
// 	"fmt"

// 	installer "github.com/threeport/threeport/pkg/installer/v0"
// )

// // cpi.getDelveArgs returns the args that are passed to delve.
// // func (cpi *ControlPlaneInstaller) getDelveArgs(name string) []string {
// func (i *installer.Installer) getDelveArgs(name string) []string {
// 	args := []string{
// 		"--continue",
// 		"--accept-multiclient",
// 		"--listen=:40000",
// 		"--headless=true",
// 		"--api-version=2",
// 	}

// 	if i.Debug {
// 		args = append(args, "--log")
// 	}

// 	args = append(args, "exec")
// 	args = append(args, fmt.Sprintf("/%s", name))
// 	return args
// }

// // func (cpi *ControlPlaneInstaller) getReadinessProbe() map[string]interface{} {
// func (i *installer.Installer) getReadinessProbe() map[string]interface{} {
// 	var readinessProbe map[string]interface{}
// 	if !i.Debug {
// 		readinessProbe = map[string]interface{}{
// 			"failureThreshold": 1,
// 			"httpGet": map[string]interface{}{
// 				"path":   "/readyz",
// 				"port":   8081,
// 				"scheme": "HTTP",
// 			},
// 			"initialDelaySeconds": 1,
// 			"periodSeconds":       2,
// 			"successThreshold":    1,
// 			"timeoutSeconds":      1,
// 		}
// 	}
// 	return readinessProbe
// }

// // getCommand returns the args that are passed to the container.
// // func (cpi *ControlPlaneInstaller) getCommand(name string) []interface{} {
// func (i *installer.Installer) getCommand(name string) []interface{} {

// 	switch {
// 	case i.Debug:
// 		return []interface{}{
// 			"/usr/local/bin/dlv",
// 		}
// 	default:
// 		return []interface{}{
// 			fmt.Sprintf("/%s", name),
// 		}
// 	}
// }
