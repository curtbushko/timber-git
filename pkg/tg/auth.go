package tg

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing/transport"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
	cryptossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const (
	urlSectionName = "url"
)

func getAuthMethod(repoURL string) (transport.AuthMethod, error) {
	// Apply git config URL rewriting first
	rewrittenURL := applyGitURLRewriting(repoURL)
	
	// Use the rewritten URL for authentication method determination
	
	// Handle SSH URLs in the format git@host:path before standard URL parsing
	if strings.HasPrefix(rewrittenURL, "git@") {
		return getSSHAuth()
	}

	parsedURL, err := url.Parse(rewrittenURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "ssh" {
		return getSSHAuth()
	}

	if parsedURL.Scheme == "http" || parsedURL.Scheme == "https" {
		return getHTTPAuth(parsedURL)
	}

	return nil, nil
}

func getSSHAuth() (transport.AuthMethod, error) {
	errors := make([]string, 0, 5) // Pre-allocate for common case

	// Try SSH agent first, but only if it has keys loaded
	if hasSSHAgentKeys() {
		auth, err := ssh.NewSSHAgentAuth("git")
		if err == nil {
			// Configure host key verification
			auth.HostKeyCallback = cryptossh.InsecureIgnoreHostKey()
			return auth, nil
		}
		errors = append(errors, fmt.Sprintf("SSH agent auth failed: %v", err))
	} else {
		errors = append(errors, "SSH agent has no keys loaded")
	}

	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("error getting current user: %w", err)
	}

	// Check git config for SSH key
	if keyPath := getGitConfigSSHKey(); keyPath != "" {
		if auth := trySSHKeyWithoutPassphrase(keyPath); auth != nil {
			return auth, nil
		}
		errors = append(errors, fmt.Sprintf("git config SSH key (%s) failed", keyPath))
		
		if passphrase := os.Getenv("SSH_PASSPHRASE"); passphrase != "" {
			if auth := trySSHKeyWithPassphrase(keyPath, passphrase); auth != nil {
				return auth, nil
			}
			errors = append(errors, fmt.Sprintf("git config SSH key (%s) with passphrase failed", keyPath))
		}
	}

	// Fall back to standard SSH key locations
	sshDir := filepath.Join(currentUser.HomeDir, ".ssh")
	keyFiles := []string{
		filepath.Join(sshDir, "id_ed25519"),
		filepath.Join(sshDir, "id_ecdsa"),
		filepath.Join(sshDir, "id_rsa"),
	}

	for _, keyFile := range keyFiles {
		if _, err := os.Stat(keyFile); err != nil {
			errors = append(errors, fmt.Sprintf("SSH key file (%s) not found", keyFile))
			continue
		}
		
		// Try without passphrase first
		if auth := trySSHKeyWithoutPassphrase(keyFile); auth != nil {
			return auth, nil
		}
		errors = append(errors, fmt.Sprintf("SSH key (%s) without passphrase failed", keyFile))
		
		// Try with passphrase if available
		if passphrase := os.Getenv("SSH_PASSPHRASE"); passphrase != "" {
			if auth := trySSHKeyWithPassphrase(keyFile, passphrase); auth != nil {
				return auth, nil
			}
			errors = append(errors, fmt.Sprintf("SSH key (%s) with passphrase failed", keyFile))
		}
	}

	return nil, fmt.Errorf("no SSH authentication method available. Tried methods: %s", strings.Join(errors, "; "))
}

// trySSHKeyWithoutPassphrase attempts to create SSH auth without passphrase
func trySSHKeyWithoutPassphrase(keyPath string) transport.AuthMethod {
	auth, err := ssh.NewPublicKeysFromFile("git", keyPath, "")
	if err != nil {
		return nil
	}
	auth.HostKeyCallback = cryptossh.InsecureIgnoreHostKey()
	return auth
}

// trySSHKeyWithPassphrase attempts to create SSH auth with passphrase
func trySSHKeyWithPassphrase(keyPath, passphrase string) transport.AuthMethod {
	auth, err := ssh.NewPublicKeysFromFile("git", keyPath, passphrase)
	if err != nil {
		return nil
	}
	auth.HostKeyCallback = cryptossh.InsecureIgnoreHostKey()
	return auth
}

