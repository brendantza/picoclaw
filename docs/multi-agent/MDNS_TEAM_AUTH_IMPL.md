# mDNS Discovery with Team Authentication - Implementation Plan

## Overview

This document provides a detailed implementation plan for adding mDNS-based discovery with team key authentication to PicoClaw's multi-agent system.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    AGENT DISCOVERY LAYER                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │  mDNS Client │  │  mDNS Server │  │  Team Auth Manager   │  │
│  │  (browse)    │  │  (advertise) │  │  (challenge-response)│  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                    SECURE COMMUNICATION                          │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  TCP Transport  →  Team Auth  →  Encrypted Channel (AES)  │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Implementation Files

### 1. Core Discovery Package

**File:** `pkg/agent/discovery/mdns.go`
```go
package discovery

import (
    "context"
    "fmt"
    "strings"
    "time"
    
    "github.com/grandcat/zeroconf"
    "github.com/sipeed/picoclaw/pkg/logger"
)

const (
    ServiceType = "_picoclaw-agent._tcp"
    Domain      = "local."
)

// MDNSDiscovery handles mDNS-based agent discovery
type MDNSDiscovery struct {
    agentID      string
    role         string
    teamHash     string
    port         int
    capabilities []string
    
    server   *zeroconf.Server
    resolver *zeroconf.Resolver
    
    // Callbacks
    onPeerDiscovered func(PeerInfo)
    onPeerLost       func(string)
}

// PeerInfo represents a discovered agent
type PeerInfo struct {
    AgentID      string
    Role         string
    TeamHash     string
    Hostname     string
    Port         int
    Addresses    []string
    Capabilities []string
    Version      string
    LastSeen     time.Time
}

// NewMDNSDiscovery creates a new mDNS discovery instance
func NewMDNSDiscovery(config MDNSConfig) (*MDNSDiscovery, error) {
    resolver, err := zeroconf.NewResolver(nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create mDNS resolver: %w", err)
    }
    
    return &MDNSDiscovery{
        agentID:      config.AgentID,
        role:         config.Role,
        teamHash:     config.TeamHash,
        port:         config.Port,
        capabilities: config.Capabilities,
        resolver:     resolver,
    }, nil
}

// StartAdvertisement begins advertising this agent on the network
func (m *MDNSDiscovery) StartAdvertisement() error {
    if m.server != nil {
        return fmt.Errorf("already advertising")
    }
    
    txtRecords := []string{
        fmt.Sprintf("team_hash=%s", m.teamHash),
        fmt.Sprintf("role=%s", m.role),
        fmt.Sprintf("agent_id=%s", m.agentID),
        fmt.Sprintf("version=%s", getVersion()),
        fmt.Sprintf("capabilities=%s", strings.Join(m.capabilities, ",")),
    }
    
    server, err := zeroconf.Register(
        m.agentID,
        ServiceType,
        Domain,
        m.port,
        txtRecords,
        nil,
    )
    if err != nil {
        return fmt.Errorf("failed to register mDNS service: %w", err)
    }
    
    m.server = server
    logger.InfoCF("discovery", "Started mDNS advertisement",
        map[string]any{
            "agent_id": m.agentID,
            "role":     m.role,
            "port":     m.port,
        })
    
    return nil
}

// StopAdvertisement stops advertising this agent
func (m *MDNSDiscovery) StopAdvertisement() {
    if m.server != nil {
        m.server.Shutdown()
        m.server = nil
        logger.InfoC("discovery", "Stopped mDNS advertisement")
    }
}

// DiscoverPeers searches for other agents on the network
func (m *MDNSDiscovery) DiscoverPeers(ctx context.Context, targetTeamHash string) ([]PeerInfo, error) {
    entries := make(chan *zeroconf.ServiceEntry)
    var peers []PeerInfo
    
    go func() {
        for entry := range entries {
            peer := parseServiceEntry(entry)
            
            // Only include peers from the same team
            if peer.TeamHash == targetTeamHash {
                peers = append(peers, peer)
                
                if m.onPeerDiscovered != nil {
                    m.onPeerDiscovered(peer)
                }
            }
        }
    }()
    
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    err := m.resolver.Browse(ctx, ServiceType, Domain, entries)
    if err != nil {
        return nil, fmt.Errorf("failed to browse mDNS: %w", err)
    }
    
    <-ctx.Done()
    return peers, nil
}

// SetPeerDiscoveredCallback sets the callback for peer discovery
func (m *MDNSDiscovery) SetPeerDiscoveredCallback(cb func(PeerInfo)) {
    m.onPeerDiscovered = cb
}

func parseServiceEntry(entry *zeroconf.ServiceEntry) PeerInfo {
    peer := PeerInfo{
        Hostname:  entry.HostName,
        Port:      entry.Port,
        Addresses: entry.AddrIPv4,
        LastSeen:  time.Now(),
    }
    
    // Parse TXT records
    for _, txt := range entry.Text {
        parts := strings.SplitN(txt, "=", 2)
        if len(parts) != 2 {
            continue
        }
        
        key, value := parts[0], parts[1]
        switch key {
        case "team_hash":
            peer.TeamHash = value
        case "role":
            peer.Role = value
        case "agent_id":
            peer.AgentID = value
        case "version":
            peer.Version = value
        case "capabilities":
            peer.Capabilities = strings.Split(value, ",")
        }
    }
    
    // Also include IPv6 addresses
    peer.Addresses = append(peer.Addresses, entry.AddrIPv6...)
    
    return peer
}
```

