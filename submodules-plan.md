# Submodule Support via Symlinks - Implementation Plan

## Overview

This plan outlines the implementation of submodule support in timber-git using symbolic links instead of traditional git submodules. Each submodule becomes a first-class timber-git repository with full worktree support, connected via symlinks.

## Implementation Checklist

### Phase 1: Core Infrastructure
- [ ] Implement URL parsing for SSH format (`git@github.com:org/repo.git`)
- [ ] Implement URL parsing for HTTPS format (`https://github.com/org/repo.git`)
- [ ] Extract organization, repository name, and host from URLs
- [ ] Create directory structure logic: `<base-path>/github.com/<org>/<repo>/`
- [ ] Support configurable base path from config
- [ ] Implement `.gitmodules` file reading using go-git
- [ ] Parse `.gitmodules` to extract submodule path, URL, and branch
- [ ] Support recursive submodule detection
- [ ] Handle missing or malformed `.gitmodules` gracefully
- [ ] Define Go structs for link metadata (`.timber-git-links.yaml`)
- [ ] Implement YAML serialization/deserialization for link metadata
- [ ] Implement `ReadLinks(worktreePath string) (*LinkMetadata, error)`
- [ ] Implement `WriteLinks(worktreePath string, metadata *LinkMetadata) error`
- [ ] Implement `AddLink(metadata *LinkMetadata, link Link) error`
- [ ] Implement `RemoveLink(metadata *LinkMetadata, path string) error`
- [ ] Implement `UpdateLink(metadata *LinkMetadata, path string, newBranch string) error`
- [ ] Write tests for URL parsing (SSH/HTTPS/edge cases)
- [ ] Write tests for submodule detection and parsing
- [ ] Write tests for metadata management functions

### Phase 2: Core Commands Enhancement
- [ ] Add submodule detection to clone command
- [ ] Implement recursive submodule cloning as timber-git repos
- [ ] Create initial `.timber-git-links.yaml` during clone
- [ ] Create symlinks from parent to submodule worktrees during clone
- [ ] Add `.timber-git-links.yaml` to parent repo's `.gitignore`
- [ ] Add `--no-submodules` flag to clone command
- [ ] Provide progress feedback during submodule cloning
- [ ] Read `.timber-git-links.yaml` during checkout
- [ ] Recreate symlinks when creating new worktrees
- [ ] Add flags to checkout for specifying submodule branches
- [ ] Update `.timber-git-links.yaml` in new worktrees
- [ ] Validate target submodule branches exist
- [ ] Implement relative path calculation between worktrees
- [ ] Handle multi-org path resolution correctly
- [ ] Validate symlink targets exist before creation
- [ ] Handle existing symlinks (update vs. error)
- [ ] Support both Unix symlinks and Windows junctions
- [ ] Write integration tests for clone with submodules
- [ ] Write integration tests for checkout with symlinks

### Phase 3: Link Management Commands
- [ ] Create `cmd/link.go` - main link command
- [ ] Implement `tg link add <path> <repo-url> <branch>`
- [ ] Implement `tg link remove <path>`
- [ ] Implement `tg link list`
- [ ] Implement `tg link update <path> <new-branch>`
- [ ] Implement `tg link sync`
- [ ] Add confirmation prompts for destructive operations
- [ ] Add link status validation (valid/broken)
- [ ] Write tests for each link subcommand
- [ ] Write integration tests for link workflows

### Phase 4: Testing & Documentation
- [ ] Write comprehensive URL parsing tests
- [ ] Write `.gitmodules` parsing tests
- [ ] Write symlink creation/update tests
- [ ] Write cross-platform symlink tests (Unix/Windows)
- [ ] Write multi-org scenario tests
- [ ] Write full integration test suite
- [ ] Create `docs/submodules.md` - main submodule documentation
- [ ] Create `docs/commands/link.md` - link command reference
- [ ] Update `CLAUDE.md` with submodule development guidelines
- [ ] Update `README.md` with submodule overview
- [ ] Write migration guide from git submodules
- [ ] Write troubleshooting guide
- [ ] Add workflow examples to documentation

## Directory Structure Design

### Multi-Organization Support

Repositories are cloned into an organization-based hierarchy to support submodules from different GitHub organizations:

