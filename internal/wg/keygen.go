package wg

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"wg_ui/internal/shell"
)

func GenerateKeypair(ctx context.Context) (privateKey, publicKey string, err error) {
	privOut, privErr, err := shell.Run(ctx, "wg", "genkey")
	if err != nil {
		return "", "", fmt.Errorf("wg genkey failed: %w: %s", err, privErr)
	}
	privateKey = strings.TrimSpace(privOut)

	cmd := exec.CommandContext(ctx, "wg", "pubkey")
	cmd.Stdin = strings.NewReader(privateKey)
	pubBytes, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("wg pubkey failed: %w", err)
	}
	publicKey = strings.TrimSpace(string(pubBytes))

	return privateKey, publicKey, nil
}
