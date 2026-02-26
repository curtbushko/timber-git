//go:build ignore
// +build ignore

package tg

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
)

func TestGetAuthMethod(t *testing.T) {
	originalPATH := os.Getenv("PATH")
	originalGitHubToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		os.Setenv("PATH", originalPATH)
		if originalGitHubToken == "" {
			os.Unsetenv("GITHUB_TOKEN")
		} else {
			os.Setenv("GITHUB_TOKEN", originalGitHubToken)
		}
	}()

	// Set a test GitHub token
	os.Setenv("GITHUB_TOKEN", "test-token")

	// Create a temporary directory with fake git that has no URL rewriting
	tmpDir := t.TempDir()
	fakeBin := filepath.Join(tmpDir, "bin")
	os.MkdirAll(fakeBin, 0755)
	gitScript := filepath.Join(fakeBin, "git")
	os.WriteFile(gitScript, []byte("#!/bin/bash\nexit 1\n"), 0755)
	os.Setenv("PATH", fakeBin+":"+originalPATH)
	tests := []struct {
		name        string
		repoURL     string
		wantSSH     bool
		wantHTTP    bool
		wantNil     bool
		expectError bool
	}{
		{
			name:    "SSH URL with git@ format",
			repoURL: "git@github.com:user/repo.git",
			wantSSH: true,
		},
		{
			name:    "SSH URL with ssh:// scheme",
			repoURL: "ssh://git@github.com/user/repo.git",
			wantSSH: true,
		},
		{
			name:     "HTTPS GitHub URL with no rewriting",
			repoURL:  "https://github.com/user/repo.git",
			wantHTTP: true,
		},
		{
			name:     "HTTP URL",
			repoURL:  "http://github.com/user/repo.git",
			wantHTTP: true,
		},
		{
			name:        "Invalid URL",
			repoURL:     "://invalid-url",
			expectError: true,
		},
		{
			name:    "Unsupported scheme",
			repoURL: "ftp://example.com/repo.git",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := getAuthMethod(tt.repoURL)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.wantNil {
				if auth != nil {
					t.Errorf("expected nil auth method but got %T", auth)
				}
				return
			}

			if tt.wantSSH {
				if auth == nil {
					t.Errorf("expected SSH auth method but got nil")
					return
				}
				// Accept both PublicKeys and PublicKeysCallback as valid SSH auth types
				if _, ok := auth.(*ssh.PublicKeys); !ok {
					if _, ok := auth.(*ssh.PublicKeysCallback); !ok {
						t.Errorf("expected SSH auth method but got %T", auth)
					}
				}
			}

			if tt.wantHTTP {
				if auth == nil {
					return
				}
				if _, ok := auth.(*http.BasicAuth); !ok {
					t.Errorf("expected HTTP auth method but got %T", auth)
				}
			}
		})
	}
}

