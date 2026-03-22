package wg

import (
	"fmt"
	"strings"

	"wg_ui/internal/db"
)

func GenerateServerConfig(server *db.Server, clients []db.Client) string {
	var b strings.Builder

	b.WriteString("[Interface]\n")
	b.WriteString(fmt.Sprintf("Address = %s\n", server.Address))
	b.WriteString(fmt.Sprintf("ListenPort = %d\n", server.ListenPort))
	b.WriteString(fmt.Sprintf("PrivateKey = %s\n", server.PrivateKey))
	if server.PostUp != "" {
		b.WriteString(fmt.Sprintf("PostUp = %s\n", server.PostUp))
	}
	if server.PostDown != "" {
		b.WriteString(fmt.Sprintf("PostDown = %s\n", server.PostDown))
	}
	if server.DNS != "" {
		b.WriteString(fmt.Sprintf("DNS = %s\n", server.DNS))
	}
	b.WriteString(fmt.Sprintf("MTU = %d\n", server.MTU))

	for _, c := range clients {
		if c.Disabled != 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("\n[Peer]\n"))
		b.WriteString(fmt.Sprintf("PublicKey = %s\n", c.PublicKey))
		b.WriteString(fmt.Sprintf("AllowedIPs = %s\n", c.Address))
	}

	return b.String()
}

func GenerateClientConfig(server *db.Server, client *db.Client, endpoint string) string {
	var b strings.Builder

	b.WriteString("[Interface]\n")
	b.WriteString(fmt.Sprintf("PrivateKey = %s\n", client.PrivateKey))
	b.WriteString(fmt.Sprintf("Address = %s\n", client.Address))
	if client.DNS != "" {
		b.WriteString(fmt.Sprintf("DNS = %s\n", client.DNS))
	}
	b.WriteString(fmt.Sprintf("MTU = %d\n", client.MTU))

	b.WriteString(fmt.Sprintf("\n[Peer]\n"))
	b.WriteString(fmt.Sprintf("PublicKey = %s\n", server.PublicKey))
	b.WriteString(fmt.Sprintf("AllowedIPs = %s\n", client.AllowIPs))
	b.WriteString(fmt.Sprintf("Endpoint = %s:%d\n", endpoint, server.ListenPort))
	if client.PersistentKeepalive > 0 {
		b.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", client.PersistentKeepalive))
	}

	return b.String()
}