### 2. Team Authentication Package

**File:** `pkg/agent/auth/team.go`
```go
package auth

import (
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "sync"
    "time"
    
    "golang.org/x/crypto/hkdf"
)

const (
    ChallengeTTL = 30 * time.Second
    NonceSize    = 32
    KeySize      = 32
)

// TeamAuthManager manages team-based authentication
type TeamAuthManager struct {
    teamKey   string
    teamHash  string
    challenges map[string]*Challenge // nonce -> challenge
    mu         sync.RWMutex
}

// Challenge represents an authentication challenge
type Challenge struct {
    Nonce     []byte
    Timestamp time.Time
    PeerAddr  string
}

// AuthMessage types
const (
    MsgTypeChallenge      = "challenge"
    MsgTypeResponse       = "response"
    MsgTypeSuccess        = "auth_success"
    MsgTypeFailure        = "auth_failure"
)

// AuthMessage is the base authentication message
type AuthMessage struct {
    Type      string          `json:"type"`
    Payload   json.RawMessage `json:"payload"`
    Timestamp int64           `json:"timestamp"`
}

// ChallengePayload sent by controller
type ChallengePayload struct {
    Nonce     string `json:"nonce"`     // base64 encoded
    Timestamp int64  `json:"timestamp"`
}

// ResponsePayload sent by worker
type ResponsePayload struct {
    AgentID      string   `json:"agent_id"`
    Signature    string   `json:"signature"`    // base64 HMAC
    Capabilities []string `json:"capabilities"`
    Role         string   `json:"role"`
}

// SuccessPayload sent after successful auth
type SuccessPayload struct {
    SessionID  string `json:"session_id"`
    ExpiresAt  int64  `json:"expires_at"`
    ServerTime int64  `json:"server_time"`
}

// NewTeamAuthManager creates a new team auth manager
func NewTeamAuthManager(teamKey string) (*TeamAuthManager, error) {
    if err := ValidateTeamKey(teamKey); err != nil {
        return nil, err
    }
    
    return &TeamAuthManager{
        teamKey:    teamKey,
        teamHash:   ComputeTeamHash(teamKey),
        challenges: make(map[string]*Challenge),
    }, nil
}

// GenerateChallenge creates a new authentication challenge (called by controller)
func (m *TeamAuthManager) GenerateChallenge(peerAddr string) (*AuthMessage, error) {
    nonce := make([]byte, NonceSize)
    if _, err := rand.Read(nonce); err != nil {
        return nil, err
    }
    
    nonceStr := base64.StdEncoding.EncodeToString(nonce)
    
    m.mu.Lock()
    m.challenges[nonceStr] = &Challenge{
        Nonce:     nonce,
        Timestamp: time.Now(),
        PeerAddr:  peerAddr,
    }
    m.mu.Unlock()
    
    // Cleanup old challenges
    go m.cleanupOldChallenges()
    
    payload := ChallengePayload{
        Nonce:     nonceStr,
        Timestamp: time.Now().Unix(),
    }
    
    payloadBytes, _ := json.Marshal(payload)
    
    return &AuthMessage{
        Type:      MsgTypeChallenge,
        Payload:   payloadBytes,
        Timestamp: time.Now().Unix(),
    }, nil
}

// CreateResponse creates an authentication response (called by worker)
func (m *TeamAuthManager) CreateResponse(
    challenge *ChallengePayload,
    agentID string,
    role string,
    capabilities []string,
) (*AuthMessage, error) {
    signature, err := m.signChallenge(challenge.Nonce, challenge.Timestamp)
    if err != nil {
        return nil, err
    }
    
    payload := ResponsePayload{
        AgentID:      agentID,
        Signature:    base64.StdEncoding.EncodeToString(signature),
        Role:         role,
        Capabilities: capabilities,
    }
    
    payloadBytes, _ := json.Marshal(payload)
    
    return &AuthMessage{
        Type:      MsgTypeResponse,
        Payload:   payloadBytes,
        Timestamp: time.Now().Unix(),
    }, nil
}

// VerifyResponse verifies a worker's response (called by controller)
func (m *TeamAuthManager) VerifyResponse(
    nonceStr string,
    response *ResponsePayload,
    peerAddr string,
) (*SuccessPayload, error) {
    m.mu.RLock()
    challenge, exists := m.challenges[nonceStr]
    m.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("challenge not found or expired")
    }
    
    // Verify challenge hasn't expired
    if time.Since(challenge.Timestamp) > ChallengeTTL {
        m.removeChallenge(nonceStr)
        return nil, fmt.Errorf("challenge expired")
    }
    
    // Verify peer address matches (optional, prevents replay from different IP)
    if challenge.PeerAddr != "" && challenge.PeerAddr != peerAddr {
        // Log but don't fail - NAT might cause address changes
    }
    
    // Verify signature
    expectedSig, err := m.signChallenge(nonceStr, challenge.Timestamp.Unix())
    if err != nil {
        return nil, err
    }
    
    actualSig, err := base64.StdEncoding.DecodeString(response.Signature)
    if err != nil {
        return nil, fmt.Errorf("invalid signature encoding: %w", err)
    }
    
    if !hmac.Equal(expectedSig, actualSig) {
        return nil, fmt.Errorf("invalid signature")
    }
    
    // Remove used challenge
    m.removeChallenge(nonceStr)
    
    // Generate session info
    sessionID := generateSessionID()
    
    return &SuccessPayload{
        SessionID:  sessionID,
        ExpiresAt:  time.Now().Add(24 * time.Hour).Unix(),
        ServerTime: time.Now().Unix(),
    }, nil
}

// DeriveSessionKey derives a session encryption key
func (m *TeamAuthManager) DeriveSessionKey(nonceStr string, sessionID string) ([]byte, error) {
    nonce, err := base64.StdEncoding.DecodeString(nonceStr)
    if err != nil {
        return nil, err
    }
    
    hkdfReader := hkdf.New(
        sha256.New,
        []byte(m.teamKey),
        nonce,
        []byte("picoclaw-session-"+sessionID),
    )
    
    key := make([]byte, KeySize)
    if _, err := hkdfReader.Read(key); err != nil {
        return nil, err
    }
    
    return key, nil
}

// signChallenge creates HMAC signature
func (m *TeamAuthManager) signChallenge(nonceStr string, timestamp int64) ([]byte, error) {
    nonce, err := base64.StdEncoding.DecodeString(nonceStr)
    if err != nil {
        return nil, err
    }
    
    data := fmt.Sprintf("%s:%d", nonceStr, timestamp)
    
    h := hmac.New(sha256.New, []byte(m.teamKey))
    h.Write([]byte(data))
    return h.Sum(nil), nil
}

// removeChallenge removes a used/expired challenge
func (m *TeamAuthManager) removeChallenge(nonce string) {
    m.mu.Lock()
    delete(m.challenges, nonce)
    m.mu.Unlock()
}

// cleanupOldChallenges removes expired challenges
func (m *TeamAuthManager) cleanupOldChallenges() {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    for nonce, challenge := range m.challenges {
        if time.Since(challenge.Timestamp) > ChallengeTTL*2 {
            delete(m.challenges, nonce)
        }
    }
}

// GetTeamHash returns the team hash for this manager
func (m *TeamAuthManager) GetTeamHash() string {
    return m.teamHash
}

func generateSessionID() string {
    b := make([]byte, 16)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}
```

