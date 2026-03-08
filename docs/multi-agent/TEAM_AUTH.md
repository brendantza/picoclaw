# Team-Based Authentication for Multi-Agent Systems

## Overview

This document describes a secure team-based authentication system for PicoClaw multi-agent deployments. Agents use **team keys** to authenticate with controllers, preventing unauthorized agents from joining teams.

## Core Concepts

### Team Key

A team key is a shared secret that identifies and authenticates members of a specific team.

```
Team Key Format: pk_team_<base64url_encoded_32_bytes>
Example: pk_team_aHR0cHM6Ly9naXRodWIuY29tL3NpX1c5V1J0...
```

Properties:
- **High entropy**: 256-bit random value
- **Shared secret**: All team members know the key
- **Never transmitted in plaintext**: Used for challenge-response
- **Rotatable**: Teams can rotate keys periodically

### Authentication Flow

```
┌──────────────┐                    ┌──────────────┐
│   Worker     │                    │  Controller  │
│   Agent      │                    │   Agent      │
└──────┬───────┘                    └──────┬───────┘
       │                                   │
       │ 1. mDNS Discovery                 │
       │    (advertise team hash)          │
       │◄─────────────────────────────────►│
       │                                   │
       │ 2. TCP Connection                 │
       │ ─────────────────────────────────>│
       │                                   │
       │ 3. Challenge-Response Auth        │
       │    Controller sends nonce         │
       │<───────────────────────────────── │
       │                                   │
       │    Worker signs with team key     │
       │ ─────────────────────────────────>│
       │                                   │
       │    Controller verifies            │
       │<───────────────────────────────── │
       │                                   │
       │ 4. Secure Channel Established     │
       │    (session key derived)          │
       │◄═════════════════════════════════►│
       │                                   │
       │ 5. Heartbeat (encrypted)          │
       │◄─────────────────────────────────►│
```

## Detailed Protocol

### 1. mDNS Advertisement

Agents advertise their presence without revealing the team key.

**Service Type:** `_picoclaw-agent._tcp`

**TXT Records:**
```
team_hash=<SHA256(team_key)[0:16]>  # First 16 chars of team key hash
role=controller|worker|autonomous
capabilities=api,frontend,backend
version=1.0.0
agent_id=dev-worker-01
```

**Go Implementation:**
```go
package discovery

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    
    "github.com/grandcat/zeroconf"
)

type AgentAdvertisement struct {
    AgentID      string
    Role         string
    TeamHash     string
    Capabilities []string
    Version      string
    Port         int
}

// ComputeTeamHash creates a hash of the team key for public advertisement
func ComputeTeamHash(teamKey string) string {
    hash := sha256.Sum256([]byte(teamKey))
    return hex.EncodeToString(hash[:8]) // First 8 bytes = 16 hex chars
}

// Advertise starts mDNS advertisement
func (d *MDNSDiscovery) Advertise(agent AgentInfo, teamKey string) error {
    teamHash := ComputeTeamHash(teamKey)
    
    txtRecords := []string{
        fmt.Sprintf("team_hash=%s", teamHash),
        fmt.Sprintf("role=%s", agent.Role),
        fmt.Sprintf("agent_id=%s", agent.ID),
        fmt.Sprintf("version=%s", agent.Version),
        fmt.Sprintf("capabilities=%s", strings.Join(agent.Capabilities, ",")),
    }
    
    server, err := zeroconf.Register(
        agent.ID,                    // instance name
        "_picoclaw-agent._tcp",      // service type
        "local.",                    // domain
        agent.Port,                  // port
        txtRecords,                  // TXT records
        nil,                         // interfaces
    )
    if err != nil {
        return err
    }
    
    d.server = server
    return nil
}

// Discover finds agents matching our team
func (d *MDNSDiscovery) Discover(teamKey string) ([]AgentInfo, error) {
    teamHash := ComputeTeamHash(teamKey)
    
    resolver, err := zeroconf.NewResolver(nil)
    if err != nil {
        return nil, err
    }
    
    entries := make(chan *zeroconf.ServiceEntry)
    var agents []AgentInfo
    
    go func() {
        for entry := range entries {
            // Check team hash matches
            if matchesTeam(entry, teamHash) {
                agents = append(agents, parseAgentInfo(entry))
            }
        }
    }()
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    err = resolver.Browse(ctx, "_picoclaw-agent._tcp", "local.", entries)
    if err != nil {
        return nil, err
    }
    
    <-ctx.Done()
    return agents, nil
}
```

### 2. Challenge-Response Authentication

