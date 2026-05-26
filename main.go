package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const Version = "1.0.0"

// Exit codes
const (
	ExitSuccess           = 0
	ExitInvalidArgs       = 85
	ExitResourceNotFound  = 92
	ExitAPIError          = 100
	ExitInternalError     = 110
)

// GitLab API response structures
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

type Event struct {
	ID          int    `json:"id"`
	ProjectID   int    `json:"project_id"`
	ActionName  string `json:"action_name"`
	TargetTitle string `json:"target_title"`
	CreatedAt   string `json:"created_at"`
	PushData    struct {
		CommitTitle string `json:"commit_title"`
		CommitCount int    `json:"commit_count"`
	} `json:"push_data"`
}

type ActivityResponse struct {
	Version    string    `json:"version"`
	User       string    `json:"user"`
	Instance   string    `json:"instance"`
	Period     Period    `json:"period"`
	TotalEvents int      `json:"total_events"`
	Projects   []Project `json:"projects"`
}

type Period struct {
	Days int `json:"days"`
}

type Project struct {
	Name     string   `json:"name"`
	Events   int      `json:"events"`
	Activity []Event  `json:"activity"`
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(ExitInvalidArgs)
	}

	command := os.Args[1]

	switch command {
	case "me":
		handleMe()
	case "user":
		handleUser()
	case "version":
		handleVersion()
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printHelp()
		os.Exit(ExitInvalidArgs)
	}
}

func handleMe() {
	flags := flag.NewFlagSet("me", flag.ExitOnError)
	days := flags.Int("days", 7, "Number of days to look back")
	instance := flags.String("instance", "", "GitLab instance URL (default: auto-detect from token)")
	tokenPath := flags.String("token", "", "Path to token file (default: auto-detect)")
	project := flags.String("project", "", "Filter by project name")
	jsonOutput := flags.Bool("json", false, "Output in JSON format")
	since := flags.String("since", "", "Start date (YYYY-MM-DD)")
	until := flags.String("until", "", "End date (YYYY-MM-DD)")

	flags.Parse(os.Args[2:])

	// Auto-detect token and instance if not provided
	tokenFile, instanceURL, err := autoDetectToken(*instance, *tokenPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitAPIError)
	}

	token, err := os.ReadFile(tokenFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading token file: %v\n", err)
		os.Exit(ExitAPIError)
	}

	// Get current user info
	user, err := getCurrentUser(instanceURL, string(token))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user info: %v\n", err)
		os.Exit(ExitAPIError)
	}

	// Get events
	events, err := getEvents(instanceURL, string(token), user.ID, *days, *since, *until, *project)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting events: %v\n", err)
		os.Exit(ExitAPIError)
	}

	// Output
	if *jsonOutput {
		outputJSON(user.Username, instanceURL, events, *days)
	} else {
	 outputText(events)
	}
}

func handleUser() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Error: username required\n")
		printHelp()
		os.Exit(ExitInvalidArgs)
	}

	username := os.Args[2]

	flags := flag.NewFlagSet("user", flag.ExitOnError)
	days := flags.Int("days", 7, "Number of days to look back")
	instance := flags.String("instance", "", "GitLab instance URL (default: auto-detect from token)")
	tokenPath := flags.String("token", "", "Path to token file (default: auto-detect)")
	project := flags.String("project", "", "Filter by project name")
	jsonOutput := flags.Bool("json", false, "Output in JSON format")
	since := flags.String("since", "", "Start date (YYYY-MM-DD)")
	until := flags.String("until", "", "End date (YYYY-MM-DD)")

	flags.Parse(os.Args[3:])

	// Auto-detect token and instance if not provided
	tokenFile, instanceURL, err := autoDetectToken(*instance, *tokenPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitAPIError)
	}

	token, err := os.ReadFile(tokenFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading token file: %v\n", err)
		os.Exit(ExitAPIError)
	}

	// Get user ID from username
	userId, err := getUserIdByUsername(instanceURL, string(token), username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user ID: %v\n", err)
		os.Exit(ExitResourceNotFound)
	}

	// Get events
	events, err := getEvents(instanceURL, string(token), userId, *days, *since, *until, *project)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting events: %v\n", err)
		os.Exit(ExitAPIError)
	}

	// Output
	if *jsonOutput {
		outputJSON(username, instanceURL, events, *days)
	} else {
		outputText(events)
	}
}

func autoDetectToken(instance, tokenPath string) (string, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("could not determine home directory")
	}

	gitlabDir := filepath.Join(homeDir, ".gitlab")

	// If token path is provided, use it
	if tokenPath != "" {
		if instance == "" {
			return "", "", fmt.Errorf("--instance required when --token is provided")
		}
		return tokenPath, instance, nil
	}

	// Auto-detect based on instance
	if instance != "" {
		// Map known instances to token files
		if strings.Contains(instance, "gitlab.com") {
			tokenFile := filepath.Join(gitlabDir, "jar-token")
			return tokenFile, instance, nil
		}
		if strings.Contains(instance, "git.geored.fr") {
			tokenFile := filepath.Join(gitlabDir, "geored")
			return tokenFile, instance, nil
		}
		// For custom instances, default to jar-token
		tokenFile := filepath.Join(gitlabDir, "jar-token")
		return tokenFile, instance, nil
	}

	// Try to auto-detect from available token files
	if _, err := os.Stat(filepath.Join(gitlabDir, "geored")); err == nil {
		return filepath.Join(gitlabDir, "geored"), "https://git.geored.fr", nil
	}
	if _, err := os.Stat(filepath.Join(gitlabDir, "jar-token")); err == nil {
		return filepath.Join(gitlabDir, "jar-token"), "https://gitlab.com", nil
	}

	return "", "", fmt.Errorf("no token file found in ~/.gitlab/ (please specify --token)")
}

