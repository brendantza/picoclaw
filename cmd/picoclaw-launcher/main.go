// PicoClaw Launcher - Standalone HTTP service
//
// Provides a web-based JSON editor for picoclaw config files,
// with OAuth provider authentication support and team management.
//
// Usage:
//
//	go build -o picoclaw-launcher ./cmd/picoclaw-launcher/
//	./picoclaw-launcher [config.json]
//	./picoclaw-launcher -public config.json

package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/sipeed/picoclaw/cmd/picoclaw-launcher/internal/server"
	"github.com/sipeed/picoclaw/pkg/discovery"
	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/teams"
)

//go:embed internal/ui/index.html
var staticFiles embed.FS

func main() {
	public := flag.Bool("public", false, "Listen on all interfaces (0.0.0.0) instead of localhost only")
	noMDNS := flag.Bool("no-mdns", false, "Disable mDNS advertisement for teams")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "PicoClaw Launcher - A web-based configuration editor\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [config.json]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  config.json    Path to the configuration file (default: ~/.picoclaw/config.json)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s                          Use default config path\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s ./config.json             Specify a config file\n", os.Args[0])
		fmt.Fprintf(
			os.Stderr,
			"  %s -public ./config.json     Allow access from other devices on the network\n",
			os.Args[0],
		)
	}
	flag.Parse()

	configPath := server.DefaultConfigPath()
	if flag.NArg() > 0 {
		configPath = flag.Arg(0)
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		log.Fatalf("Failed to resolve config path: %v", err)
	}

	// Initialize team service
	teamsStoragePath := server.DefaultTeamsStoragePath()
	teamService, err := teams.NewService(teamsStoragePath)
	if err != nil {
		log.Fatalf("Failed to initialize team service: %v", err)
	}

	var addr string
	if *public {
		addr = "0.0.0.0:" + server.DefaultPort
	} else {
		addr = "127.0.0.1:" + server.DefaultPort
	}

	mux := http.NewServeMux()
	server.RegisterConfigAPI(mux, absPath)
	server.RegisterAuthAPI(mux, absPath)
	server.RegisterProcessAPI(mux, absPath)
	server.RegisterTeamsAPI(mux, teamService)

	staticFS, err := fs.Sub(staticFiles, "internal/ui")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// Start mDNS discovery for teams if enabled
	var mdnsServer *discovery.Server
	if !*noMDNS {
		mdnsServer = startMDNSForTeams(teamService)
	}

	// Print startup banner
	fmt.Println("=============================================")
	fmt.Println("  PicoClaw Launcher")
	fmt.Println("=============================================")
	fmt.Printf("  Config file : %s\n", absPath)
	fmt.Printf("  Teams store : %s\n", teamsStoragePath)
	fmt.Printf("  Listen addr : %s\n", addr)
	if mdnsServer != nil {
		fmt.Println("  mDNS        : Enabled (_picoclaw._tcp)")
	}
	fmt.Println()
	fmt.Println("  Open the following URL in your browser")
	fmt.Println("  to view and edit the configuration:")
	fmt.Println()
	fmt.Printf("    >> http://localhost:%s <<\n", server.DefaultPort)
	if *public {
		if ip := server.GetLocalIP(); ip != "" {
			fmt.Printf("    >> http://%s:%s <<\n", ip, server.DefaultPort)
		}
	}
	fmt.Println()
	// fmt.Println("=============================================")

	go func() {
		// Wait briefly to ensure the server is ready before opening the browser
		time.Sleep(500 * time.Millisecond)
		url := "http://localhost:" + server.DefaultPort
		if err := openBrowser(url); err != nil {
			log.Printf("Warning: Failed to auto-open browser: %v\n", err)
		}
	}()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\nShutting down...")
		if mdnsServer != nil {
			mdnsServer.Stop()
		}
		os.Exit(0)
	}()

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// openBrowser automatically opens the given URL in the default browser.
func openBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

// startMDNSForTeams starts mDNS advertisement for the first active team
func startMDNSForTeams(teamService *teams.Service) *discovery.Server {
	// List teams and advertise the first active one
	teamList, err := teamService.ListTeams()
	if err != nil {
		logger.WarnCF("launcher", "Failed to list teams for mDNS", map[string]any{"error": err})
		return nil
	}

	// Find first active team
	var activeTeam *teams.Team
	for _, t := range teamList {
		if t.Status == teams.TeamStatusActive {
			team, err := teamService.GetTeam(t.ID)
			if err == nil && team.Status == teams.TeamStatusActive {
				activeTeam = team
				break
			}
		}
	}

	if activeTeam == nil {
		// No active teams, nothing to advertise
		return nil
	}

	// Get the raw team key from metadata (needed for hash computation)
	rawKey := activeTeam.Metadata["raw_key"]
	if rawKey == "" {
		// Try to use the hashed key as fallback
		rawKey = activeTeam.TeamKey
	}

	port := 18800 // Default launcher port
	mdnsServer, err := discovery.NewServer(rawKey, port)
	if err != nil {
		logger.WarnCF("launcher", "Failed to create mDNS server", map[string]any{"error": err})
		return nil
	}

	if err := mdnsServer.Start(); err != nil {
		logger.WarnCF("launcher", "Failed to start mDNS server", map[string]any{"error": err})
		return nil
	}

	logger.InfoCF("launcher", "mDNS advertisement started for team",
		map[string]any{
			"team_id":   activeTeam.ID,
			"team_hash": mdnsServer.GetTeamHash()[:16] + "...",
		})

	return mdnsServer
}
