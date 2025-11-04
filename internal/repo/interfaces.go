package repo

// RepositoryType represents the type of repository implementation
type RepositoryType string

const (
	RepositoryTypeFile   RepositoryType = "file"
	RepositoryTypeMemory RepositoryType = "memory"
	RepositoryTypeCached RepositoryType = "cached"
)