```
~/workspace/
├── github.com/
│   ├── org1/
│   │   └── project-bar/
│   │       ├── .bare/              # Bare repository
│   │       ├── main/               # Main branch worktree
│   │       └── feature-x/          # Feature branch worktree
│   │
│   └── org2/
│       └── project-foo/
│           ├── .bare/
│           ├── main/
│           │   └── vendor/
│           │       └── submodule1/ # Symlink -> ../../../../org1/project-bar/main/
│           └── feature-branch/
│               └── vendor/
│                   └── submodule1/ # Symlink -> ../../../../org1/project-bar/feature-x/
```

### Key Design Decisions

1. **Organization-based hierarchy**: Repositories are cloned into `github.com/<org>/<repo>/` structure
2. **Relative symlinks**: Use relative paths for portability (e.g., `../../../../org1/project-bar/main/`)
3. **Per-worktree linking**: Each worktree can link to different branches of submodules
4. **Configurable base path**: Support custom base paths via configuration

## Metadata Format

### Link Tracking Metadata

A `.timber-git-links.yaml` file stored in each worktree tracks submodule symlinks:

```yaml
# .timber-git-links.yaml
# Tracks symlink mappings for this worktree

version: 1
worktree_branch: feature-branch

links:
  - path: vendor/submodule1              # Path within this worktree
    target_org: org1                      # GitHub organization
    target_repo: project-bar              # Repository name
    target_branch: feature-x              # Branch to link to
    url: git@github.com:org1/project-bar.git  # Original git URL

  - path: vendor/another-module
    target_org: org3
    target_repo: shared-lib
    target_branch: main
    url: https://github.com/org3/shared-lib.git
```

### Metadata Location

- **Per-worktree**: `.timber-git-links.yaml` in each worktree root
- Not tracked in git (added to `.gitignore`)
- Each worktree can have independent link configurations

## Implementation Plan

### Phase 1: Core Infrastructure

#### 1. URL Parsing & Repository Structure

**Goal**: Parse git URLs to extract organization and repository information, supporting multiple URL formats.

**Tasks**:
- Implement URL parser for SSH format: `git@github.com:org/repo.git`
- Implement URL parser for HTTPS format: `https://github.com/org/repo.git`
- Extract organization, repository name, and host
- Create directory structure: `<base-path>/github.com/<org>/<repo>/`
- Support configurable base path from config

**Files to create/modify**:
- `pkg/tg/url.go` - URL parsing functions
- `pkg/tg/url_test.go` - URL parsing tests

#### 2. Submodule Detection

**Goal**: Detect and parse `.gitmodules` file to identify submodule dependencies.

**Tasks**:
- Use go-git to read `.gitmodules` from repository
- Parse INI-style format to extract:
  - Submodule path (local directory)
  - Submodule URL (git repository)
  - Submodule branch (if specified)
- Support recursive submodule detection
- Handle missing or malformed `.gitmodules` gracefully

**Files to create/modify**:
- `pkg/tg/submodule.go` - Submodule detection and parsing
- `pkg/tg/submodule_test.go` - Submodule detection tests

#### 3. Metadata Management

**Goal**: Implement functions to read, write, and manage `.timber-git-links.yaml` files.

**Tasks**:
- Define Go structs for link metadata
- Implement YAML serialization/deserialization
- Create functions:
  - `ReadLinks(worktreePath string) (*LinkMetadata, error)`
  - `WriteLinks(worktreePath string, metadata *LinkMetadata) error`
  - `AddLink(metadata *LinkMetadata, link Link) error`
  - `RemoveLink(metadata *LinkMetadata, path string) error`
  - `UpdateLink(metadata *LinkMetadata, path string, newBranch string) error`

**Files to create/modify**:
- `pkg/tg/links.go` - Link metadata management
- `pkg/tg/links_test.go` - Metadata management tests

### Phase 2: Core Commands Enhancement

#### 4. Enhanced `clone` Command

**Goal**: Automatically detect and clone submodules as timber-git repositories during clone operation.

**Tasks**:
- After cloning main repository, check for `.gitmodules`
- For each submodule:
  - Parse submodule URL to extract org/repo
  - Clone submodule as separate timber-git repo to correct org path
  - Track cloned submodules
- Create initial `.timber-git-links.yaml` in default branch worktree
- Map each submodule path to its default branch
- Create symlinks from parent repo to submodule worktrees
- Add `.timber-git-links.yaml` to parent repo's `.gitignore`
- Provide progress feedback during submodule cloning

