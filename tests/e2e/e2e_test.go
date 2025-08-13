package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

// GiteaUser represents a Gitea user
type GiteaUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

// GiteaRepo represents a Gitea repository
type GiteaRepo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
}

// GiteaIssue represents a Gitea issue
type GiteaIssue struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// E2ETest manages the end-to-end testing process
type E2ETest struct {
	giteaURL           string
	accessToken        string
	username           string
	repoName           string
	containerName      string
	dataVolumeName     string
	giteaContainerName string
	httpClient         *http.Client
}

func main() {
	logger.Info("Starting Gitea backup/restore E2E test")

	// Get configuration from environment variables
	giteaURL := os.Getenv("GITEA_URL")
	if giteaURL == "" {
		giteaURL = "http://localhost:3000"
	}

	containerName := os.Getenv("CONTAINER_NAME")
	if containerName == "" {
		containerName = "gitea-backup-e2e"
	}

	dataVolumeName := os.Getenv("DATA_VOLUME_NAME")
	if dataVolumeName == "" {
		dataVolumeName = "gitea-backup-restore-process_gitea-data"
	}

	giteaContainerName := os.Getenv("GITEA_CONTAINER_NAME")
	if giteaContainerName == "" {
		giteaContainerName = "gitea-e2e"
	}

	test := &E2ETest{
		giteaURL:           giteaURL,
		username:           "e2euser",
		repoName:           "e2e-test-repo",
		containerName:      containerName,
		dataVolumeName:     dataVolumeName,
		giteaContainerName: giteaContainerName,
		httpClient:         &http.Client{Timeout: 30 * time.Second},
	}

	if err := test.runE2ETest(); err != nil {
		logger.Errorf("E2E test failed: %v", err)
		os.Exit(1)
	}

	logger.Info("E2E test completed successfully")
}

func (t *E2ETest) runE2ETest() error {
	// Step 1: Wait for services to be ready
	if err := t.waitForServices(); err != nil {
		return fmt.Errorf("failed to wait for services: %w", err)
	}

	// Step 2: Initialize Gitea and create user
	if err := t.initializeGitea(); err != nil {
		return fmt.Errorf("failed to initialize Gitea: %w", err)
	}

	// Step 3: Create test data (repository and issue)
	if err := t.createTestData(); err != nil {
		return fmt.Errorf("failed to create test data: %w", err)
	}

	// Step 4: Perform backup
	if err := t.performBackup(); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// Step 5: Simulate failure by clearing data
	if err := t.simulateDataLoss(); err != nil {
		return fmt.Errorf("failed to simulate data loss: %w", err)
	}

	// Step 6: Perform restore
	if err := t.performRestore(); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	// Step 7: Verify restoration
	if err := t.verifyRestoration(); err != nil {
		return fmt.Errorf("restoration verification failed: %w", err)
	}

	return nil
}

func (t *E2ETest) waitForServices() error {
	logger.Info("Waiting for services to be ready...")

	// Wait for Gitea
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		resp, err := t.httpClient.Get(t.giteaURL + "/api/healthz")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			logger.Info("Gitea is ready")
			break
		}
		if resp != nil {
			resp.Body.Close()
		}

		if i == maxRetries-1 {
			return fmt.Errorf("Gitea not ready after %d attempts", maxRetries)
		}

		logger.Debugf("Waiting for Gitea... attempt %d/%d", i+1, maxRetries)
		time.Sleep(10 * time.Second)
	}

	// Wait for MinIO (check via Gitea backup container)
	time.Sleep(5 * time.Second)
	logger.Info("All services are ready")
	return nil
}