### 3. Secure Transport

**File:** `pkg/agent/transport/secure.go`
```go
package transport

import (
    "crypto/aes"
    "crypto/cipher"
    "encoding/binary"
    "encoding/json"
    "fmt"
    "io"
    "net"
    "sync"
)

const (
    MaxMessageSize = 10 * 1024 * 1024 // 10MB
    HeaderSize     = 4                 // uint32 for length
)

// SecureConn wraps a net.Conn with encryption
type SecureConn struct {
    conn       net.Conn
    sendCipher cipher.AEAD
    recvCipher cipher.AEAD
    sendNonce  uint64
    recvNonce  uint64
    mu         sync.Mutex
}

// NewSecureConn creates an encrypted connection
func NewSecureConn(conn net.Conn, sessionKey []byte) (*SecureConn, error) {
    block, err := aes.NewCipher(sessionKey)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    return &SecureConn{
        conn:       conn,
        sendCipher: gcm,
        recvCipher: gcm,
    }, nil
}

// Send encrypts and sends a message
func (s *SecureConn) Send(msg interface{}) error {
    plaintext, err := json.Marshal(msg)
    if err != nil {
        return err
    }
    
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Create nonce
    nonce := make([]byte, s.sendCipher.NonceSize())
    binary.BigEndian.PutUint64(nonce, s.sendNonce)
    s.sendNonce++
    
    // Encrypt
    ciphertext := s.sendCipher.Seal(nonce, nonce, plaintext, nil)
    
    // Send length prefix
    length := uint32(len(ciphertext))
    if err := binary.Write(s.conn, binary.BigEndian, length); err != nil {
        return err
    }
    
    // Send ciphertext
    if _, err := s.conn.Write(ciphertext); err != nil {
        return err
    }
    
    return nil
}

// Receive decrypts a message
func (s *SecureConn) Receive(msg interface{}) error {
    // Read length
    var length uint32
    if err := binary.Read(s.conn, binary.BigEndian, &length); err != nil {
        return err
    }
    
    if length > MaxMessageSize {
        return fmt.Errorf("message too large: %d", length)
    }
    
    // Read ciphertext
    ciphertext := make([]byte, length)
    if _, err := io.ReadFull(s.conn, ciphertext); err != nil {
        return err
    }
    
    // Extract nonce
    nonceSize := s.recvCipher.NonceSize()
    if len(ciphertext) < nonceSize {
        return fmt.Errorf("ciphertext too short")
    }
    nonce := ciphertext[:nonceSize]
    ciphertext = ciphertext[nonceSize:]
    
    // Decrypt
    plaintext, err := s.recvCipher.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return fmt.Errorf("decryption failed: %w", err)
    }
    
    return json.Unmarshal(plaintext, msg)
}

// Close closes the underlying connection
func (s *SecureConn) Close() error {
    return s.conn.Close()
}

// RemoteAddr returns the remote address
func (s *SecureConn) RemoteAddr() net.Addr {
    return s.conn.RemoteAddr()
}
```