**Files to modify**:
- `pkg/tg/clone.go` - Add submodule handling
- `cmd/clone.go` - Add flags for submodule behavior (e.g., `--no-submodules`)

#### 5. Enhanced `checkout` Command

**Goal**: Automatically manage submodule symlinks when creating new worktrees.

**Tasks**:
- When checking out new branch:
  - Read `.timber-git-links.yaml` from source worktree (typically main)
  - Or use default mappings if no metadata exists
- Create new worktree for the branch
- Recreate symlinks based on metadata
- Allow user to specify different submodule branches via flags
- Update `.timber-git-links.yaml` in new worktree
- Validate that target submodule branches exist

**Files to modify**:
- `pkg/tg/checkout.go` - Add symlink management
- `cmd/checkout.go` - Add flags for submodule branch selection

#### 6. Symlink Creation Logic

**Goal**: Implement robust symlink creation with proper path resolution across organizations.

**Tasks**:
- Calculate relative paths between parent and submodule worktrees
- Handle different organization paths correctly
- Validate symlink targets exist before creation
- Handle existing symlinks (update vs. error)
- Support both Unix symlinks and Windows junctions
- Provide clear error messages for symlink failures

**Files to create/modify**:
- `pkg/tg/symlink.go` - Symlink creation and management
- `pkg/tg/symlink_test.go` - Symlink tests

### Phase 3: Link Management Commands

#### 7. New `tg link` Command

**Goal**: Provide manual control over submodule symlink management.

**Subcommands**:

##### `tg link add <path> <repo-url> <branch>`
- Parse repo URL to extract org/repo
- Verify target repository and branch exist
- Create symlink from `<path>` to target branch worktree
- Add entry to `.timber-git-links.yaml`

##### `tg link remove <path>`
- Remove symlink at specified path
- Remove entry from `.timber-git-links.yaml`
- Optionally prompt for confirmation

##### `tg link list`
- Read `.timber-git-links.yaml`
- Display all current symlinks in human-readable format
- Show: path, target org/repo, target branch, link status (valid/broken)

##### `tg link update <path> <new-branch>`
- Verify new branch exists in target repository
- Update symlink to point to new branch worktree
- Update `.timber-git-links.yaml` with new branch

##### `tg link sync`
- Read `.timber-git-links.yaml`
- Ensure all symlinks match metadata
- Fix broken or missing symlinks
- Report any inconsistencies

**Files to create**:
- `cmd/link.go` - Main link command
- `cmd/link_add.go` - Add subcommand
- `cmd/link_remove.go` - Remove subcommand
- `cmd/link_list.go` - List subcommand
- `cmd/link_update.go` - Update subcommand
- `cmd/link_sync.go` - Sync subcommand

### Phase 4: Testing & Documentation

#### 8. Comprehensive Tests

**Test Coverage**:

##### URL Parsing Tests
- SSH URL format variations
- HTTPS URL format variations
- Various GitHub/GitLab/Bitbucket URLs
- Invalid URL handling
- Edge cases (missing .git extension, etc.)

##### Submodule Detection Tests
- Valid `.gitmodules` parsing
- Missing `.gitmodules` handling
- Malformed `.gitmodules` handling
- Multiple submodules
- Nested/recursive submodules

##### Symlink Tests
- Symlink creation across different orgs
- Relative path calculation
- Broken symlink detection
- Symlink update scenarios
- Cross-platform symlink support (Unix/Windows)

##### Integration Tests
- Full clone workflow with submodules
- Checkout with submodule linking
- Link command operations
- Multi-org scenarios

**Files to create/modify**:
- `pkg/tg/url_test.go`
- `pkg/tg/submodule_test.go`
- `pkg/tg/symlink_test.go`
- `pkg/tg/links_test.go`
- Integration test suite

#### 9. Documentation

**Documentation Topics**:

##### User Guide
- Introduction to symlink-based submodules
- Basic workflow examples
- Multi-org repository management
- Common use cases and patterns

##### Command Reference
- Detailed documentation for each command
- Flag descriptions and examples
- Use case examples

##### Migration Guide
- Converting from git submodules to timber-git links
- Step-by-step migration process
- Gotchas and considerations

##### Troubleshooting
- Common issues and solutions
- Debugging broken symlinks
- Path resolution problems
- Organization structure issues

**Files to create/modify**:
- `docs/submodules.md` - Main submodule documentation
- `docs/commands/link.md` - Link command documentation
- `CLAUDE.md` - Update development guide
- `README.md` - Add submodule section