func TestGetSSHAuth(t *testing.T) {
	originalHome := os.Getenv("HOME")
	originalSSHPassphrase := os.Getenv("SSH_PASSPHRASE")
	originalSSHAuthSock := os.Getenv("SSH_AUTH_SOCK")
	originalPATH := os.Getenv("PATH")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("PATH", originalPATH)
		if originalSSHPassphrase == "" {
			os.Unsetenv("SSH_PASSPHRASE")
		} else {
			os.Setenv("SSH_PASSPHRASE", originalSSHPassphrase)
		}
		if originalSSHAuthSock == "" {
			os.Unsetenv("SSH_AUTH_SOCK")
		} else {
			os.Setenv("SSH_AUTH_SOCK", originalSSHAuthSock)
		}
	}()

	tests := []struct {
		name          string
		setupFunc     func(t *testing.T, tmpDir string)
		setPassphrase bool
		expectError   bool
		errorContains string
	}{
		{
			name: "SSH key without passphrase",
			setupFunc: func(t *testing.T, tmpDir string) {
				createTestSSHKey(t, filepath.Join(tmpDir, ".ssh", "id_rsa"), "")
			},
		},
		{
			name: "SSH key with passphrase",
			setupFunc: func(t *testing.T, tmpDir string) {
				createTestSSHKey(t, filepath.Join(tmpDir, ".ssh", "id_rsa"), "testpass")
			},
			setPassphrase: true,
		},
		{
			name: "ed25519 key without passphrase",
			setupFunc: func(t *testing.T, tmpDir string) {
				createTestSSHKey(t, filepath.Join(tmpDir, ".ssh", "id_ed25519"), "")
			},
		},
		{
			name: "Multiple keys, use first available",
			setupFunc: func(t *testing.T, tmpDir string) {
				createTestSSHKey(t, filepath.Join(tmpDir, ".ssh", "id_ed25519"), "")
				createTestSSHKey(t, filepath.Join(tmpDir, ".ssh", "id_rsa"), "")
			},
		},
		{
			name: "No SSH keys available",
			setupFunc: func(t *testing.T, tmpDir string) {
				// Disable SSH agent for this test
				os.Unsetenv("SSH_AUTH_SOCK")
				// Create a fake ssh-add and git that fail
				fakeBin := filepath.Join(tmpDir, "bin")
				os.MkdirAll(fakeBin, 0755)
				sshAddScript := filepath.Join(fakeBin, "ssh-add")
				os.WriteFile(sshAddScript, []byte("#!/bin/bash\nexit 1\n"), 0755)
				gitScript := filepath.Join(fakeBin, "git")
				os.WriteFile(gitScript, []byte("#!/bin/bash\nexit 1\n"), 0755)
				os.Setenv("PATH", fakeBin+":"+os.Getenv("PATH"))
				// Temporarily rename real SSH directory to hide keys
				currentUser, _ := user.Current()
				realSSHDir := filepath.Join(currentUser.HomeDir, ".ssh")
				backupSSHDir := filepath.Join(currentUser.HomeDir, ".ssh.backup.test")
				if _, err := os.Stat(realSSHDir); err == nil {
					os.Rename(realSSHDir, backupSSHDir)
					t.Cleanup(func() {
						os.Rename(backupSSHDir, realSSHDir)
					})
				}
				// Create empty SSH directory
				os.MkdirAll(realSSHDir, 0700)
				t.Cleanup(func() {
					os.RemoveAll(realSSHDir)
				})
			},
			expectError:   true,
			errorContains: "no SSH authentication method available",
		},
		{
			name: "Invalid SSH key format",
			setupFunc: func(t *testing.T, tmpDir string) {
				// Disable SSH agent for this test
				os.Unsetenv("SSH_AUTH_SOCK")
				// Create a fake ssh-add and git that fail
				fakeBin := filepath.Join(tmpDir, "bin")
				os.MkdirAll(fakeBin, 0755)
				sshAddScript := filepath.Join(fakeBin, "ssh-add")
				os.WriteFile(sshAddScript, []byte("#!/bin/bash\nexit 1\n"), 0755)
				gitScript := filepath.Join(fakeBin, "git")
				os.WriteFile(gitScript, []byte("#!/bin/bash\nexit 1\n"), 0755)
				os.Setenv("PATH", fakeBin+":"+os.Getenv("PATH"))
				// Temporarily rename real SSH directory and create invalid key
				currentUser, _ := user.Current()
				realSSHDir := filepath.Join(currentUser.HomeDir, ".ssh")
				backupSSHDir := filepath.Join(currentUser.HomeDir, ".ssh.backup.test")
				if _, err := os.Stat(realSSHDir); err == nil {
					os.Rename(realSSHDir, backupSSHDir)
					t.Cleanup(func() {
						os.Rename(backupSSHDir, realSSHDir)
					})
				}
				// Create SSH directory with invalid key
				os.MkdirAll(realSSHDir, 0700)
				os.WriteFile(filepath.Join(realSSHDir, "id_rsa"), []byte("invalid key content"), 0600)
				t.Cleanup(func() {
					os.RemoveAll(realSSHDir)
				})
			},
			expectError:   true,
			errorContains: "no SSH authentication method available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			os.Setenv("HOME", tmpDir)

			if tt.setPassphrase {
				os.Setenv("SSH_PASSPHRASE", "testpass")
			} else {
				os.Unsetenv("SSH_PASSPHRASE")
			}

			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			}

			auth, err := getSSHAuth()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q but got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if auth == nil {
				t.Errorf("expected auth method but got nil")
				return
			}

			// Accept both PublicKeys and PublicKeysCallback as valid SSH auth types
			if publicKeys, ok := auth.(*ssh.PublicKeys); ok {
				if publicKeys.User != "git" {
					t.Errorf("expected user 'git' but got %q", publicKeys.User)
				}
			} else if publicKeysCallback, ok := auth.(*ssh.PublicKeysCallback); ok {
				if publicKeysCallback.User != "git" {
					t.Errorf("expected user 'git' but got %q", publicKeysCallback.User)
				}
			} else {
				t.Errorf("expected *ssh.PublicKeys or *ssh.PublicKeysCallback but got %T", auth)
				return
			}
		})
	}
}