### 4. Controller-Side Integration

**File:** `pkg/agent/controller_server.go`
```go
package agent

import (
    "context"
    "fmt"
    "net"
    "sync"
    
    "github.com/sipeed/picoclaw/pkg/agent/auth"
    "github.com/sipeed/picoclaw/pkg/agent/discovery"
    "github.com/sipeed/picoclaw/pkg/agent/transport"
    "github.com/sipeed/picoclaw/pkg/logger"
)

// ControllerServer accepts connections from worker agents
type ControllerServer struct {
    listener      net.Listener
    authManager   *auth.TeamAuthManager
    secureConns   map[string]*transport.SecureConn // agent_id -> conn
    mu            sync.RWMutex
    
    onAgentConnected    func(AgentInfo)
    onAgentDisconnected func(string)
    onMessage           func(string, []byte)
}

// NewControllerServer creates a new controller server
func NewControllerServer(bindAddr string, teamKey string) (*ControllerServer, error) {
    authManager, err := auth.NewTeamAuthManager(teamKey)
    if err != nil {
        return nil, err
    }
    
    listener, err := net.Listen("tcp", bindAddr)
    if err != nil {
        return nil, err
    }
    
    return &ControllerServer{
        listener:    listener,
        authManager: authManager,
        secureConns: make(map[string]*transport.SecureConn),
    }, nil
}

// Start begins accepting connections
func (s *ControllerServer) Start(ctx context.Context) error {
    logger.InfoCF("controller", "Starting controller server",
        map[string]any{"address": s.listener.Addr()})
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        conn, err := s.listener.Accept()
        if err != nil {
            if ctx.Err() != nil {
                return ctx.Err()
            }
            logger.ErrorCF("controller", "Accept failed", map[string]any{"error": err})
            continue
        }
        
        go s.handleConnection(ctx, conn)
    }
}

// handleConnection performs authentication and sets up secure channel
func (s *ControllerServer) handleConnection(ctx context.Context, conn net.Conn) {
    defer conn.Close()
    
    peerAddr := conn.RemoteAddr().String()
    logger.DebugCF("controller", "New connection", map[string]any{"from": peerAddr})
    
    // 1. Send challenge
    challengeMsg, err := s.authManager.GenerateChallenge(peerAddr)
    if err != nil {
        logger.ErrorCF("controller", "Failed to generate challenge", map[string]any{"error": err})
        return
    }
    
    if err := sendJSON(conn, challengeMsg); err != nil {
        logger.ErrorCF("controller", "Failed to send challenge", map[string]any{"error": err})
        return
    }
    
    // 2. Receive response
    var responseMsg auth.AuthMessage
    if err := recvJSON(conn, &responseMsg); err != nil {
        logger.ErrorCF("controller", "Failed to receive response", map[string]any{"error": err})
        return
    }
    
    if responseMsg.Type != auth.MsgTypeResponse {
        logger.ErrorCF("controller", "Expected response message", map[string]any{"got": responseMsg.Type})
        return
    }
    
    var response auth.ResponsePayload
    if err := json.Unmarshal(responseMsg.Payload, &response); err != nil {
        logger.ErrorCF("controller", "Failed to parse response", map[string]any{"error": err})
        return
    }
    
    // Extract challenge nonce from challenge payload
    var challengePayload auth.ChallengePayload
    json.Unmarshal(challengeMsg.Payload, &challengePayload)
    
    // 3. Verify response
    success, err := s.authManager.VerifyResponse(
        challengePayload.Nonce,
        &response,
        peerAddr,
    )
    if err != nil {
        logger.ErrorCF("controller", "Authentication failed",
            map[string]any{
                "agent_id": response.AgentID,
                "error":    err,
            })
        sendJSON(conn, auth.AuthMessage{Type: auth.MsgTypeFailure})
        return
    }
    
    // 4. Derive session key and create secure connection
    sessionKey, err := s.authManager.DeriveSessionKey(challengePayload.Nonce, success.SessionID)
    if err != nil {
        logger.ErrorCF("controller", "Failed to derive session key", map[string]any{"error": err})
        return
    }
    
    secureConn, err := transport.NewSecureConn(conn, sessionKey)
    if err != nil {
        logger.ErrorCF("controller", "Failed to create secure connection", map[string]any{"error": err})
        return
    }
    
    // 5. Send success
    successPayload, _ := json.Marshal(success)
    if err := secureConn.Send(auth.AuthMessage{
        Type:      auth.MsgTypeSuccess,
        Payload:   successPayload,
        Timestamp: time.Now().Unix(),
    }); err != nil {
        logger.ErrorCF("controller", "Failed to send success", map[string]any{"error": err})
        return
    }
    
    // 6. Register connection
    s.mu.Lock()
    s.secureConns[response.AgentID] = secureConn
    s.mu.Unlock()
    
    logger.InfoCF("controller", "Agent authenticated",
        map[string]any{
            "agent_id":   response.AgentID,
            "role":       response.Role,
            "session_id": success.SessionID,
        })
    
    if s.onAgentConnected != nil {
        s.onAgentConnected(AgentInfo{
            ID:           response.AgentID,
            Role:         response.Role,
            Capabilities: response.Capabilities,
        })
    }
    
    // 7. Handle messages
    s.handleMessages(ctx, response.AgentID, secureConn)
    
    // Cleanup
    s.mu.Lock()
    delete(s.secureConns, response.AgentID)
    s.mu.Unlock()
    
    if s.onAgentDisconnected != nil {
        s.onAgentDisconnected(response.AgentID)
    }
    
    logger.InfoCF("controller", "Agent disconnected", map[string]any{"agent_id": response.AgentID})
}

// handleMessages processes messages from an agent
func (s *ControllerServer) handleMessages(ctx context.Context, agentID string, conn *transport.SecureConn) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        var msg []byte // or define proper message type
        if err := conn.Receive(&msg); err != nil {
            if err != io.EOF {
                logger.ErrorCF("controller", "Receive error",
                    map[string]any{"agent_id": agentID, "error": err})
            }
            return
        }
        
        if s.onMessage != nil {
            s.onMessage(agentID, msg)
        }
    }
}

// SendToAgent sends a message to a specific agent
func (s *ControllerServer) SendToAgent(agentID string, msg interface{}) error {
    s.mu.RLock()
    conn, exists := s.secureConns[agentID]
    s.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("agent not connected: %s", agentID)
    }
    
    return conn.Send(msg)
}

// Broadcast sends a message to all connected agents
func (s *ControllerServer) Broadcast(msg interface{}) {
    s.mu.RLock()
    conns := make(map[string]*transport.SecureConn)
    for id, conn := range s.secureConns {
        conns[id] = conn
    }
    s.mu.RUnlock()
    
    for id, conn := range conns {
        if err := conn.Send(msg); err != nil {
            logger.ErrorCF("controller", "Broadcast failed",
                map[string]any{"agent_id": id, "error": err})
        }
    }
}

// GetConnectedAgents returns list of connected agent IDs
func (s *ControllerServer) GetConnectedAgents() []string {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    agents := make([]string, 0, len(s.secureConns))
    for id := range s.secureConns {
        agents = append(agents, id)
    }
    return agents
}
```