func getCurrentUser(instanceURL, token string) (*User, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", instanceURL+"/api/v4/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("PRIVATE-TOKEN", strings.TrimSpace(token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func getUserIdByUsername(instanceURL, token, username string) (int, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", instanceURL+"/api/v4/users?username="+username, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Add("PRIVATE-TOKEN", strings.TrimSpace(token))

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var users []User
	err = json.Unmarshal(body, &users)
	if err != nil {
		return 0, err
	}

	if len(users) == 0 {
		return 0, fmt.Errorf("user not found")
	}

	return users[0].ID, nil
}

func getEvents(instanceURL, token string, userId int, days int, since, until, projectFilter string) ([]Event, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/v4/users/%d/events?per_page=100", instanceURL, userId)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("PRIVATE-TOKEN", strings.TrimSpace(token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var events []Event
	err = json.Unmarshal(body, &events)
	if err != nil {
		return nil, err
	}

	// Filter by date
	var filteredEvents []Event
	now := time.Now()
	var startDate, endDate time.Time

	if since != "" {
		startDate, err = time.Parse("2006-01-02", since)
		if err != nil {
			return nil, fmt.Errorf("invalid since date format: %v", err)
		}
	} else {
		startDate = now.AddDate(0, 0, -days)
	}

	if until != "" {
		endDate, err = time.Parse("2006-01-02", until)
		if err != nil {
			return nil, fmt.Errorf("invalid until date format: %v", err)
		}
	} else {
		endDate = now
	}

	for _, event := range events {
		eventTime, err := time.Parse(time.RFC3339, event.CreatedAt)
		if err != nil {
			continue
		}

		if eventTime.After(startDate) && eventTime.Before(endDate.AddDate(0, 0, 1)) {
			// Filter by project if specified
			if projectFilter != "" && event.TargetTitle != projectFilter {
				continue
			}
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents, nil
}

func outputText(events []Event) {
	if len(events) == 0 {
		fmt.Println("No activity found in the specified period.")
		return
	}

	// Group by project
	byProject := make(map[string][]Event)
	for _, event := range events {
		project := event.TargetTitle
		if project == "" {
			project = "Unknown"
		}
		byProject[project] = append(byProject[project], event)
	}

	fmt.Printf("Total events: %d\n\n", len(events))

	for project, projectEvents := range byProject {
		fmt.Printf("%s (%d events):\n", project, len(projectEvents))
		for _, event := range projectEvents {
			eventTime, _ := time.Parse(time.RFC3339, event.CreatedAt)
			fmt.Printf("  %s - %s\n", eventTime.Format("2006-01-02 15:04:05"), event.ActionName)
			if event.PushData.CommitTitle != "" {
				fmt.Printf("    %s\n", event.PushData.CommitTitle)
			}
		}
		fmt.Println()
	}
}

func outputJSON(username, instanceURL string, events []Event, days int) {
	// Group by project
	byProject := make(map[string][]Event)
	for _, event := range events {
		project := event.TargetTitle
		if project == "" {
			project = "Unknown"
		}
		byProject[project] = append(byProject[project], event)
	}

	var projects []Project
	for project, projectEvents := range byProject {
		projects = append(projects, Project{
			Name:     project,
			Events:   len(projectEvents),
			Activity: projectEvents,
		})
	}

	response := ActivityResponse{
		Version:     "1.0",
		User:        username,
		Instance:    instanceURL,
		Period:      Period{Days: days},
		TotalEvents: len(events),
		Projects:    projects,
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(ExitInternalError)
	}

	fmt.Println(string(jsonData))
}

func handleVersion() {
	fmt.Printf("gitlab-activity-cli v%s\n", Version)
}

func printHelp() {
	fmt.Println("gitlab-activity-cli - Retrieve GitLab user activity for agents")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gitlab-activity-cli me [options]")
	fmt.Println("  gitlab-activity-cli user <username> [options]")
	fmt.Println("  gitlab-activity-cli version")
	fmt.Println("  gitlab-activity-cli help")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  me       Get current user's activity")
	fmt.Println("  user     Get specific user's activity")
	fmt.Println("  version  Show version information")
	fmt.Println("  help     Show this help message")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -days int          Number of days to look back (default: 7)")
	fmt.Println("  -instance string   GitLab instance URL (default: auto-detect)")
	fmt.Println("  -token string      Path to token file (default: auto-detect)")
	fmt.Println("  -project string    Filter by project name")
	fmt.Println("  -since string      Start date (YYYY-MM-DD)")
	fmt.Println("  -until string      End date (YYYY-MM-DD)")
	fmt.Println("  -json              Output in JSON format")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gitlab-activity-cli me")
	fmt.Println("  gitlab-activity-cli me --days 2 --instance https://git.geored.fr")
	fmt.Println("  gitlab-activity-cli me --project georedv3 --json")
	fmt.Println("  gitlab-activity-cli user jarancibia --days 3")
	fmt.Println()
	fmt.Println("Exit Codes:")
	fmt.Println("  0    Success")
	fmt.Println("  85   Invalid arguments")
	fmt.Println("  92   Resource not found")
	fmt.Println("  100  API error")
	fmt.Println("  110  Internal error")
}