After TCP connection, perform secure authentication.

**Protocol:**
```
Worker                                          Controller
------                                          ---------
                                                Generate nonce: N (32 random bytes)
                                                
                ---- Auth Challenge ---->
                { "type": "challenge",
                  "nonce": "base64(N)",
                  "timestamp": 1234567890 }
                
Sign challenge:
  S = HMAC-SHA256(team_key, nonce || timestamp)
                
                <--- Auth Response ----
                { "type": "response",
                  "agent_id": "worker-01",
                  "signature": "base64(S)",
                  "capabilities": [...],
                  "public_key": "..." }
                
                                                Verify:
                                                  1. timestamp not expired
                                                  2. S == HMAC(team_key, nonce || ts)
                                                  3. team_hash matches
                                                  
                <--- Auth Success ----
                { "type": "auth_success",
                  "session_id": "sess_...",
                  "expires_at": 1234567890 }
                
                                                Derive session key:
                                                  SK = HKDF(team_key, nonce || session_id)
```

**Go Implementation:**
```go
package auth

import (
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "time"
    
    "golang.org/x/crypto/hkdf"
)

const (
    ChallengeTTL = 30 * time.Second
    NonceSize    = 32
)

// TeamAuthenticator handles team-based authentication
type TeamAuthenticator struct {
    teamKey string
}

func NewTeamAuthenticator(teamKey string) *TeamAuthenticator {
    return &TeamAuthenticator{teamKey: teamKey}
}

// Challenge represents an authentication challenge
type Challenge struct {
    Type      string `json:"type"`
    Nonce     string `json:"nonce"`     // base64 encoded random bytes
    Timestamp int64  `json:"timestamp"` // Unix timestamp
}

// AuthResponse represents worker's response to challenge
type AuthResponse struct {
    Type         string   `json:"type"`
    AgentID      string   `json:"agent_id"`
    Signature    string   `json:"signature"`    // HMAC of challenge
    Capabilities []string `json:"capabilities"`
    Role         string   `json:"role"`
}

// GenerateChallenge creates a new authentication challenge
func (a *TeamAuthenticator) GenerateChallenge() (*Challenge, error) {
    nonce := make([]byte, NonceSize)
    if _, err := rand.Read(nonce); err != nil {
        return nil, err
    }
    
    return &Challenge{
        Type:      "challenge",
        Nonce:     base64.StdEncoding.EncodeToString(nonce),
        Timestamp: time.Now().Unix(),
    }, nil
}

// CreateResponse creates an authentication response
func (a *TeamAuthenticator) CreateResponse(
    challenge *Challenge,
    agentID string,
    capabilities []string,
    role string,
) (*AuthResponse, error) {
    // Create signature
    sig, err := a.signChallenge(challenge)
    if err != nil {
        return nil, err
    }
    
    return &AuthResponse{
        Type:         "response",
        AgentID:      agentID,
        Signature:    base64.StdEncoding.EncodeToString(sig),
        Capabilities: capabilities,
        Role:         role,
    }, nil
}

// VerifyResponse verifies a worker's authentication response
func (a *TeamAuthenticator) VerifyResponse(
    challenge *Challenge,
    response *AuthResponse,
) error {
    // Check timestamp
    challengeTime := time.Unix(challenge.Timestamp, 0)
    if time.Since(challengeTime) > ChallengeTTL {
        return fmt.Errorf("challenge expired")
    }
    
    // Verify signature
    expectedSig, err := a.signChallenge(challenge)
    if err != nil {
        return err
    }
    
    actualSig, err := base64.StdEncoding.DecodeString(response.Signature)
    if err != nil {
        return fmt.Errorf("invalid signature encoding: %w", err)
    }
    
    if !hmac.Equal(expectedSig, actualSig) {
        return fmt.Errorf("invalid signature")
    }
    
    return nil
}

// signChallenge creates HMAC signature for challenge
func (a *TeamAuthenticator) signChallenge(challenge *Challenge) ([]byte, error) {
    nonce, err := base64.StdEncoding.DecodeString(challenge.Nonce)
    if err != nil {
        return nil, err
    }
    
    // Data to sign: nonce || timestamp
    data := make([]byte, 0, len(nonce)+8)
    data = append(data, nonce...)
    data = append(data, []byte(fmt.Sprintf("%d", challenge.Timestamp))...)
    
    h := hmac.New(sha256.New, []byte(a.teamKey))
    h.Write(data)
    return h.Sum(nil), nil
}

// DeriveSessionKey creates a session encryption key
func (a *TeamAuthenticator) DeriveSessionKey(
    challenge *Challenge,
    sessionID string,
) ([]byte, error) {
    nonce, err := base64.StdEncoding.DecodeString(challenge.Nonce)
    if err != nil {
        return nil, err
    }
    
    // Use HKDF to derive key
    hkdfReader := hkdf.New(
        sha256.New,
        []byte(a.teamKey),
        nonce,                              // salt
        []byte("picoclaw-session-"+sessionID), // info
    )
    
    sessionKey := make([]byte, 32)
    if _, err := hkdfReader.Read(sessionKey); err != nil {
        return nil, err
    }
    
    return sessionKey, nil
}
```

