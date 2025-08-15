package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

// E2ETest manages the end-to-end testing process
type E2ETest struct {
	giteaURL           string
	username           string
	password           string
	repoName           string
	containerName      string
	dataVolumeName     string
	giteaContainerName string
	giteaClient        *gitea.Client
	dbVolumeName     string
	dbContainerName string
	giteaBackupfilelog string
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

	dbVolumeName := os.Getenv("DB_VOLUME_NAME")
	if dbVolumeName == "" {
		dbVolumeName = "gitea-backup-restore-process_gitea-data"
	}

	dbContainerName := os.Getenv("DB_CONTAINER_NAME")
	if dbContainerName == "" {
		dbContainerName = "gitea-e2e"
	}

	giteaBackupfilelog := os.Getenv("GITEA_BACKUP_FILE_LOG")
	if giteaBackupfilelog == "" {
		giteaBackupfilelog = "/data/gitea/backupFileLog.txt"
	}

	test := &E2ETest{
		giteaURL:           giteaURL,
		username:           "e2euser",
		password:           "e2epassword",
		repoName:           "e2e-test-repo",
		containerName:      containerName,
		dataVolumeName:     dataVolumeName,
		giteaContainerName: giteaContainerName,
		dbContainerName:	dbContainerName,
		dbVolumeName:		dbVolumeName,
		giteaBackupfilelog:	giteaBackupfilelog,
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

	backfile, err:= t.getRestoreFilename()
	if err != nil {
		return fmt.Errorf("Error to fetch backup file name: %w", err)
	}

	if err := t.performBackup(); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// Step 5: Simulate failure by clearing data
	if err := t.simulateDataLoss(); err != nil {
		return fmt.Errorf("failed to simulate data loss: %w", err)
	}

	if err := t.waitForServices(); err != nil {
		return fmt.Errorf("failed to wait for services: %w", err)
	}

	// Step 6: Perform restore
	if err := t.performRestore(backfile); err != nil {
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
		// Try to create a basic Gitea client to test connectivity
		client, err := gitea.NewClient(t.giteaURL)
		if err == nil {
			// Test basic connectivity
			_, _, err = client.ServerVersion()
			if err == nil {
				logger.Info("Gitea is ready")
				break
			}
		}

		if i == maxRetries-1 {
			return fmt.Errorf("Gitea not ready after %d attempts", maxRetries)
		}

		logger.Debugf("Waiting for Gitea... attempt %d/%d", i+1, maxRetries)
		time.Sleep(10 * time.Second)
	}

	// Wait for other services
	time.Sleep(5 * time.Second)
	logger.Info("All services are ready")
	return nil
}

func (t *E2ETest) initializeGitea() error {
	logger.Info("Initializing Gitea...")

	// Create Gitea client with basic authentication
	client, err := gitea.NewClient(t.giteaURL, gitea.SetBasicAuth(t.username, t.password))
	if err != nil {
		return fmt.Errorf("failed to create Gitea client: %w", err)
	}

	// Test if we can connect and the user exists
	_, _, err = client.GetMyUserInfo()
	if err == nil {
		logger.Info("Gitea is already initialized and user is available")
		t.giteaClient = client
		return nil
	}

	// If connection failed, try without authentication to check if Gitea is initialized
	basicClient, err := gitea.NewClient(t.giteaURL)
	if err != nil {
		return fmt.Errorf("failed to create basic Gitea client: %w", err)
	}

	// Check if Gitea is already set up by getting version
	_, _, err = basicClient.ServerVersion()
	if err != nil {
		logger.Info("Gitea appears to need initialization, waiting for auto-setup...")
		time.Sleep(30 * time.Second) // Wait for auto-initialization
	}

	// Try to authenticate with the expected user
	client, err = gitea.NewClient(t.giteaURL, gitea.SetBasicAuth(t.username, t.password))
	if err != nil {
		return fmt.Errorf("failed to create authenticated Gitea client: %w", err)
	}

	// Test authentication
	_, _, err = client.GetMyUserInfo()
	if err != nil {
		logger.Info("User authentication failed, assuming user will be created during Gitea setup")
		// For E2E testing, we'll rely on external setup or Docker initialization
		time.Sleep(10 * time.Second)
		
		// Try once more
		_, _, err = client.GetMyUserInfo()
		if err != nil {
			return fmt.Errorf("failed to authenticate with Gitea after initialization: %w", err)
		}
	}

	t.giteaClient = client
	logger.Info("Gitea client initialized successfully")
	return nil
}

func (t *E2ETest) createTestData() error {
	logger.Info("Creating test data using Gitea SDK...")

	if t.giteaClient == nil {
		return fmt.Errorf("Gitea client not initialized")
	}

	// Create repository
	repoOptions := gitea.CreateRepoOption{
		Name:          t.repoName,
		Private:       false,
		AutoInit:      true,
		DefaultBranch: "main",
	}

	repo, _, err := t.giteaClient.CreateRepo(repoOptions)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	logger.Infof("Repository created: %s", repo.FullName)

	// Create issue
	issueOptions := gitea.CreateIssueOption{
		Title: "E2E Test Issue",
		Body:  "This is a test issue created for end-to-end testing of backup and restore functionality.",
	}

	issue, _, err := t.giteaClient.CreateIssue(t.username, t.repoName, issueOptions)
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	logger.Infof("Issue created: #%d - %s", issue.Index, issue.Title)
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

	// Stop DB service
	cmd = exec.Command("docker", "stop", t.dbContainerName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop db: %w", err)
	}

	// Clear Gitea data volume (wipe instead of remove)
	cmd = exec.Command(
		"docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/data", t.dataVolumeName),
		"alpine", "sh", "-c", "rm -rf /data/* /data/.[!.]* /data/..?*",
	)
	if err := cmd.Run(); err != nil {
		logger.Debugf("Warning: failed to wipe gitea-data volume: %v", err)
	}

	// Clear DB data volume (wipe instead of remove)
	cmd = exec.Command(
		"docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/data", t.dbVolumeName),
		"alpine", "sh", "-c", "rm -rf /data/* /data/.[!.]* /data/..?*",
	)
	if err := cmd.Run(); err != nil {
		logger.Debugf("Warning: failed to wipe DB volume: %v", err)
	}

	// Restart DB service
	cmd = exec.Command("docker", "start", t.dbContainerName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart DB: %w", err)
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

func (t *E2ETest) getRestoreFilename() (string, error) {
	logger.Info("Performing restore...")

	// Step 1: Get the last line from the backupFileLog file
	cmd := exec.Command(
		"docker", "exec", t.giteaContainerName,
		"sh", "-c", "tail -n 1 " + t.giteaBackupfilelog,
	)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get last backup file: %w", err)
	}

	backupFile := strings.TrimSpace(string(output))
	if backupFile == "" {
		return "", fmt.Errorf("no backup file found in log")
	}

	// Step 2: Remove that last line from the backupFileLog file to allow restore
	cmd = exec.Command(
		"docker", "exec", t.giteaContainerName,
		"sh", "-c", "sed -i '$d' " + t.giteaBackupfilelog,
	)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to remove backup file from log: %w", err)
	}

	fmt.Println("Last backup file:", backupFile)

	return backupFile, nil
}

func (t *E2ETest) performRestore(backupFile string) error {
	logger.Info("Performing restore...")

	// Execute restore command in the backup container
	cmd := exec.Command(
		"docker", "exec", t.containerName,
		"sh", "-c", "BACKUP_FILENAME=" + backupFile + " gitea-restore",
	)
	output, err := cmd.CombinedOutput()
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

	if t.giteaClient == nil {
		return fmt.Errorf("Gitea client not available")
	}

	// Verify repository exists
	repo, _, err := t.giteaClient.GetRepo(t.username, t.repoName)
	if err != nil {
		return fmt.Errorf("repository not found after restore: %w", err)
	}

	logger.Infof("Repository verified: %s", repo.FullName)

	// Verify issue exists
	issues, _, err := t.giteaClient.ListRepoIssues(t.username, t.repoName, gitea.ListIssueOption{})
	if err != nil {
		return fmt.Errorf("failed to get issues after restore: %w", err)
	}

	if len(issues) == 0 {
		return fmt.Errorf("no issues found after restore")
	}

	// Check if our test issue exists
	found := false
	for _, issue := range issues {
		if issue.Title == "E2E Test Issue" {
			found = true
			logger.Infof("Test issue verified: #%d - %s", issue.Index, issue.Title)
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