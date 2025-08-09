# Timber-Git Development Guide

## Project Overview

**Timber-Git** is an opinionated CLI tool for Git worktree management that simplifies working with multiple branches simultaneously. The tool creates bare repositories with organized worktree structures, emphasizing frequent rebasing to keep worktrees in sync with the main branch.

### Key Features
- Bare repository cloning with automatic worktree setup
- Branch-based worktree organization in `<repo-name>/<branch-name>/` structure
- Multiple authentication methods (SSH keys, SSH agent, HTTP tokens)
- FZF integration for branch selection
- Convert existing repositories to timber-git format

## Development Rules

### Core Guidelines

1. **Do not use `exec.Command` in code** - Avoid using Go's `exec.Command` for executing external commands
2. **Do not use the `git` command in tests** - Tests should not rely on executing the `git` command directly  
3. **Do use 'go-git' for ALL git functionality** - Go git will be able to do everything that the git cli can do but in code. Use this exclusively.
4. **Do not needlessly rename functions** - Prefer fixing a functionality over renaming a function.
5. **Do not use 'Main' in function names - The default branch for a git repository may not be main.
6. **Do use a './temp' directory if needing to test the timber-git binary directly** 

### Code Standards

1. **Error Handling**: Always return descriptive errors with context using `fmt.Errorf` and error wrapping
2. **Authentication**: Support multiple auth methods with graceful fallbacks (SSH keys → SSH agent → HTTP)
3. **Path Handling**: Use `filepath.Join()` for cross-platform path construction
4. **Logging**: Provide progress feedback during long operations
5. **Testing**: Write comprehensive tests for all public functions
6. **Configuration**: Use Viper for configuration management

## Architecture

### Directory Structure
```
├── cmd/                    # Cobra CLI commands (add, checkout, clone, etc.)
├── pkg/tg/                 # Core timber-git functionality  
├── main.go                 # Application entry point
├── Makefile                # Build automation
└── go.mod                  # Module dependencies
```

### Key Dependencies
- `github.com/go-git/go-git/v6` - All Git operations
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `github.com/junegunn/fzf` - Interactive branch selection
- `github.com/stretchr/testify` - Testing framework

## Development Workflow

### Quality Assurance Commands
```bash
make test          # Run all tests: go test ./...
make lint          # Run golangci-lint
make vet           # Run go vet ./...
make tidy          # Run go mod tidy
```

### Build Commands
```bash
make build         # Build binary for current platform
make install       # Install using go install
make run           # Run the built binary
make clean         # Clean bin/ directory
```

### Testing Guidelines

1. **Use testify/assert**: Prefer `assert` and `require` for test assertions
2. **In-memory testing**: Use go-git's memory storage for isolated tests when possible
3. **Temp directories**: Always clean up temporary test directories with `defer`
4. **Network isolation**: Tests should handle network failures gracefully
5. **No git command**: Never use `exec.Command` to run git in tests

### Example Test Structure
```go
func TestFunction(t *testing.T) {
    // Setup
    tempDir := t.TempDir() // Automatically cleaned up
    
    // Test logic using go-git
    repo, err := git.Init(memory.NewStorage(), nil)
    require.NoError(t, err)
    
    // Assertions
    assert.NotNil(t, repo)
}
```

## Git Operations Patterns

### Repository Cloning
```go
// Use PlainClone with authentication
repo, err := git.PlainClone(path, true, &git.CloneOptions{
    URL:  repoURL,
    Auth: auth, // SSH or HTTP auth
})
```

### Worktree Management  
```go
// Create worktree
worktree, err := repo.CreateWorktree(&git.WorktreeOptions{
    Path:   worktreePath,
    Branch: plumbing.NewBranchReferenceName(branch),
})
```

### Authentication Handling
```go
// Try multiple auth methods in order:
// 1. SSH keys from ~/.ssh/
// 2. SSH agent
// 3. HTTP basic auth
// 4. Token-based auth
```

## Configuration

### Config File Location
- Primary: `~/.timber-git.yaml`
- Future: `~/.config/tg/config.yaml`

### Supported Settings
```yaml
auth:
  ssh_key_paths: ["~/.ssh/id_rsa", "~/.ssh/id_ed25519"]
  use_ssh_agent: true
  http_username: ""
  http_token: ""

worktree:
  base_path: "."
  auto_rebase: true
```

## File Organization

### Worktree Structure
```
project-name/
├── .bare/              # Bare repository
└── branch-name/        # Worktree for branch
    ├── .git           # Points to .bare
    └── project files...
```

### Important Files
- `pkg/tg/auth.go` - Authentication methods
- `pkg/tg/clone.go` - Repository cloning logic  
- `pkg/tg/add.go` - Worktree creation
- `cmd/*.go` - CLI command implementations

## Common Patterns

### Error Context
```go
if err != nil {
    return fmt.Errorf("failed to clone repository %s: %w", repoURL, err)
}
```

### Progress Reporting
```go
fmt.Printf("Cloning repository %s...\n", repoURL)
// Long operation
fmt.Printf("Repository cloned successfully to %s\n", targetPath)
```

### Path Safety
```go
targetPath := filepath.Join(baseDir, repoName)
targetPath, err := filepath.Abs(targetPath)
```

## Security Guidelines

1. **Credential Handling**: Never log or expose authentication credentials
2. **Path Traversal**: Validate all user-provided paths to prevent directory traversal
3. **Temporary Files**: Clean up temporary files and directories
4. **SSH Keys**: Support standard SSH key locations and types
5. **HTTPS Verification**: Always verify HTTPS certificates in production

## Performance Considerations

1. **Large Repositories**: Handle large repositories with progress reporting
2. **Network Timeouts**: Set appropriate timeouts for network operations
3. **Memory Usage**: Use streaming operations for large file transfers when possible
4. **Concurrent Operations**: Avoid concurrent modifications to the same repository

## Troubleshooting

### Common Issues
- **Authentication failures**: Check SSH keys, agent, and HTTP credentials
- **Path conflicts**: Ensure target directories don't already exist
- **Network issues**: Handle timeouts and connection failures gracefully
- **Permission issues**: Check filesystem permissions for target directories

### Debug Information
- Enable verbose logging during development
- Provide clear error messages with actionable steps
- Include relevant context (paths, URLs, auth methods tried)

## Future Enhancements

1. **Configuration Migration**: Move config to `~/.config/tg/`
2. **Convert Command**: Complete convert functionality with tests
3. **Git Helper Functions**: Add utilities for common git operations
4. **Performance Optimization**: Improve handling of large repositories
5. **Shell Completion**: Add bash/zsh completion support

# Important Instruction Reminders

Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.
