// +build ignore

// gen-team-key generates a secure team key for multi-agent authentication
// Usage: go run scripts/gen-team-key.go
package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
)

const teamKeyPrefix = "pk_team_"

func main() {
	// Generate 32 random bytes
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating random key: %v\n", err)
		os.Exit(1)
	}

	// Encode to URL-safe base64
	encoded := base64.URLEncoding.EncodeToString(key)

	// Create team key
	teamKey := teamKeyPrefix + encoded

	fmt.Println("Generated Team Key:")
	fmt.Println("===================")
	fmt.Println(teamKey)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  export TEAM_KEY=\"" + teamKey + "\"")
	fmt.Println()
	fmt.Println("Or add to your config.json:")
	fmt.Printf(`  "agent_network": {`+"\n")
	fmt.Printf(`    "security": {`+"\n")
	fmt.Printf(`      "team_key": "%s"`+"\n", teamKey)
	fmt.Printf(`    }`+"\n")
	fmt.Printf(`  }`+"\n")
}