### 3. Secure Communication

After authentication, derive session keys for encrypted communication.

```go
package secure

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/json"
    "io"
)

// SecureChannel provides encrypted communication
type SecureChannel struct {
    sessionKey []byte
    sendCipher cipher.AEAD
    recvCipher cipher.AEAD
    sendNonce  uint64
    recvNonce  uint64
}

// NewSecureChannel creates encrypted channel from session key
func NewSecureChannel(sessionKey []byte) (*SecureChannel, error) {
    block, err := aes.NewCipher(sessionKey)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    return &SecureChannel{
        sessionKey: sessionKey,
        sendCipher: gcm,
        recvCipher: gcm,
    }, nil
}

// Send encrypts and sends a message
func (c *SecureChannel) Send(conn net.Conn, msg interface{}) error {
    plaintext, err := json.Marshal(msg)
    if err != nil {
        return err
    }
    
    nonce := make([]byte, c.sendCipher.NonceSize())
    binary.BigEndian.PutUint64(nonce, c.sendNonce)
    c.sendNonce++
    
    // Append random bytes for remaining nonce space
    if _, err := rand.Read(nonce[8:]); err != nil {
        return err
    }
    
    ciphertext := c.sendCipher.Seal(nonce, nonce, plaintext, nil)
    
    // Send length prefix + ciphertext
    lengthBuf := make([]byte, 4)
    binary.BigEndian.PutUint32(lengthBuf, uint32(len(ciphertext)))
    
    if _, err := conn.Write(lengthBuf); err != nil {
        return err
    }
    if _, err := conn.Write(ciphertext); err != nil {
        return err
    }
    
    return nil
}

// Receive decrypts a received message
func (c *SecureChannel) Receive(conn net.Conn, msg interface{}) error {
    // Read length
    lengthBuf := make([]byte, 4)
    if _, err := io.ReadFull(conn, lengthBuf); err != nil {
        return err
    }
    length := binary.BigEndian.Uint32(lengthBuf)
    
    if length > MaxMessageSize {
        return fmt.Errorf("message too large: %d", length)
    }
    
    // Read ciphertext
    ciphertext := make([]byte, length)
    if _, err := io.ReadFull(conn, ciphertext); err != nil {
        return err
    }
    
    // Extract nonce
    nonceSize := c.recvCipher.NonceSize()
    if len(ciphertext) < nonceSize {
        return fmt.Errorf("ciphertext too short")
    }
    nonce := ciphertext[:nonceSize]
    ciphertext = ciphertext[nonceSize:]
    
    // Decrypt
    plaintext, err := c.recvCipher.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return fmt.Errorf("decryption failed: %w", err)
    }
    
    return json.Unmarshal(plaintext, msg)
}
```

## Configuration

### Controller Configuration

```json
{
  "agent_network": {
    "enabled": true,
    "mode": "local",
    "discovery": {
      "enabled": true,
      "method": "mdns",
      "service_type": "_picoclaw-agent._tcp"
    },
    "security": {
      "team_key": "${TEAM_KEY}",
      "auth_method": "challenge_response",
      "encryption": "aes256_gcm",
      "challenge_ttl": 30
    },
    "controller": {
      "bind_address": "0.0.0.0:5001",
      "max_workers": 10,
      "require_auth": true,
      "allowed_roles": ["worker", "autonomous"]
    }
  }
}
```

### Worker Configuration

```json
{
  "agent_network": {
    "enabled": true,
    "mode": "local",
    "discovery": {
      "enabled": true,
      "method": "mdns"
    },
    "security": {
      "team_key": "${TEAM_KEY}",
      "auth_method": "challenge_response"
    },
    "worker": {
      "role": "worker",
      "specialization": "backend",
      "auto_register": true,
      "heartbeat_interval": 30,
      "controller_discovery": "mdns"
    }
  }
}
```

### Environment Variable Setup

