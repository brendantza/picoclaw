// PicoClaw - Team task command
// License: MIT

package team

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

func taskCmd() *cobra.Command {
	var gateway string
	var teamID string
	var agentID string
	var taskType string
	var title string
	var priority int

	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage team tasks",
		Long:  "Create, list, and manage tasks for agents in a team.",
	}

	// Add create subcommand
	createCmd := &cobra.Command{
		Use:   "create [command]",
		Short: "Create a new task for an agent",
		Long: `Create a task to be executed by an agent.

Examples:
  # Create a shell task
  picoclaw team task create "ls -la" --agent dev-agent-docker

  # Create an echo task
  picoclaw team task create --type echo --title "Hello" --agent dev-agent-docker

  # Create a shell task with custom title
  picoclaw team task create "uname -a" --title "System Info" --agent dev-agent-docker
`,
		Args: cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get command from args or stdin
			command := ""
			if len(args) > 0 {
				command = args[0]
			}

			// Load team config
			teamConfig, err := loadTeamConfig(teamID)
			if err != nil {
				return fmt.Errorf("failed to load team config: %w", err)
			}

			// Use configured gateway if not specified
			if gateway == "" {
				gateway = teamConfig.GatewayAddr
			}
			if gateway == "" {
				return fmt.Errorf("gateway address required (use --gateway or ensure team is joined)")
			}

			// Use configured team ID if not specified
			if teamID == "" {
				teamID = teamConfig.TeamID
			}
			if teamID == "" {
				return fmt.Errorf("team ID required (use --team or ensure team is joined)")
			}

			// Build task request
			req := createTaskRequest{
				AgentID:     agentID,
				Type:        taskType,
				Title:       title,
				Description: "",
				Priority:    priority,
				Payload: map[string]any{
					"command": command,
				},
			}

			if title == "" && command != "" {
				req.Title = command
			}

			// Create task
			task, err := createTask(gateway, teamID, req)
			if err != nil {
				return fmt.Errorf("failed to create task: %w", err)
			}

			fmt.Printf("✓ Task created: %s\n", task.ID)
			fmt.Printf("  Type: %s\n", task.Type)
			fmt.Printf("  Title: %s\n", task.Title)
			fmt.Printf("  Agent: %s\n", task.AgentID)
			fmt.Printf("  Status: %s\n", task.Status)

			return nil
		},
	}

	createCmd.Flags().StringVar(&gateway, "gateway", "", "Gateway address (e.g., http://192.168.6.122:18790)")
	createCmd.Flags().StringVar(&teamID, "team", "", "Team ID")
	createCmd.Flags().StringVar(&agentID, "agent", "", "Target agent ID (required)")
	createCmd.Flags().StringVar(&taskType, "type", "shell", "Task type (shell, echo)")
	createCmd.Flags().StringVar(&title, "title", "", "Task title")
	createCmd.Flags().IntVar(&priority, "priority", 5, "Task priority (1-10)")
	_ = createCmd.MarkFlagRequired("agent")

	// Add list subcommand
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks for a team",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load team config
			teamConfig, err := loadTeamConfig(teamID)
			if err != nil {
				return fmt.Errorf("failed to load team config: %w", err)
			}

			if gateway == "" {
				gateway = teamConfig.GatewayAddr
			}
			if teamID == "" {
				teamID = teamConfig.TeamID
			}

			tasks, err := listTasks(gateway, teamID)
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}

			if len(tasks) == 0 {
				fmt.Println("No tasks found")
				return nil
			}

			fmt.Printf("Tasks for team %s:\n\n", teamID)
			fmt.Printf("%-20s %-10s %-12s %-20s %s\n", "ID", "TYPE", "STATUS", "AGENT", "TITLE")
			fmt.Println(string(make([]byte, 80)))
			for _, t := range tasks {
				id := t.ID
				if len(id) > 18 {
					id = id[:15] + "..."
				}
				fmt.Printf("%-20s %-10s %-12s %-20s %s\n", id, t.Type, t.Status, t.AgentID, t.Title)
			}

			return nil
		},
	}

	listCmd.Flags().StringVar(&gateway, "gateway", "", "Gateway address")
	listCmd.Flags().StringVar(&teamID, "team", "", "Team ID")

	// Add get subcommand
	getCmd := &cobra.Command{
		Use:   "get [task-id]",
		Short: "Get task details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			teamConfig, err := loadTeamConfig(teamID)
			if err != nil {
				return fmt.Errorf("failed to load team config: %w", err)
			}

			if gateway == "" {
				gateway = teamConfig.GatewayAddr
			}
			if teamID == "" {
				teamID = teamConfig.TeamID
			}

			task, err := getTask(gateway, teamID, taskID)
			if err != nil {
				return fmt.Errorf("failed to get task: %w", err)
			}

			fmt.Printf("Task: %s\n", task.ID)
			fmt.Printf("  Type: %s\n", task.Type)
			fmt.Printf("  Status: %s\n", task.Status)
			fmt.Printf("  Title: %s\n", task.Title)
			fmt.Printf("  Agent: %s\n", task.AgentID)
			fmt.Printf("  Priority: %d\n", task.Priority)
			fmt.Printf("  Created: %s\n", task.CreatedAt.Format(time.RFC3339))
			if task.CompletedAt != nil {
				fmt.Printf("  Completed: %s\n", task.CompletedAt.Format(time.RFC3339))
			}
			if task.Result != nil {
				fmt.Printf("  Result: %v\n", task.Result)
			}
			if task.Error != "" {
				fmt.Printf("  Error: %s\n", task.Error)
			}

			return nil
		},
	}

	getCmd.Flags().StringVar(&gateway, "gateway", "", "Gateway address")
	getCmd.Flags().StringVar(&teamID, "team", "", "Team ID")

	cmd.AddCommand(createCmd)
	cmd.AddCommand(listCmd)
	cmd.AddCommand(getCmd)

	return cmd
}