func TestHasSSHAgentKeys(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func() func()
		expectHasKeys bool
	}{
		{
			name: "ssh-add command succeeds",
			setupFunc: func() func() {
				oldPath := os.Getenv("PATH")
				tmpDir := t.TempDir()

				sshAddScript := filepath.Join(tmpDir, "ssh-add")
				scriptContent := "#!/bin/bash\nexit 0\n"
				os.WriteFile(sshAddScript, []byte(scriptContent), 0755)

				os.Setenv("PATH", tmpDir+":"+oldPath)

				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			expectHasKeys: true,
		},
		{
			name: "ssh-add command fails",
			setupFunc: func() func() {
				oldPath := os.Getenv("PATH")
				tmpDir := t.TempDir()

				sshAddScript := filepath.Join(tmpDir, "ssh-add")
				scriptContent := "#!/bin/bash\nexit 1\n"
				os.WriteFile(sshAddScript, []byte(scriptContent), 0755)

				os.Setenv("PATH", tmpDir+":"+oldPath)

				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			expectHasKeys: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupFunc()
			defer cleanup()

			hasKeys := hasSSHAgentKeys()
			if hasKeys != tt.expectHasKeys {
				t.Errorf("expected hasSSHAgentKeys() = %v but got %v", tt.expectHasKeys, hasKeys)
			}
		})
	}
}

func TestGetGitConfigSSHKey(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tmpDir string) func()
		expectedKey string
	}{
		{
			name: "core.sshCommand with -i flag",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := fmt.Sprintf(`#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "core.sshCommand" ]]; then
    echo "ssh -i %s/.ssh/custom_key"
    exit 0
fi
exit 1
`, tmpDir)
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)

				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			expectedKey: ".ssh/custom_key",
		},
		{
			name: "user.signingkey with file path",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := fmt.Sprintf(`#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "core.sshCommand" ]]; then
    exit 1
elif [[ "$1" == "config" && "$2" == "--global" && "$3" == "user.signingkey" ]]; then
    echo "%s/.ssh/signing_key"
    exit 0
fi
exit 1
`, tmpDir)
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)

				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			expectedKey: ".ssh/signing_key",
		},
		{
			name: "user.signingkey with SSH key format and existing id_rsa",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				// Temporarily move real SSH directory and create test setup
				currentUser, _ := user.Current()
				realSSHDir := filepath.Join(currentUser.HomeDir, ".ssh")
				backupSSHDir := filepath.Join(currentUser.HomeDir, ".ssh.backup.test")
				if _, err := os.Stat(realSSHDir); err == nil {
					os.Rename(realSSHDir, backupSSHDir)
				}
				// Create SSH directory with test key in real location
				os.MkdirAll(realSSHDir, 0700)
				os.WriteFile(filepath.Join(realSSHDir, "id_rsa"), []byte("dummy key"), 0600)

				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "core.sshCommand" ]]; then
    exit 1
elif [[ "$1" == "config" && "$2" == "--global" && "$3" == "user.signingkey" ]]; then
    echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC..."
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)

				return func() {
					os.Setenv("PATH", oldPath)
					// Restore original SSH directory
					os.RemoveAll(realSSHDir)
					if _, err := os.Stat(backupSSHDir); err == nil {
						os.Rename(backupSSHDir, realSSHDir)
					}
				}
			},
			expectedKey: ".ssh/id_rsa",
		},
		{
			name: "no git config found",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)

				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			expectedKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cleanup := tt.setupFunc(t, tmpDir)
			defer cleanup()

			key := getGitConfigSSHKey()

			if tt.expectedKey != "" {
				if key == "" {
					t.Errorf("expected key but got empty string")
				} else if !strings.Contains(key, filepath.Base(tt.expectedKey)) {
					t.Errorf("expected key path to contain %q but got %q", filepath.Base(tt.expectedKey), key)
				}
			} else if tt.expectedKey == "" && key != "" {
				t.Errorf("expected empty key but got %q", key)
			}
		})
	}
}

func createTestSSHKey(tb testing.TB, keyPath, passphrase string) {
	tb.Helper()

	keyDir := filepath.Dir(keyPath)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		tb.Fatalf("failed to create key directory: %v", err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		tb.Fatalf("failed to generate RSA key: %v", err)
	}

	privateKeyDER := x509.MarshalPKCS1PrivateKey(privateKey)

	var privateKeyPEM *pem.Block
	if passphrase != "" {
		privateKeyPEM, err = x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", privateKeyDER, []byte(passphrase), x509.PEMCipherAES256)
		if err != nil {
			tb.Fatalf("failed to encrypt private key: %v", err)
		}
	} else {
		privateKeyPEM = &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyDER,
		}
	}

	privateKeyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		tb.Fatalf("failed to create private key file: %v", err)
	}
	defer privateKeyFile.Close()

	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		tb.Fatalf("failed to write private key: %v", err)
	}
}

func TestParseCredentialOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected map[string]string
	}{
		{
			name:   "valid credential output",
			output: "username=testuser\npassword=testpass\nprotocol=https\nhost=github.com\n",
			expected: map[string]string{
				"username": "testuser",
				"password": "testpass",
				"protocol": "https",
				"host":     "github.com",
			},
		},
		{
			name:     "empty output",
			output:   "",
			expected: map[string]string{},
		},
		{
			name:   "malformed lines ignored",
			output: "username=testuser\ninvalid-line\npassword=testpass\n",
			expected: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
		},
		{
			name:   "values with equals signs",
			output: "username=test@example.com\npassword=pass=with=equals\n",
			expected: map[string]string{
				"username": "test@example.com",
				"password": "pass=with=equals",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCredentialOutput(tt.output)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d credentials but got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("expected credential %q not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("credential %q: expected %q but got %q", key, expectedValue, actualValue)
				}
			}
		})
	}
}

// benchmarkSSHKeyGeneration benchmarks the SSH key generation for performance testing
func BenchmarkSSHKeyGeneration(b *testing.B) {
	tmpDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		keyPath := filepath.Join(tmpDir, fmt.Sprintf("test_key_%d", i))
		createTestSSHKey(b, keyPath, "")
	}
}

// Test helper to verify SSH auth method behavior with actual transport operations
func TestSSHAuthMethodIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	createTestSSHKey(t, filepath.Join(tmpDir, ".ssh", "id_rsa"), "")

	auth, err := getSSHAuth()
	if err != nil {
		t.Skipf("SSH auth setup failed, skipping integration test: %v", err)
	}

	if auth == nil {
		t.Error("expected non-nil auth method")
		return
	}

	// Accept both PublicKeys and PublicKeysCallback as valid SSH auth types
	if publicKeys, ok := auth.(*ssh.PublicKeys); ok {
		if publicKeys.User != "git" {
			t.Errorf("expected user 'git' but got %q", publicKeys.User)
		}
		if publicKeys.HostKeyCallback == nil {
			t.Error("expected HostKeyCallback to be set")
		}
	} else if publicKeysCallback, ok := auth.(*ssh.PublicKeysCallback); ok {
		if publicKeysCallback.User != "git" {
			t.Errorf("expected user 'git' but got %q", publicKeysCallback.User)
		}
		if publicKeysCallback.HostKeyCallback == nil {
			t.Error("expected HostKeyCallback to be set")
		}
	} else {
		t.Errorf("expected *ssh.PublicKeys or *ssh.PublicKeysCallback but got %T", auth)
		return
	}
}

func TestApplyGitURLRewriting(t *testing.T) {
	originalPATH := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPATH)

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tmpDir string) func()
		inputURL    string
		expectedURL string
		expectError bool
	}{
		{
			name: "GitHub HTTPS to SSH rewriting",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "--get-regexp" && "$4" == "url\\..*\\.insteadof" ]]; then
    echo "url.ssh://git@github.com.insteadof https://github.com"
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			inputURL:    "https://github.com/user/repo.git",
			expectedURL: "ssh://git@github.com/user/repo.git",
		},
		{
			name: "Multiple rewrite rules - first match wins",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "--get-regexp" && "$4" == "url\\..*\\.insteadof" ]]; then
    echo "url.ssh://git@github.com.insteadof https://github.com"
    echo "url.ssh://git@gitlab.com.insteadof https://gitlab.com"
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			inputURL:    "https://github.com/user/repo.git",
			expectedURL: "ssh://git@github.com/user/repo.git",
		},
		{
			name: "GitLab HTTPS to SSH rewriting",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "--get-regexp" && "$4" == "url\\..*\\.insteadof" ]]; then
    echo "url.ssh://git@gitlab.com.insteadof https://gitlab.com"
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			inputURL:    "https://gitlab.com/user/repo.git",
			expectedURL: "ssh://git@gitlab.com/user/repo.git",
		},
		{
			name: "No rewrite rules - URL unchanged",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			inputURL:    "https://example.com/user/repo.git",
			expectedURL: "https://example.com/user/repo.git",
		},
		{
			name: "SSH URL unchanged",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "--get-regexp" && "$4" == "url\\..*\\.insteadof" ]]; then
    echo "url.ssh://git@github.com.insteadof https://github.com"
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			inputURL:    "ssh://git@github.com/user/repo.git",
			expectedURL: "ssh://git@github.com/user/repo.git",
		},
		{
			name: "Prefix match behavior",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "--get-regexp" && "$4" == "url\\..*\\.insteadof" ]]; then
    echo "url.ssh://git@github.com.insteadof https://github.com"
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			inputURL:    "https://github.company.com/user/repo.git",
			expectedURL: "ssh://git@github.company.com/user/repo.git", // Matches prefix
		},
		{
			name: "No prefix match",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "--get-regexp" && "$4" == "url\\..*\\.insteadof" ]]; then
    echo "url.ssh://git@github.com.insteadof https://github.com"
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			inputURL:    "https://gitlab.com/user/repo.git",
			expectedURL: "https://gitlab.com/user/repo.git", // No match
		},
		{
			name: "Empty git config output",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "--get-regexp" && "$4" == "url\\..*\\.insteadof" ]]; then
    echo ""
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			inputURL:    "https://github.com/user/repo.git",
			expectedURL: "https://github.com/user/repo.git",
		},
		{
			name: "Malformed git config line",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "--get-regexp" && "$4" == "url\\..*\\.insteadof" ]]; then
    echo "malformed-line-without-space"
    echo "url.ssh://git@github.com.insteadof https://github.com"
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			inputURL:    "https://github.com/user/repo.git",
			expectedURL: "ssh://git@github.com/user/repo.git", // Should skip malformed line
		},
		{
			name: "Corporate GitHub instance",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "--get-regexp" && "$4" == "url\\..*\\.insteadof" ]]; then
    echo "url.ssh://git@github.ibm.com.insteadof https://github.ibm.com"
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			inputURL:    "https://github.ibm.com/user/repo.git",
			expectedURL: "ssh://git@github.ibm.com/user/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cleanup := tt.setupFunc(t, tmpDir)
			defer cleanup()

			result, err := applyGitURLRewriting(tt.inputURL)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expectedURL {
				t.Errorf("expected URL %q but got %q", tt.expectedURL, result)
			}
		})
	}
}