```bash
# Generate a secure team key
export TEAM_KEY=$(openssl rand -base64 32)

# Or use a predefined key for your team
export TEAM_KEY="pk_team_aHR0cHM6Ly9naXRodWIuY29tL3NpX1c5V1J0..."

# Start controller
picoclaw gateway --agent-id controller-01

# Start workers (in separate terminals)
picoclaw agent --agent-id backend-dev-01 --role worker
picoclaw agent --agent-id frontend-dev-01 --role worker
```

## Security Best Practices

### 1. Team Key Management

```go
// pkg/security/teamkey.go

package security

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "strings"
)

const TeamKeyPrefix = "pk_team_"

// GenerateTeamKey creates a new cryptographically secure team key
func GenerateTeamKey() (string, error) {
    key := make([]byte, 32)
    if _, err := rand.Read(key); err != nil {
        return "", err
    }
    return TeamKeyPrefix + base64.URLEncoding.EncodeToString(key), nil
}

// ValidateTeamKey checks if a team key is valid format
func ValidateTeamKey(key string) error {
    if !strings.HasPrefix(key, TeamKeyPrefix) {
        return fmt.Errorf("team key must start with %s", TeamKeyPrefix)
    }
    
    encoded := strings.TrimPrefix(key, TeamKeyPrefix)
    decoded, err := base64.URLEncoding.DecodeString(encoded)
    if err != nil {
        return fmt.Errorf("invalid team key encoding: %w", err)
    }
    
    if len(decoded) != 32 {
        return fmt.Errorf("team key must be 32 bytes, got %d", len(decoded))
    }
    
    return nil
}

// RedactTeamKey masks team key for logging
func RedactTeamKey(key string) string {
    if len(key) < 20 {
        return "***"
    }
    return key[:10] + "..." + key[len(key)-5:]
}
```

### 2. Key Rotation

```go
// Support for rotating team keys without breaking existing sessions

type KeyManager struct {
    currentKey string
    previousKey string
    rotationTime time.Time
}

// Verify accepts both current and previous key during rotation window
func (km *KeyManager) Verify(challenge *Challenge, response *AuthResponse) error {
    // Try current key
    auth := NewTeamAuthenticator(km.currentKey)
    if err := auth.VerifyResponse(challenge, response); err == nil {
        return nil
    }
    
    // Try previous key (during rotation window)
    if km.previousKey != "" && time.Since(km.rotationTime) < RotationWindow {
        auth = NewTeamAuthenticator(km.previousKey)
        return auth.VerifyResponse(challenge, response)
    }
    
    return fmt.Errorf("authentication failed")
}
```

### 3. Audit Logging

```go
// Log all authentication attempts

func (c *Controller) logAuthAttempt(workerID string, success bool, err error) {
    logger.InfoCF("auth", "Authentication attempt",
        map[string]any{
            "worker_id": workerID,
            "success":   success,
            "error":     err,
            "timestamp": time.Now().UTC(),
        })
}
```

## Network Segmentation

For additional security, support network-level segmentation:

```json
{
  "agent_network": {
    "security": {
      "team_key": "${TEAM_KEY}",
      "network_acl": {
        "allowed_subnets": ["10.0.0.0/8", "192.168.1.0/24"],
        "blocked_subnets": ["10.0.5.0/24"],
        "require_mtls": true
      }
    }
  }
}
```

## Testing

```bash
# Generate test team key
export TEAM_KEY=$(go run scripts/gen-team-key.go)
echo "Team Key: $TEAM_KEY"

# Test discovery
picoclaw agent --discover --team-key "$TEAM_KEY"

# Test authentication
picoclaw agent --test-auth --team-key "$TEAM_KEY" --controller localhost:5001
```

## Comparison with Alternatives

| Method | Pros | Cons | Best For |
|--------|------|------|----------|
| **Team Key** (proposed) | Simple, no central authority, easy rotation | Shared secret risk | Small-medium teams |
| **mTLS Certs** | Strong auth, individual identities | Complex PKI | Enterprise |
| **OAuth/OIDC** | Centralized auth, SSO | Requires IdP | Corporate environments |
| **API Tokens** | Simple, revocable | Token management overhead | Cloud deployments |

## Summary

The team key approach provides:
- ✅ **Simple setup** - Single shared secret
- ✅ **No external dependencies** - No CA or IdP required
- ✅ **Automatic discovery** - mDNS finds team members
- ✅ **Secure** - Challenge-response prevents replay attacks
- ✅ **Encrypted** - Session keys for all communication
- ✅ **Auditable** - All auth attempts logged

