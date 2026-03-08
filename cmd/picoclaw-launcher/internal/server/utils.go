package server

import (
	"net"
	"os"
	"path/filepath"
)

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(home, ".picoclaw", "config.json")
}

func DefaultTeamsStoragePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "teams"
	}
	return filepath.Join(home, ".picoclaw", "teams")
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}