func TestGetAuthMethodWithURLRewriting(t *testing.T) {
	originalHome := os.Getenv("HOME")
	originalSSHPassphrase := os.Getenv("SSH_PASSPHRASE")
	originalSSHAuthSock := os.Getenv("SSH_AUTH_SOCK")
	originalPATH := os.Getenv("PATH")
	originalGitHubToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("PATH", originalPATH)
		if originalSSHPassphrase == "" {
			os.Unsetenv("SSH_PASSPHRASE")
		} else {
			os.Setenv("SSH_PASSPHRASE", originalSSHPassphrase)
		}
		if originalSSHAuthSock == "" {
			os.Unsetenv("SSH_AUTH_SOCK")
		} else {
			os.Setenv("SSH_AUTH_SOCK", originalSSHAuthSock)
		}
		if originalGitHubToken == "" {
			os.Unsetenv("GITHUB_TOKEN")
		} else {
			os.Setenv("GITHUB_TOKEN", originalGitHubToken)
		}
	}()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tmpDir string) func()
		inputURL    string
		expectSSH   bool
		expectHTTP  bool
		expectNil   bool
		expectError bool
	}{
		{
			name: "HTTPS URL rewritten to SSH",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				// Disable SSH agent and setup fake git
				os.Unsetenv("SSH_AUTH_SOCK")

				// Temporarily move real SSH directory and create test setup
				currentUser, _ := user.Current()
				realSSHDir := filepath.Join(currentUser.HomeDir, ".ssh")
				backupSSHDir := filepath.Join(currentUser.HomeDir, ".ssh.backup.test")
				if _, err := os.Stat(realSSHDir); err == nil {
					os.Rename(realSSHDir, backupSSHDir)
				}
				// Create SSH directory with test key in real location
				os.MkdirAll(realSSHDir, 0700)
				createTestSSHKey(t, filepath.Join(realSSHDir, "id_rsa"), "")

				oldPath := os.Getenv("PATH")
				fakeBin := filepath.Join(tmpDir, "bin")
				os.MkdirAll(fakeBin, 0755)

				// Create fake ssh-add that fails
				sshAddScript := filepath.Join(fakeBin, "ssh-add")
				os.WriteFile(sshAddScript, []byte("#!/bin/bash\nexit 1\n"), 0755)

				// Create fake git with URL rewriting
				gitScript := filepath.Join(fakeBin, "git")
				gitContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "--get-regexp" && "$4" == "url\\..*\\.insteadof" ]]; then
    echo "url.ssh://git@github.com.insteadof https://github.com"
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(gitContent), 0755)
				os.Setenv("PATH", fakeBin+":"+oldPath)

				return func() {
					os.Setenv("PATH", oldPath)
					// Restore original SSH directory
					os.RemoveAll(realSSHDir)
					if _, err := os.Stat(backupSSHDir); err == nil {
						os.Rename(backupSSHDir, realSSHDir)
					}
				}
			},
			inputURL:  "https://github.com/user/repo.git",
			expectSSH: true,
		},
		{
			name: "HTTPS URL with no rewriting uses HTTP auth",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				// Set GitHub token for HTTP auth
				os.Setenv("GITHUB_TOKEN", "test-token")

				oldPath := os.Getenv("PATH")
				fakeBin := filepath.Join(tmpDir, "bin")
				os.MkdirAll(fakeBin, 0755)

				// Create fake git with no URL rewriting
				gitScript := filepath.Join(fakeBin, "git")
				gitContent := `#!/bin/bash
exit 1
`
				os.WriteFile(gitScript, []byte(gitContent), 0755)
				os.Setenv("PATH", fakeBin+":"+oldPath)

				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			inputURL:   "https://github.com/user/repo.git",
			expectHTTP: true,
		},
		{
			name: "SSH URL stays SSH",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				// Disable SSH agent
				os.Unsetenv("SSH_AUTH_SOCK")

				// Temporarily move real SSH directory and create test setup
				currentUser, _ := user.Current()
				realSSHDir := filepath.Join(currentUser.HomeDir, ".ssh")
				backupSSHDir := filepath.Join(currentUser.HomeDir, ".ssh.backup.test")
				if _, err := os.Stat(realSSHDir); err == nil {
					os.Rename(realSSHDir, backupSSHDir)
				}
				// Create SSH directory with test key in real location
				os.MkdirAll(realSSHDir, 0700)
				createTestSSHKey(t, filepath.Join(realSSHDir, "id_rsa"), "")

				oldPath := os.Getenv("PATH")
				fakeBin := filepath.Join(tmpDir, "bin")
				os.MkdirAll(fakeBin, 0755)

				// Create fake ssh-add that fails
				sshAddScript := filepath.Join(fakeBin, "ssh-add")
				os.WriteFile(sshAddScript, []byte("#!/bin/bash\nexit 1\n"), 0755)

				// Create fake git (not used for SSH URLs)
				gitScript := filepath.Join(fakeBin, "git")
				os.WriteFile(gitScript, []byte("#!/bin/bash\nexit 1\n"), 0755)
				os.Setenv("PATH", fakeBin+":"+oldPath)

				return func() {
					os.Setenv("PATH", oldPath)
					// Restore original SSH directory
					os.RemoveAll(realSSHDir)
					if _, err := os.Stat(backupSSHDir); err == nil {
						os.Rename(backupSSHDir, realSSHDir)
					}
				}
			},
			inputURL:  "ssh://git@github.com/user/repo.git",
			expectSSH: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cleanup := tt.setupFunc(t, tmpDir)
			defer cleanup()

			auth, err := getAuthMethod(tt.inputURL)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectNil {
				if auth != nil {
					t.Errorf("expected nil auth method but got %T", auth)
				}
				return
			}

			if tt.expectSSH {
				if auth == nil {
					t.Errorf("expected SSH auth method but got nil")
					return
				}
				// Accept both PublicKeys and PublicKeysCallback as valid SSH auth types
				if _, ok := auth.(*ssh.PublicKeys); !ok {
					if _, ok := auth.(*ssh.PublicKeysCallback); !ok {
						t.Errorf("expected SSH auth method but got %T", auth)
					}
				}
			}

			if tt.expectHTTP {
				if auth == nil {
					t.Errorf("expected HTTP auth method but got nil")
					return
				}
				if _, ok := auth.(*http.BasicAuth); !ok {
					t.Errorf("expected HTTP auth method but got %T", auth)
				}
			}
		})
	}
}

