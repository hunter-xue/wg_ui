package wg

import (
	"context"
	"fmt"
	"os"

	"wg_ui/internal/shell"
)

const configPath = "/etc/wireguard/wg0.conf"

func WriteServerConfig(content string) error {
	return os.WriteFile(configPath, []byte(content), 0600)
}

func SyncConfig(ctx context.Context) error {
	_, stderr, err := shell.Run(ctx, "bash", "-c", "wg syncconf wg0 <(wg-quick strip wg0)")
	if err != nil {
		return fmt.Errorf("wg syncconf failed: %w: %s", err, stderr)
	}
	return nil
}

func RestartService(ctx context.Context) error {
	_, stderr, err := shell.Run(ctx, "systemctl", "restart", "wg-quick@wg0")
	if err != nil {
		return fmt.Errorf("systemctl restart failed: %w: %s", err, stderr)
	}
	return nil
}

func ServiceStatus(ctx context.Context) (string, error) {
	stdout, _, _ := shell.Run(ctx, "systemctl", "status", "wg-quick@wg0")
	return stdout, nil
}
