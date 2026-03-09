// PicoClaw - Team API proxy to gateway
// License: MIT

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// GatewayClient proxies team API requests to the gateway
type GatewayClient struct {
	gatewayURL string
	client     *http.Client
}

// NewGatewayClient creates a new gateway client
func NewGatewayClient() *GatewayClient {
	gatewayURL := os.Getenv("PICOCLAW_GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost:18790"
	}

	return &GatewayClient{
		gatewayURL: gatewayURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsAvailable checks if the gateway is accessible
func (c *GatewayClient) IsAvailable() bool {
	resp, err := c.client.Get(c.gatewayURL + "/api/teams")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ProxyTeamsList proxies the team list request to gateway
func (c *GatewayClient) ProxyTeamsList(w http.ResponseWriter, r *http.Request) {
	c.proxyRequest(w, r, "/api/teams")
}

// ProxyTeamGet proxies get team request to gateway
func (c *GatewayClient) ProxyTeamGet(w http.ResponseWriter, r *http.Request, teamID string) {
	c.proxyRequest(w, r, fmt.Sprintf("/api/teams/%s", teamID))
}

// ProxyTeamAgents proxies team agents request to gateway
func (c *GatewayClient) ProxyTeamAgents(w http.ResponseWriter, r *http.Request, teamID string) {
	c.proxyRequest(w, r, fmt.Sprintf("/api/teams/%s/agents", teamID))
}

// ProxyTeamCreate proxies create team request to gateway
func (c *GatewayClient) ProxyTeamCreate(w http.ResponseWriter, r *http.Request) {
	c.proxyRequestWithBody(w, r, "/api/teams")
}

// ProxyTeamUpdate proxies update team request to gateway
func (c *GatewayClient) ProxyTeamUpdate(w http.ResponseWriter, r *http.Request, teamID string) {
	c.proxyRequestWithBody(w, r, fmt.Sprintf("/api/teams/%s", teamID))
}

// ProxyTeamDelete proxies delete team request to gateway
func (c *GatewayClient) ProxyTeamDelete(w http.ResponseWriter, r *http.Request, teamID string) {
	c.proxyRequest(w, r, fmt.Sprintf("/api/teams/%s", teamID))
}

// ProxyTeamRotateKey proxies rotate key request to gateway
func (c *GatewayClient) ProxyTeamRotateKey(w http.ResponseWriter, r *http.Request, teamID string) {
	c.proxyRequestWithBody(w, r, fmt.Sprintf("/api/teams/%s/rotate-key", teamID))
}

// ProxyEvictAgent proxies evict agent request to gateway
func (c *GatewayClient) ProxyEvictAgent(w http.ResponseWriter, r *http.Request, teamID, agentID string) {
	c.proxyRequestWithBody(w, r, fmt.Sprintf("/api/teams/%s/agents/%s/evict", teamID, agentID))
}

// ProxyWipeAgent proxies wipe agent request to gateway
func (c *GatewayClient) ProxyWipeAgent(w http.ResponseWriter, r *http.Request, teamID, agentID string) {
	c.proxyRequestWithBody(w, r, fmt.Sprintf("/api/teams/%s/agents/%s/wipe", teamID, agentID))
}

func (c *GatewayClient) proxyRequest(w http.ResponseWriter, r *http.Request, path string) {
	url := c.gatewayURL + path

	req, err := http.NewRequest(r.Method, url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	json.NewDecoder(resp.Body).Decode(struct{}{})
}

func (c *GatewayClient) proxyRequestWithBody(w http.ResponseWriter, r *http.Request, path string) {
	url := c.gatewayURL + path

	req, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	json.NewDecoder(resp.Body).Decode(struct{}{})
}