func (t *E2ETest) initializeGitea() error {
	logger.Info("Initializing Gitea...")

	// First check if already initialized
	resp, err := t.httpClient.Get(t.giteaURL + "/api/v1/version")
	if err == nil && resp.StatusCode == 200 {
		resp.Body.Close()
		logger.Info("Gitea is already initialized")
		return t.createUserAndToken()
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Initialize Gitea through install endpoint
	installData := map[string]interface{}{
		"db_type":           "mysql",
		"db_host":           "gitea-db-e2e:3306",
		"db_user":           "gitea",
		"db_passwd":         "gitea123",
		"db_name":           "gitea",
		"app_name":          "Gitea E2E Test",
		"admin_name":        t.username,
		"admin_passwd":      "e2epassword",
		"admin_confirm_passwd": "e2epassword",
		"admin_email":       "e2e@example.com",
	}

	jsonData, _ := json.Marshal(installData)
	req, err := http.NewRequest("POST", t.giteaURL+"/", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = t.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Wait a bit for initialization to complete
	time.Sleep(10 * time.Second)

	return t.createUserAndToken()
}

func (t *E2ETest) createUserAndToken() error {
	// Create access token for API calls
	tokenData := map[string]string{
		"name": "e2e-test-token",
	}

	jsonData, _ := json.Marshal(tokenData)

	// Try to create token using basic auth first
	req, err := http.NewRequest("POST", t.giteaURL+"/api/v1/users/"+t.username+"/tokens", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(t.username, "e2epassword")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		var tokenResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			return err
		}
		t.accessToken = tokenResp["sha1"].(string)
		logger.Info("Access token created successfully")
		return nil
	}

	// If token creation failed, assume user already exists and try to use a predefined token approach
	// For E2E testing, we'll use a simpler approach - just use basic auth
	logger.Info("Using basic authentication for API calls")
	return nil
}

func (t *E2ETest) createTestData() error {
	logger.Info("Creating test data...")

	// Create repository
	if err := t.createRepository(); err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	// Create issue
	if err := t.createIssue(); err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	return nil
}

func (t *E2ETest) createRepository() error {
	repo := GiteaRepo{
		Name:        t.repoName,
		Description: "E2E test repository",
		Private:     false,
	}

	jsonData, _ := json.Marshal(repo)
	req, err := http.NewRequest("POST", t.giteaURL+"/api/v1/user/repos", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(t.username, "e2epassword")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 409 { // 409 = already exists
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create repository: %d - %s", resp.StatusCode, string(body))
	}

	logger.Info("Repository created successfully")
	return nil
}

func (t *E2ETest) createIssue() error {
	issue := GiteaIssue{
		Title: "E2E Test Issue",
		Body:  "This is a test issue created for end-to-end testing of backup and restore functionality.",
	}

	jsonData, _ := json.Marshal(issue)
	req, err := http.NewRequest("POST", t.giteaURL+"/api/v1/repos/"+t.username+"/"+t.repoName+"/issues", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(t.username, "e2epassword")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create issue: %d - %s", resp.StatusCode, string(body))
	}

	logger.Info("Issue created successfully")
	return nil
}

func (t *E2ETest) performBackup() error {
	logger.Info("Performing backup...")

	// Execute backup command in the backup container
	cmd := exec.Command("docker", "exec", t.containerName, "gitea-backup")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("backup command failed: %w\nOutput: %s", err, string(output))
	}

	logger.Info("Backup completed successfully")
	logger.Debugf("Backup output: %s", string(output))
	return nil
}

func (t *E2ETest) simulateDataLoss() error {
	logger.Info("Simulating data loss...")

	// Stop Gitea service
	cmd := exec.Command("docker", "stop", t.giteaContainerName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop Gitea: %w", err)
	}

	// Clear Gitea data volume
	cmd = exec.Command("docker", "volume", "rm", "-f", t.dataVolumeName)
	if err := cmd.Run(); err != nil {
		logger.Debugf("Warning: failed to remove gitea-data volume: %v", err)
	}

	// Restart Gitea service
	cmd = exec.Command("docker", "start", t.giteaContainerName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart Gitea: %w", err)
	}

	// Wait for Gitea to be ready again
	time.Sleep(15 * time.Second)

	logger.Info("Data loss simulation completed")
	return nil
}

func (t *E2ETest) performRestore() error {
	logger.Info("Performing restore...")

	// Get the latest backup file name
	cmd := exec.Command("docker", "exec", t.containerName, "sh", "-c", "ls -t /tmp/backup*.zip 2>/dev/null | head -1 || echo ''")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to find backup file: %w", err)
	}

	backupFile := strings.TrimSpace(string(output))
	if backupFile == "" {
		// Try to find backup in storage or use environment variable approach
		logger.Info("No local backup file found, attempting restore from storage...")
	}

	// Execute restore command in the backup container
	cmd = exec.Command("docker", "exec", t.containerName, "gitea-restore")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restore command failed: %w\nOutput: %s", err, string(output))
	}

	logger.Info("Restore completed successfully")
	logger.Debugf("Restore output: %s", string(output))

	// Wait for Gitea to be fully ready after restore
	time.Sleep(10 * time.Second)
	return nil
}

func (t *E2ETest) verifyRestoration() error {
	logger.Info("Verifying restoration...")

	// Reinitialize if needed
	if err := t.initializeGitea(); err != nil {
		return fmt.Errorf("failed to reinitialize after restore: %w", err)
	}

	// Verify repository exists
	req, err := http.NewRequest("GET", t.giteaURL+"/api/v1/repos/"+t.username+"/"+t.repoName, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(t.username, "e2epassword")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("repository not found after restore: status %d", resp.StatusCode)
	}

	// Verify issue exists
	req, err = http.NewRequest("GET", t.giteaURL+"/api/v1/repos/"+t.username+"/"+t.repoName+"/issues", nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(t.username, "e2epassword")

	resp, err = t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get issues after restore: status %d", resp.StatusCode)
	}

	var issues []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return err
	}

	if len(issues) == 0 {
		return fmt.Errorf("no issues found after restore")
	}

	// Check if our test issue exists
	found := false
	for _, issue := range issues {
		if title, ok := issue["title"].(string); ok && title == "E2E Test Issue" {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("test issue not found after restore")
	}

	logger.Info("Restoration verification completed successfully")
	logger.Info("✅ Repository and issue were successfully restored")
	return nil
}