type teamConfig struct {
	TeamID      string `json:"team_id"`
	GatewayAddr string `json:"gateway_addr"`
}

func loadTeamConfig(teamID string) (*teamConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Try specific team config
	if teamID != "" {
		configFile := filepath.Join(home, ".picoclaw", "agent_teams", teamID+".json")
		if data, err := os.ReadFile(configFile); err == nil {
			var config teamConfig
			if err := json.Unmarshal(data, &config); err == nil {
				return &config, nil
			}
		}
	}

	// Try to find any joined team
	teamsDir := filepath.Join(home, ".picoclaw", "agent_teams")
	entries, err := os.ReadDir(teamsDir)
	if err == nil && len(entries) > 0 {
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
				configFile := filepath.Join(teamsDir, entry.Name())
				if data, err := os.ReadFile(configFile); err == nil {
					var config teamConfig
					if err := json.Unmarshal(data, &config); err == nil {
						return &config, nil
					}
				}
			}
		}
	}

	return &teamConfig{}, nil
}

type createTaskRequest struct {
	AgentID     string         `json:"agent_id"`
	Type        string         `json:"type"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Priority    int            `json:"priority"`
	Payload     map[string]any `json:"payload"`
}

type taskResponse struct {
	ID          string         `json:"id"`
	TeamID      string         `json:"team_id"`
	AgentID     string         `json:"agent_id"`
	Type        string         `json:"type"`
	Status      string         `json:"status"`
	Title       string         `json:"title"`
	Priority    int            `json:"priority"`
	Payload     map[string]any `json:"payload"`
	Result      map[string]any `json:"result"`
	Error       string         `json:"error"`
	CreatedAt   time.Time      `json:"created_at"`
	CompletedAt *time.Time     `json:"completed_at"`
}

func createTask(gateway, teamID string, req createTaskRequest) (*taskResponse, error) {
	url := fmt.Sprintf("%s/api/teams/%s/tasks", gateway, teamID)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var task taskResponse
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}

	return &task, nil
}

func listTasks(gateway, teamID string) ([]taskResponse, error) {
	url := fmt.Sprintf("%s/api/teams/%s/tasks", gateway, teamID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var tasks []taskResponse
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func getTask(gateway, teamID, taskID string) (*taskResponse, error) {
	url := fmt.Sprintf("%s/api/teams/%s/tasks/%s", gateway, teamID, taskID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var task taskResponse
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}

	return &task, nil
}