func hasSSHAgentKeys() bool {
	// Check if SSH_AUTH_SOCK is set
	authSock := os.Getenv("SSH_AUTH_SOCK")
	if authSock == "" {
		return false
	}

	// Try to connect to the SSH agent with timeout
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(context.Background(), "unix", authSock)
	if err != nil {
		return false
	}
	defer func() {
		_ = conn.Close()
	}()

	// Create an SSH agent client
	agentClient := agent.NewClient(conn)
	
	// List keys in the agent with timeout
	keys, err := agentClient.List()
	if err != nil {
		return false
	}
	
	// Return true if there are any keys loaded
	return len(keys) > 0
}

// getGitConfigValue gets a git config value using go-git when possible
func getGitConfigValue(key string) string {
	// Try to open the current repository to get its config
	if repo, err := git.PlainOpen("."); err == nil {
		if cfg, err := repo.Config(); err == nil {
			// Check repository config first
			if value := getConfigValueFromConfig(cfg, key); value != "" {
				return value
			}
		}
	}

	// Try to read global git config file directly
	if value := getGlobalGitConfigValue(key); value != "" {
		return value
	}

	return ""
}

// getConfigValueFromConfig extracts a value from git config
func getConfigValueFromConfig(cfg *config.Config, key string) string {
	// Handle common config keys that go-git supports
	switch key {
	case "core.sshCommand":
		// go-git doesn't directly expose core.sshCommand, check raw config
		if cfg.Raw != nil {
			for _, section := range cfg.Raw.Sections {
				if section.Name == "core" {
					for _, option := range section.Options {
						if option.Key == "sshCommand" {
							return option.Value
						}
					}
				}
			}
		}
	case "user.signingkey":
		// Check user section for signing key
		if cfg.Raw != nil {
			for _, section := range cfg.Raw.Sections {
				if section.Name == "user" {
					for _, option := range section.Options {
						if option.Key == "signingkey" {
							return option.Value
						}
					}
				}
			}
		}
	}
	return ""
}

// getGlobalGitConfigValue reads git config from global config files
func getGlobalGitConfigValue(key string) string {
	currentUser, err := user.Current()
	if err != nil {
		return ""
	}

	// Try ~/.gitconfig first
	globalConfigPath := filepath.Join(currentUser.HomeDir, ".gitconfig")
	if value := parseGitConfigFile(globalConfigPath, key); value != "" {
		return value
	}

	// Try ~/.config/git/config
	xdgConfigPath := filepath.Join(currentUser.HomeDir, ".config", "git", "config")
	if value := parseGitConfigFile(xdgConfigPath, key); value != "" {
		return value
	}

	return ""
}

// parseGitConfigFile parses a git config file and returns the value for a given key
func parseGitConfigFile(configPath, key string) string {
	file, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	var currentSection string
	keyParts := strings.SplitN(key, ".", 2)
	if len(keyParts) != 2 {
		return ""
	}
	targetSection := keyParts[0]
	targetKey := keyParts[1]

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			// Handle subsections like [url "https://github.com"]
			if spaceIndex := strings.Index(currentSection, " "); spaceIndex != -1 {
				currentSection = currentSection[:spaceIndex]
			}
			continue
		}

		// Check for key-value pairs in the target section
		if currentSection != targetSection {
			continue
		}
		
		equalIndex := strings.Index(line, "=")
		if equalIndex == -1 {
			continue
		}
		
		configKey := strings.TrimSpace(line[:equalIndex])
		if configKey != targetKey {
			continue
		}
		
		configValue := strings.TrimSpace(line[equalIndex+1:])
		// Remove quotes if present
		if len(configValue) >= 2 && configValue[0] == '"' && configValue[len(configValue)-1] == '"' {
			configValue = configValue[1 : len(configValue)-1]
		}
		return configValue
	}

	return ""
}