func TestGlobalGitConfigErrors(t *testing.T) {
	originalPATH := os.Getenv("PATH")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("PATH", originalPATH)
		os.Setenv("HOME", originalHome)
	}()

	tests := []struct {
		name          string
		setupFunc     func(t *testing.T, tmpDir string) func()
		testFunc      func() (interface{}, error)
		expectError   bool
		errorContains string
	}{
		{
			name: "getGitConfigSSHKey - git command not found",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				// Set PATH to empty directory to simulate git not found
				emptyBin := filepath.Join(tmpDir, "empty")
				os.MkdirAll(emptyBin, 0755)
				os.Setenv("PATH", emptyBin)
				return func() {}
			},
			testFunc: func() (interface{}, error) {
				key := getGitConfigSSHKey()
				if key != "" {
					return nil, fmt.Errorf("expected empty key when git not found, got: %s", key)
				}
				return key, nil
			},
		},
		{
			name: "getGitConfigSSHKey - corrupted git config output",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "core.sshCommand" ]]; then
    echo "ssh -i" # Missing key path
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			testFunc: func() (interface{}, error) {
				key := getGitConfigSSHKey()
				if key != "" {
					return nil, fmt.Errorf("expected empty key with corrupted config, got: %s", key)
				}
				return key, nil
			},
		},
		{
			name: "applyGitURLRewriting - git command fails",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "--get-regexp" ]]; then
    exit 1
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			testFunc: func() (interface{}, error) {
				result, err := applyGitURLRewriting("https://github.com/user/repo.git")
				// Should not error, just return original URL
				if err != nil {
					return nil, err
				}
				if result != "https://github.com/user/repo.git" {
					return nil, fmt.Errorf("expected original URL, got: %s", result)
				}
				return result, nil
			},
		},
		{
			name: "getAuthMethod - URL rewriting error handling",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				// Disable SSH agent and setup environment
				os.Unsetenv("SSH_AUTH_SOCK")
				os.Unsetenv("GITHUB_TOKEN")

				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "--get-regexp" ]]; then
    exit 1  # Simulate git config failure
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)

				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			testFunc: func() (interface{}, error) {
				// Should handle git config errors gracefully and continue with auth
				auth, err := getAuthMethod("https://example.com/repo.git")
				// Should not error, just return nil auth for unknown host
				return auth, err
			},
		},
		{
			name: "getSSHAuth - git config SSH key errors",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				// Disable SSH agent
				os.Unsetenv("SSH_AUTH_SOCK")

				// Set fake home directory
				os.Setenv("HOME", tmpDir)

				// Temporarily move real SSH directory if it exists
				currentUser, _ := user.Current()
				realSSHDir := filepath.Join(currentUser.HomeDir, ".ssh")
				backupSSHDir := filepath.Join(currentUser.HomeDir, ".ssh.backup.test")
				if _, err := os.Stat(realSSHDir); err == nil {
					os.Rename(realSSHDir, backupSSHDir)
				}

				oldPath := os.Getenv("PATH")
				fakeBin := filepath.Join(tmpDir, "bin")
				os.MkdirAll(fakeBin, 0755)

				// Create fake ssh-add that fails
				sshAddScript := filepath.Join(fakeBin, "ssh-add")
				os.WriteFile(sshAddScript, []byte("#!/bin/bash\nexit 1\n"), 0755)

				// Create fake git that returns invalid SSH key path
				gitScript := filepath.Join(fakeBin, "git")
				scriptContent := fmt.Sprintf(`#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "core.sshCommand" ]]; then
    echo "ssh -i %s/nonexistent_key"
    exit 0
fi
exit 1
`, tmpDir)
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", fakeBin+":"+oldPath)

				return func() {
					os.Setenv("PATH", oldPath)
					// Restore original SSH directory
					if _, err := os.Stat(backupSSHDir); err == nil {
						os.Rename(backupSSHDir, realSSHDir)
					}
				}
			},
			testFunc: func() (interface{}, error) {
				auth, err := getSSHAuth()
				return auth, err
			},
			expectError:   true,
			errorContains: "no SSH authentication method available",
		},
		{
			name: "getGitConfigSSHKey - user.signingkey with nonexistent SSH directory",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				// Temporarily move real SSH directory if it exists
				currentUser, _ := user.Current()
				realSSHDir := filepath.Join(currentUser.HomeDir, ".ssh")
				backupSSHDir := filepath.Join(currentUser.HomeDir, ".ssh.backup.test")
				if _, err := os.Stat(realSSHDir); err == nil {
					os.Rename(realSSHDir, backupSSHDir)
				}

				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "core.sshCommand" ]]; then
    exit 1
