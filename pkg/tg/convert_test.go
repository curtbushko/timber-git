package tg

import (
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
)

func TestConvert(t *testing.T) {
	cases := []struct {
		name   string
		actual string
		want   string
	}{
		{
			name:   "Standard git repo",
			actual: "",
			want:   "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			repo, err := createInMemoryRepo()
			assert.NoError(t, err)

			repo, err := loadRepo(tmpDir)
			assert.NoError(t, err)

			status, err := isWorktree(repo)
			assert.NoError(t, err)

			got, err := convertRepo(repo)
			assert.NoError(t, err)

		})
	}
}

func createInMemoryRepo() *git.Repository {
	// Filesystem abstraction based on memory
	fs := memfs.New()
	// Git objects storer based on memory
	storer := memory.NewStorage()

	// Clones the repository into the worktree (fs) and stores all the .git
	// content into the storer
	repo, err := git.Clone(storer, fs, &git.CloneOptions{
		URL: "https://github.com/git-fixtures/basic.git",
	})
	if err != nil {
		return nil, ErrCreateNewRepo
	}

	return repo, nil
}