func getGitConfigSSHKey() string {
	// Check for core.sshCommand in git config
	if sshCommand := getGitConfigValue("core.sshCommand"); sshCommand != "" {
		if strings.Contains(sshCommand, "-i") {
			parts := strings.Fields(sshCommand)
			for i, part := range parts {
				if part == "-i" && i+1 < len(parts) {
					return parts[i+1]
				}
			}
		}
	}

	// Check for user.signingkey (for SSH signing keys)
	signingKey := getGitConfigValue("user.signingkey")
	if signingKey == "" {
		return ""
	}
	
	if strings.HasPrefix(signingKey, "ssh-") {
		// If it's an SSH key format, look for corresponding private key
		currentUser, err := user.Current()
		if err != nil {
			return ""
		}
		
		sshDir := filepath.Join(currentUser.HomeDir, ".ssh")
		// Common SSH key file patterns
		keyFiles := []string{
			filepath.Join(sshDir, "id_rsa"),
			filepath.Join(sshDir, "id_ecdsa"),
			filepath.Join(sshDir, "id_ed25519"),
		}
		for _, keyFile := range keyFiles {
			if _, err := os.Stat(keyFile); err == nil {
				return keyFile
			}
		}
		return ""
	}
	
	if strings.Contains(signingKey, "/") {
		// If it's a file path
		return signingKey
	}

	return ""
}

func getHTTPAuth(parsedURL *url.URL) (transport.AuthMethod, error) {
	if parsedURL.User != nil {
		password, _ := parsedURL.User.Password()
		return &http.BasicAuth{
			Username: parsedURL.User.Username(),
			Password: password,
		}, nil
	}

	if username := os.Getenv("GIT_USERNAME"); username != "" {
		if password := os.Getenv("GIT_PASSWORD"); password != "" {
			return &http.BasicAuth{
				Username: username,
				Password: password,
			}, nil
		}
		if token := os.Getenv("GIT_TOKEN"); token != "" {
			return &http.BasicAuth{
				Username: username,
				Password: token,
			}, nil
		}
	}

	if strings.Contains(parsedURL.Host, "github.com") {
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			return &http.BasicAuth{
				Username: "token",
				Password: token,
			}, nil
		}
	}

	if strings.Contains(parsedURL.Host, "gitlab.com") {
		if token := os.Getenv("GITLAB_TOKEN"); token != "" {
			return &http.BasicAuth{
				Username: "oauth2",
				Password: token,
			}, nil
		}
	}

	if auth, err := getCredentialFromHelper(parsedURL); err == nil {
		return auth, nil
	}

	return nil, nil
}

func getCredentialFromHelper(parsedURL *url.URL) (transport.AuthMethod, error) {
	// Try to get credentials from git credential helpers configured in git config
	helper := getGitConfigValue("credential.helper")
	if helper == "" {
		return nil, errors.New("no git credential helper configured")
	}

	// For now, we'll focus on common credential helpers that can be accessed without exec.Command
	// This is a simplified implementation - full credential helper support would require more work
	
	// Try to read from git credential store file if store helper is configured
	if strings.Contains(helper, "store") {
		return getCredentialFromStore(parsedURL)
	}

	return nil, errors.New("credential helper not available without external command execution")
}

// getCredentialFromStore reads credentials from git credential store
func getCredentialFromStore(parsedURL *url.URL) (transport.AuthMethod, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// Default credential store location
	storePath := filepath.Join(currentUser.HomeDir, ".git-credentials")
	
	// TODO: Check if custom store path is configured in git config
	// For now, use the default location

	file, err := os.Open(storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open credential store: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse credential line: https://username:password@host/path
		credURL, err := url.Parse(line)
		if err != nil {
			continue
		}

		// Check if this credential matches our target URL
		if credURL.Host == parsedURL.Host && credURL.Scheme == parsedURL.Scheme {
			if credURL.User != nil {
				password, _ := credURL.User.Password()
				if password != "" {
					return &http.BasicAuth{
						Username: credURL.User.Username(),
						Password: password,
					}, nil
				}
			}
		}
	}

	return nil, errors.New("no matching credentials found in store")
}