elif [[ "$1" == "config" && "$2" == "--global" && "$3" == "user.signingkey" ]]; then
    echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC..."
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)

				return func() {
					os.Setenv("PATH", oldPath)
					// Restore original SSH directory
					if _, err := os.Stat(backupSSHDir); err == nil {
						os.Rename(backupSSHDir, realSSHDir)
					}
				}
			},
			testFunc: func() (interface{}, error) {
				key := getGitConfigSSHKey()
				if key != "" {
					return nil, fmt.Errorf("expected empty key when SSH directory doesn't exist, got: %s", key)
				}
				return key, nil
			},
		},
		{
			name: "getCredentialFromHelper - git credential command failure",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "credential" && "$2" == "fill" ]]; then
    echo "credential helper error" >&2
    exit 1
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			testFunc: func() (interface{}, error) {
				testURL, _ := url.Parse("https://github.com/user/repo.git")
				auth, err := getCredentialFromHelper(testURL)
				return auth, err
			},
			expectError:   true,
			errorContains: "git credential helper failed",
		},
		{
			name: "getCredentialFromHelper - empty credentials returned",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "credential" && "$2" == "fill" ]]; then
    echo ""  # Empty output
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			testFunc: func() (interface{}, error) {
				testURL, _ := url.Parse("https://github.com/user/repo.git")
				auth, err := getCredentialFromHelper(testURL)
				return auth, err
			},
			expectError:   true,
			errorContains: "no credentials returned from git credential helper",
		},
		{
			name: "getCredentialFromHelper - malformed credentials",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "credential" && "$2" == "fill" ]]; then
    echo "username=testuser"
    echo "password="  # Empty password
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			testFunc: func() (interface{}, error) {
				testURL, _ := url.Parse("https://github.com/user/repo.git")
				auth, err := getCredentialFromHelper(testURL)
				return auth, err
			},
			expectError:   true,
			errorContains: "no credentials returned from git credential helper",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cleanup := tt.setupFunc(t, tmpDir)
			defer cleanup()

			result, err := tt.testFunc()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none, result: %v", result)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q but got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGlobalGitConfigSuccess(t *testing.T) {
	originalPATH := os.Getenv("PATH")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("PATH", originalPATH)
		os.Setenv("HOME", originalHome)
	}()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string) func()
		testFunc  func() (interface{}, error)
		validate  func(t *testing.T, result interface{}, err error)
	}{
		{
			name: "getGitConfigSSHKey - successful core.sshCommand parsing",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				customKeyPath := filepath.Join(tmpDir, "custom_ssh_key")
				// Create the custom key file
				os.WriteFile(customKeyPath, []byte("dummy key content"), 0600)

				scriptContent := fmt.Sprintf(`#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "core.sshCommand" ]]; then
    echo "ssh -i %s -o UserKnownHostsFile=/dev/null"
    exit 0
fi
exit 1
`, customKeyPath)
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			testFunc: func() (interface{}, error) {
				key := getGitConfigSSHKey()
				return key, nil
			},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				key := result.(string)
				if key == "" {
					t.Error("expected SSH key path but got empty string")
				}
				if !strings.Contains(key, "custom_ssh_key") {
					t.Errorf("expected key path to contain 'custom_ssh_key', got: %s", key)
				}
			},
		},
		{
			name: "getGitConfigSSHKey - user.signingkey with absolute path",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				signingKeyPath := filepath.Join(tmpDir, "signing_key")
				// Create the signing key file
				os.WriteFile(signingKeyPath, []byte("dummy signing key"), 0600)

				scriptContent := fmt.Sprintf(`#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "core.sshCommand" ]]; then
    exit 1
elif [[ "$1" == "config" && "$2" == "--global" && "$3" == "user.signingkey" ]]; then
    echo "%s"
    exit 0
fi
exit 1
`, signingKeyPath)
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			testFunc: func() (interface{}, error) {
				key := getGitConfigSSHKey()
				return key, nil
			},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				key := result.(string)
				if key == "" {
					t.Error("expected SSH key path but got empty string")
				}
				if !strings.Contains(key, "signing_key") {
					t.Errorf("expected key path to contain 'signing_key', got: %s", key)
				}
			},
		},
		{
			name: "applyGitURLRewriting - successful GitHub to SSH rewrite",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "--get-regexp" && "$4" == "url\\..*\\.insteadof" ]]; then
    echo "url.ssh://git@github.com/.insteadof https://github.com/"
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			testFunc: func() (interface{}, error) {
				return applyGitURLRewriting("https://github.com/user/repo.git")
			},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				rewrittenURL := result.(string)
				expected := "ssh://git@github.com/user/repo.git"
				if rewrittenURL != expected {
					t.Errorf("expected %q but got %q", expected, rewrittenURL)
				}
			},
		},
		{
			name: "getCredentialFromHelper - successful credential retrieval",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "credential" && "$2" == "fill" ]]; then
    echo "username=testuser"
    echo "password=testtoken"
    echo "protocol=https"
    echo "host=github.com"
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			testFunc: func() (interface{}, error) {
				testURL, _ := url.Parse("https://github.com/user/repo.git")
				return getCredentialFromHelper(testURL)
			},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				auth := result.(*http.BasicAuth)
				if auth.Username != "testuser" {
					t.Errorf("expected username 'testuser' but got %q", auth.Username)
				}
				if auth.Password != "testtoken" {
					t.Errorf("expected password 'testtoken' but got %q", auth.Password)
				}
			},
		},
		{
			name: "applyGitURLRewriting - multiple rules, first match wins",
			setupFunc: func(t *testing.T, tmpDir string) func() {
				oldPath := os.Getenv("PATH")
				gitScript := filepath.Join(tmpDir, "git")
				scriptContent := `#!/bin/bash
if [[ "$1" == "config" && "$2" == "--global" && "$3" == "--get-regexp" && "$4" == "url\\..*\\.insteadof" ]]; then
    echo "url.ssh://git@github.com/.insteadof https://github.com/"
    echo "url.ssh://git@gitlab.com/.insteadof https://github.com/"  # Should not match
    exit 0
fi
exit 1
`
				os.WriteFile(gitScript, []byte(scriptContent), 0755)
				os.Setenv("PATH", tmpDir+":"+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			testFunc: func() (interface{}, error) {
				return applyGitURLRewriting("https://github.com/user/repo.git")
			},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				rewrittenURL := result.(string)
				expected := "ssh://git@github.com/user/repo.git"
				if rewrittenURL != expected {
					t.Errorf("expected first rule to match, got %q instead of %q", rewrittenURL, expected)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cleanup := tt.setupFunc(t, tmpDir)
			defer cleanup()

			result, err := tt.testFunc()
			tt.validate(t, result, err)
		})
	}
}
