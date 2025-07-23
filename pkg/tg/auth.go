package tg

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v6/plumbing/transport"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
	cryptossh "golang.org/x/crypto/ssh"
)

func getAuthMethod(repoURL string) (transport.AuthMethod, error) {
	// Apply git config URL rewriting first
	rewrittenURL, err := applyGitURLRewriting(repoURL)
	if err != nil {
		return nil, fmt.Errorf("error applying git URL rewriting: %v", err)
	}
	
	// Use the rewritten URL for authentication method determination
	
	// Handle SSH URLs in the format git@host:path before standard URL parsing
	if strings.HasPrefix(rewrittenURL, "git@") {
		return getSSHAuth()
	}

	parsedURL, err := url.Parse(rewrittenURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
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
	var errors []string

	// Try SSH agent first, but only if it has keys loaded
	if hasSSHAgentKeys() {
		if auth, err := ssh.NewSSHAgentAuth("git"); err == nil {
			// Configure host key verification
			auth.HostKeyCallback = cryptossh.InsecureIgnoreHostKey()
			return auth, nil
		} else {
			errors = append(errors, fmt.Sprintf("SSH agent auth failed: %v", err))
		}
	} else {
		errors = append(errors, "SSH agent has no keys loaded")
	}

	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("error getting current user: %v", err)
	}

	// Check git config for SSH key
	if keyPath := getGitConfigSSHKey(); keyPath != "" {
		if auth, err := ssh.NewPublicKeysFromFile("git", keyPath, ""); err == nil {
			auth.HostKeyCallback = cryptossh.InsecureIgnoreHostKey()
			return auth, nil
		} else {
			errors = append(errors, fmt.Sprintf("git config SSH key (%s) failed: %v", keyPath, err))
		}
		if passphrase := os.Getenv("SSH_PASSPHRASE"); passphrase != "" {
			if auth, err := ssh.NewPublicKeysFromFile("git", keyPath, passphrase); err == nil {
				auth.HostKeyCallback = cryptossh.InsecureIgnoreHostKey()
				return auth, nil
			} else {
				errors = append(errors, fmt.Sprintf("git config SSH key (%s) with passphrase failed: %v", keyPath, err))
			}
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
		if _, err := os.Stat(keyFile); err == nil {
			// Try without passphrase first
			if auth, err := ssh.NewPublicKeysFromFile("git", keyFile, ""); err == nil {
				auth.HostKeyCallback = cryptossh.InsecureIgnoreHostKey()
				return auth, nil
			} else {
				errors = append(errors, fmt.Sprintf("SSH key (%s) without passphrase failed: %v", keyFile, err))
			}
			
			// Try with passphrase if available
			if passphrase := os.Getenv("SSH_PASSPHRASE"); passphrase != "" {
				if auth, err := ssh.NewPublicKeysFromFile("git", keyFile, passphrase); err == nil {
					auth.HostKeyCallback = cryptossh.InsecureIgnoreHostKey()
					return auth, nil
				} else {
					errors = append(errors, fmt.Sprintf("SSH key (%s) with passphrase failed: %v", keyFile, err))
				}
			}
		} else {
			errors = append(errors, fmt.Sprintf("SSH key file (%s) not found", keyFile))
		}
	}

	return nil, fmt.Errorf("no SSH authentication method available. Tried methods: %s", strings.Join(errors, "; "))
}

func hasSSHAgentKeys() bool {
	cmd := exec.Command("ssh-add", "-l")
	err := cmd.Run()
	return err == nil
}

func getGitConfigSSHKey() string {
	// Check for core.sshCommand in git config
	cmd := exec.Command("git", "config", "--global", "core.sshCommand")
	output, err := cmd.Output()
	if err == nil {
		sshCommand := strings.TrimSpace(string(output))
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
	cmd = exec.Command("git", "config", "--global", "user.signingkey")
	output, err = cmd.Output()
	if err == nil {
		signingKey := strings.TrimSpace(string(output))
		if strings.HasPrefix(signingKey, "ssh-") {
			// If it's an SSH key format, look for corresponding private key
			currentUser, err := user.Current()
			if err == nil {
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
			}
		} else if strings.Contains(signingKey, "/") {
			// If it's a file path
			return signingKey
		}
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
	cmd := exec.Command("git", "credential", "fill")

	input := fmt.Sprintf("protocol=%s\nhost=%s\npath=%s\n\n",
		parsedURL.Scheme, parsedURL.Host, parsedURL.Path)

	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git credential helper failed: %v", err)
	}

	credentials := parseCredentialOutput(string(output))
	if credentials["username"] != "" && credentials["password"] != "" {
		return &http.BasicAuth{
			Username: credentials["username"],
			Password: credentials["password"],
		}, nil
	}

	return nil, fmt.Errorf("no credentials returned from git credential helper")
}

func parseCredentialOutput(output string) map[string]string {
	credentials := make(map[string]string)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				credentials[parts[0]] = parts[1]
			}
		}
	}

	return credentials
}

// applyGitURLRewriting applies git config URL rewriting rules
func applyGitURLRewriting(originalURL string) (string, error) {
	// Get all URL rewriting rules from git config
	cmd := exec.Command("git", "config", "--global", "--get-regexp", "url\\..*\\.insteadof")
	output, err := cmd.Output()
	if err != nil {
		// No rewriting rules found, return original URL
		return originalURL, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		// Parse the git config line: "url.<replacement>.insteadof <original>"
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		
		configKey := parts[0]
		originalPattern := parts[1]
		
		// Extract the replacement URL from the config key
		// Format: "url.<replacement>.insteadof"
		if !strings.HasPrefix(configKey, "url.") || !strings.HasSuffix(configKey, ".insteadof") {
			continue
		}
		
		replacement := configKey[4 : len(configKey)-10] // Remove "url." prefix and ".insteadof" suffix
		
		// Check if the original URL matches the pattern
		if strings.HasPrefix(originalURL, originalPattern) {
			// Replace the matching prefix
			rewrittenURL := replacement + originalURL[len(originalPattern):]
			return rewrittenURL, nil
		}
	}
	
	// No matching rewrite rule found
	return originalURL, nil
}