// applyGitURLRewriting applies git config URL rewriting rules
func applyGitURLRewriting(originalURL string) string {
	// Get URL rewriting rules from git config
	rewriteRules := getURLRewriteRules()
	
	// Apply the first matching rule
	for _, rule := range rewriteRules {
		if strings.HasPrefix(originalURL, rule.InsteadOf) {
			rewrittenURL := rule.URL + originalURL[len(rule.InsteadOf):]
			return rewrittenURL
		}
	}
	
	// Return original URL if no rules match
	return originalURL
}

// URLRewriteRule represents a git URL rewrite rule
type URLRewriteRule struct {
	URL       string
	InsteadOf string
}

// getURLRewriteRules gets URL rewrite rules from git config
func getURLRewriteRules() []URLRewriteRule {
	var rules []URLRewriteRule
	
	// Read URL rewrite rules from git config files
	currentUser, err := user.Current()
	if err != nil {
		return rules
	}

	// Check both global config locations
	configPaths := []string{
		filepath.Join(currentUser.HomeDir, ".gitconfig"),
		filepath.Join(currentUser.HomeDir, ".config", "git", "config"),
	}

	for _, configPath := range configPaths {
		rules = append(rules, parseURLRewriteRules(configPath)...)
	}

	// Also check repository config if we're in a git repo
	if repo, err := git.PlainOpen("."); err == nil {
		if cfg, err := repo.Config(); err == nil && cfg.Raw != nil {
			rules = append(rules, parseURLRewriteRulesFromConfig(cfg)...)
		}
	}

	return rules
}

// parseURLRewriteRules parses URL rewrite rules from a git config file
func parseURLRewriteRules(configPath string) []URLRewriteRule {
	var rules []URLRewriteRule
	
	file, err := os.Open(configPath)
	if err != nil {
		return rules
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	var currentSection string
	var currentURL string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for section headers like [url "ssh://git@github.com/"]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionContent := strings.Trim(line, "[]")
			if strings.HasPrefix(sectionContent, "url ") {
				currentSection = urlSectionName
				// Extract URL from quotes
				urlPart := strings.TrimPrefix(sectionContent, "url ")
				if len(urlPart) >= 2 && urlPart[0] == '"' && urlPart[len(urlPart)-1] == '"' {
					currentURL = urlPart[1 : len(urlPart)-1]
				}
			} else {
				currentSection = sectionContent
				currentURL = ""
			}
			continue
		}

		// Check for insteadOf in url sections
		if currentSection != urlSectionName || currentURL == "" {
			continue
		}
		
		equalIndex := strings.Index(line, "=")
		if equalIndex == -1 {
			continue
		}
		
		configKey := strings.TrimSpace(line[:equalIndex])
		if configKey != "insteadOf" {
			continue
		}
		
		configValue := strings.TrimSpace(line[equalIndex+1:])
		// Remove quotes if present
		if len(configValue) >= 2 && configValue[0] == '"' && configValue[len(configValue)-1] == '"' {
			configValue = configValue[1 : len(configValue)-1]
		}
		rules = append(rules, URLRewriteRule{
			URL:       currentURL,
			InsteadOf: configValue,
		})
	}

	return rules
}

// parseURLRewriteRulesFromConfig parses URL rewrite rules from go-git config
func parseURLRewriteRulesFromConfig(cfg *config.Config) []URLRewriteRule {
	var rules []URLRewriteRule
	
	if cfg.Raw == nil {
		return rules
	}

	for _, section := range cfg.Raw.Sections {
		if section.Name != urlSectionName {
			continue
		}
		
		// The URL is in the subsection name - iterate through subsections
		for _, subsection := range section.Subsections {
			if subsection.Name == "" {
				continue
			}
			
			urlValue := subsection.Name
			var insteadOfValue string
			
			// Look for insteadOf option in this subsection
			for _, option := range subsection.Options {
				if option.Key == "insteadOf" {
					insteadOfValue = option.Value
					break
				}
			}
			
			if urlValue != "" && insteadOfValue != "" {
				rules = append(rules, URLRewriteRule{
					URL:       urlValue,
					InsteadOf: insteadOfValue,
				})
			}
		}
	}

	return rules
}

