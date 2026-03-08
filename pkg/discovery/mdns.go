// PicoClaw - mDNS discovery for teams
// License: MIT

package discovery

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
	"github.com/sipeed/picoclaw/pkg/logger"
)

const (
	// ServiceName is the mDNS service name for PicoClaw team discovery
	ServiceName = "_picoclaw._tcp"
	
	// TeamHashPrefix is the TXT record prefix for team hash
	TeamHashPrefix = "teamhash="
)

// Server handles mDNS advertisement for team discovery
type Server struct {
	teamHash string
	port     int
	service  *mdns.Server
	mu       sync.RWMutex
}

// NewServer creates a new mDNS discovery server
func NewServer(teamKey string, port int) (*Server, error) {
	teamHash := computeTeamHash(teamKey)
	
	return &Server{
		teamHash: teamHash,
		port:     port,
	}, nil
}

// Start begins advertising the team via mDNS
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.service != nil {
		return nil // Already running
	}

	// Create service info
	info := []string{
		"PicoClaw Team Gateway",
		TeamHashPrefix + s.teamHash,
		"version=1.0",
	}

	service, err := mdns.NewMDNSService(
		"picoclaw-gateway", // instance name
		ServiceName,        // service name
		"",                 // domain
		"",                 // host (auto)
		s.port,             // port
		nil,                // IPs (auto)
		info,               // TXT records
	)
	if err != nil {
		return fmt.Errorf("failed to create mDNS service: %w", err)
	}

	// Create mDNS server
	server, err := mdns.NewServer(&mdns.Config{
		Zone: service,
	})
	if err != nil {
		return fmt.Errorf("failed to start mDNS server: %w", err)
	}

	s.service = server

	logger.InfoCF("discovery", "mDNS advertisement started",
		map[string]any{
			"service":   ServiceName,
			"port":      s.port,
			"team_hash": s.teamHash[:16] + "...",
		})

	return nil
}

// Stop stops the mDNS advertisement
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.service == nil {
		return nil
	}

	if err := s.service.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown mDNS server: %w", err)
	}

	s.service = nil

	logger.InfoCF("discovery", "mDNS advertisement stopped", nil)

	return nil
}

// UpdateTeamKey updates the advertised team hash (e.g., after key rotation)
func (s *Server) UpdateTeamKey(teamKey string) error {
	s.mu.Lock()
	oldHash := s.teamHash
	s.teamHash = computeTeamHash(teamKey)
	wasRunning := s.service != nil
	s.mu.Unlock()

	if oldHash == s.teamHash {
		return nil // No change
	}

	// Restart if running
	if wasRunning {
		if err := s.Stop(); err != nil {
			return err
		}
		return s.Start()
	}

	return nil
}

// GetTeamHash returns the current team hash being advertised
func (s *Server) GetTeamHash() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.teamHash
}

// computeTeamHash computes the first 8 bytes of SHA256(teamKey)
func computeTeamHash(teamKey string) string {
	hash := sha256.Sum256([]byte(teamKey))
	return hex.EncodeToString(hash[:8])
}

// getLocalIP returns the first non-loopback IPv4 address
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "127.0.0.1", nil // Fallback to localhost
}

// Client handles mDNS discovery of teams
type Client struct{}

// NewClient creates a new mDNS discovery client
func NewClient() *Client {
	return &Client{}
}

// Discover searches for PicoClaw team gateways on the local network
func (c *Client) Discover(timeoutSec int) ([]DiscoveredTeam, error) {
	entriesCh := make(chan *mdns.ServiceEntry, 10)
	entries := make([]*mdns.ServiceEntry, 0)
	done := make(chan struct{})

	go func() {
		for entry := range entriesCh {
			entries = append(entries, entry)
		}
		close(done)
	}()

	// Start lookup
	params := &mdns.QueryParam{
		Service:     ServiceName,
		Domain:      "local",
		Timeout:     time.Duration(timeoutSec) * time.Second,
		Entries:     entriesCh,
		DisableIPv6: true,
	}

	if err := mdns.Query(params); err != nil {
		return nil, fmt.Errorf("mDNS query failed: %w", err)
	}

	close(entriesCh)
	<-done

	// Convert to DiscoveredTeam
	teams := make([]DiscoveredTeam, 0, len(entries))
	for _, entry := range entries {
		team := DiscoveredTeam{
			Host: entry.Host,
			Port: entry.Port,
			IPs:  []net.IP{entry.AddrV4},
		}

		// Parse TXT records
		for _, txt := range entry.InfoFields {
			if strings.HasPrefix(txt, TeamHashPrefix) {
				team.TeamHash = strings.TrimPrefix(txt, TeamHashPrefix)
			}
		}

		if team.TeamHash != "" {
			teams = append(teams, team)
		}
	}

	return teams, nil
}

// DiscoveredTeam represents a discovered team gateway
type DiscoveredTeam struct {
	Host     string
	Port     int
	IPs      []net.IP
	TeamHash string
}

// MatchesTeam checks if the discovered team matches a given team key
func (d *DiscoveredTeam) MatchesTeam(teamKey string) bool {
	expectedHash := computeTeamHash(teamKey)
	return d.TeamHash == expectedHash
}

// GetBestAddress returns the best address to connect to
func (d *DiscoveredTeam) GetBestAddress() string {
	if len(d.IPs) > 0 {
		return fmt.Sprintf("http://%s:%d", d.IPs[0].String(), d.Port)
	}
	return fmt.Sprintf("http://%s:%d", d.Host, d.Port)
}
