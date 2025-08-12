package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ReadGiteaConfig reads and parses the Gitea app.ini configuration file
func ReadGiteaConfig(iniPath string) (*GiteaConfig, error) {
	file, err := os.Open(iniPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := &GiteaConfig{}
	scanner := bufio.NewScanner(file)
	
	var currentSection string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Check for section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}
		
		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes if present
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		
		switch currentSection {
		case "database":
			switch strings.ToUpper(key) {
			case "DB_TYPE":
				config.Database.DBType = strings.ToLower(value)
			case "HOST":
				config.Database.Host = value
			case "NAME":
				config.Database.Name = value
			case "USER":
				config.Database.User = value
			case "PASSWD":
				config.Database.Passwd = value
			case "PATH":
				config.Database.Path = value
			}
		case "repository":
			switch strings.ToUpper(key) {
			case "ROOT":
				config.Repository.Root = value
			}
		case "picture":
			switch strings.ToUpper(key) {
			case "AVATAR_UPLOAD_PATH":
				config.Picture.AvatarUploadPath = value
			case "REPOSITORY_AVATAR_UPLOAD_PATH":
				config.Picture.RepositoryAvatarUploadPath = value
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	
	return config, nil
}