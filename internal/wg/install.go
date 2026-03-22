package wg

import (
	"context"
	"fmt"

	"wg_ui/internal/shell"
)

// IsInstalled checks whether the wg binary is available on the system.
func IsInstalled(ctx context.Context) bool {
	_, _, err := shell.Run(ctx, "which", "wg")
	return err == nil
}

type InstallStep struct {
	Name string
	Run  func(ctx context.Context) (string, error)
}

// InstallSteps is the ordered list of installation steps.
var InstallSteps = []InstallStep{
	{
		Name: "Updating package list...",
		Run: func(ctx context.Context) (string, error) {
			stdout, stderr, err := shell.Run(ctx, "apt-get", "update", "-y")
			if err != nil {
				return stderr, fmt.Errorf("apt-get update failed: %w", err)
			}
			return stdout, nil
		},
	},
	{
		Name: "Installing WireGuard...",
		Run: func(ctx context.Context) (string, error) {
			stdout, stderr, err := shell.Run(ctx, "apt-get", "install", "-y", "wireguard")
			if err != nil {
				return stderr, fmt.Errorf("apt-get install wireguard failed: %w", err)
			}
			return stdout, nil
		},
	},
	{
		Name: "Checking iptables...",
		Run: func(ctx context.Context) (string, error) {
			// iptables is not included by default on Debian 11+
			_, _, err := shell.Run(ctx, "which", "iptables")
			if err == nil {
				return "iptables already installed", nil
			}
			stdout, stderr, err := shell.Run(ctx, "apt-get", "install", "-y", "iptables")
			if err != nil {
				return stderr, fmt.Errorf("apt-get install iptables failed: %w", err)
			}
			return stdout, nil
		},
	},
	{
		Name: "Enabling IP forwarding...",
		Run: func(ctx context.Context) (string, error) {
			stdout, stderr, err := shell.Run(ctx, "sysctl", "-w", "net.ipv4.ip_forward=1")
			if err != nil {
				return stderr, fmt.Errorf("sysctl failed: %w", err)
			}
			return stdout, nil
		},
	},
	{
		Name: "Configuring systemd service...",
		Run: func(ctx context.Context) (string, error) {
			stdout, stderr, err := shell.Run(ctx, "systemctl", "enable", "wg-quick@wg0")
			if err != nil {
				return stderr, fmt.Errorf("systemctl enable failed: %w", err)
			}
			return stdout, nil
		},
	},
}

func Install(ctx context.Context) error {
	for _, step := range InstallSteps {
		if _, err := step.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