## Configuration

### Environment Variables

```bash
# Required
export PICOCLAW_TEAM_KEY="pk_team_..."

# Optional
export PICOCLAW_AGENT_ROLE="controller"  # or "worker"
export PICOCLAW_BIND_ADDRESS="0.0.0.0:5001"
export PICOCLAW_DISCOVERY_ENABLED="true"
export PICOCLAW_DISCOVERY_METHOD="mdns"
```

### Config File

```json
{
  "agent_network": {
    "enabled": true,
    "mode": "local",
    "bind_address": "0.0.0.0:5001",
    "discovery": {
      "enabled": true,
      "method": "mdns",
      "service_type": "_picoclaw-agent._tcp",
      "interval": 60
    },
    "security": {
      "team_key": "${TEAM_KEY}",
      "auth_method": "challenge_response",
      "encryption": "aes256_gcm",
      "challenge_ttl": 30
    }
  }
}
```

## Testing

```bash
# 1. Generate team key
go run scripts/gen-team-key.go
# Output: pk_team_aHR0cHM6Ly9naXRodWIuY29tL3NpX1c5V1J0...

# 2. Start controller
export TEAM_KEY="pk_team_..."
picoclaw gateway --agent-id controller-01

# 3. Start worker (in another terminal)
export TEAM_KEY="pk_team_..."
picoclaw agent --agent-id worker-01 --role worker

# 4. Verify connection
picoclaw status --show-agents
```

## Implementation Checklist

- [ ] mDNS discovery (advertise + browse)
- [ ] Team key generation and validation
- [ ] Challenge-response authentication
- [ ] AES-256-GCM encrypted channels
- [ ] Controller server
- [ ] Worker client
- [ ] Heartbeat mechanism
- [ ] Reconnection handling
- [ ] Configuration integration
- [ ] Tests

