# Development Rules

## Code Guidelines

1. **Do not use `exec.Command` in code** - Avoid using Go's `exec.Command` for executing external commands
2. **Do not use the `git` command in tests** - Tests should not rely on executing the `git` command directly