## Key Features

- ✅ Multi-organization support with proper directory hierarchy
- ✅ Per-worktree independent submodule branch selection
- ✅ Relative symlinks for portability
- ✅ Metadata tracking for reproducibility
- ✅ No git submodule complexity
- ✅ Full timber-git workflow for submodules

## Example Workflows

### Cloning a Repository with Submodules

```bash
# Clone parent repo with submodules
$ tg clone git@github.com:org2/project-foo.git

# What happens:
# 1. Clones to github.com/org2/project-foo/
# 2. Detects submodules in .gitmodules
# 3. Clones each submodule as timber-git repo (e.g., github.com/org1/project-bar/)
# 4. Creates .timber-git-links.yaml in main/ worktree
# 5. Creates symlinks: main/vendor/submodule1 -> ../../../../org1/project-bar/main/

$ cd github.com/org2/project-foo/main
$ ls -la vendor/submodule1
lrwxrwxrwx vendor/submodule1 -> ../../../../org1/project-bar/main/
```

### Creating a New Worktree with Submodules

```bash
$ cd github.com/org2/project-foo

# Create new worktree
$ tg checkout -b feature-branch

# What happens:
# 1. Creates feature-branch/ worktree
# 2. Copies .timber-git-links.yaml from main/ (or uses defaults)
# 3. Creates symlinks based on metadata

$ cd feature-branch
$ tg link list
vendor/submodule1 → org1/project-bar (main)
```

### Changing Submodule Branch for a Worktree

```bash
$ cd github.com/org2/project-foo/feature-branch

# Update submodule to use different branch
$ tg link update vendor/submodule1 feature-x

# What happens:
# 1. Verifies feature-x branch exists in org1/project-bar
# 2. Updates symlink: vendor/submodule1 -> ../../../../org1/project-bar/feature-x/
# 3. Updates .timber-git-links.yaml

$ tg link list
vendor/submodule1 → org1/project-bar (feature-x)
```

### Manually Adding a Submodule Link

```bash
$ cd github.com/org2/project-foo/main

# Add a new submodule link
$ tg link add vendor/new-lib git@github.com:org3/shared-lib.git main

# What happens:
# 1. Checks if org3/shared-lib is cloned, clones if needed
# 2. Creates symlink: vendor/new-lib -> ../../../../org3/shared-lib/main/
# 3. Adds entry to .timber-git-links.yaml

$ tg link list
vendor/submodule1 → org1/project-bar (main)
vendor/new-lib → org3/shared-lib (main)
```

### Syncing Links After Manual Changes

```bash
$ cd github.com/org2/project-foo/feature-branch

# Manually edited .timber-git-links.yaml or symlinks are broken
$ tg link sync

# What happens:
# 1. Reads .timber-git-links.yaml
# 2. Validates each symlink exists and points to correct target
# 3. Fixes any broken or missing symlinks
# 4. Reports status

Syncing links...
✓ vendor/submodule1 - OK
✗ vendor/new-lib - broken (fixing...)
✓ vendor/new-lib - fixed
All links synchronized.
```

## Advantages Over Git Submodules

1. **No detached HEAD states**: Each submodule is a full timber-git repo
2. **Independent branch work**: Work on different branches of parent and submodules simultaneously
3. **Simpler workflow**: No `git submodule update` complexity
4. **Full git operations**: All git commands work normally in submodules
5. **Per-worktree flexibility**: Each worktree can link to different submodule branches
6. **Clear visualization**: Symlinks make dependencies explicit and visible
7. **Multi-org support**: Natural support for submodules from different organizations

## Considerations and Limitations

1. **Storage**: Each submodule is fully cloned, using more disk space than traditional submodules
2. **Symlink support**: Requires filesystem symlink support (limitations on some Windows setups)
3. **Manual coordination**: Users must manually ensure compatible versions across repos
4. **Learning curve**: New concept for users familiar with git submodules
5. **Build tools**: Some build tools may need configuration to follow symlinks

## Future Enhancements

1. **Version pinning**: Optional feature to track specific commits (like git submodules)
2. **Automatic branch matching**: Auto-link to same branch name in submodules if available
3. **Bulk operations**: Commands to update all submodule links at once
4. **Status overview**: Show status of all linked repositories
5. **GitLab/Bitbucket support**: Extend beyond GitHub to other platforms
6. **Sparse checkout**: Support sparse checkout for large submodules
