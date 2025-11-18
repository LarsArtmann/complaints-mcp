package repo

import (
	"context"

	"github.com/larsartmann/complaints-mcp/internal/domain"
)

// Repository defines the interface for complaint storage
type Repository interface {
	Save(ctx context.Context, complaint *domain.Complaint) error
	FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error)
	FindAll(ctx context.Context, limit, offset int) ([]*domain.Complaint, error)
	FindBySeverity(ctx context.Context, severity domain.Severity, limit int) ([]*domain.Complaint, error)
	FindByProject(ctx context.Context, projectName string, limit int) ([]*domain.Complaint, error)
	FindUnresolved(ctx context.Context, limit int) ([]*domain.Complaint, error)
	Update(ctx context.Context, complaint *domain.Complaint) error
	Search(ctx context.Context, query string, limit int) ([]*domain.Complaint, error)
	GetCacheStats() CacheStats           // Optional - only CachedRepository implements
	WarmCache(ctx context.Context) error // Optional - warm cache with context support
	
	// GetFilePath returns the actual file path where the complaint is stored
	GetFilePath(ctx context.Context, id domain.ComplaintID) (string, error)
	// GetDocsPath returns the documentation path (if applicable) for the complaint
	GetDocsPath(ctx context.Context, id domain.ComplaintID) (string, error)
}