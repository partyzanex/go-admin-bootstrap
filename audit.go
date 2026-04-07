package goadmin

import (
	"context"
	"time"
)

// AuditAction describes the type of event recorded in the audit log.
type AuditAction string

const (
	AuditLogin      AuditAction = "login"
	AuditLogout     AuditAction = "logout"
	AuditCreateUser AuditAction = "create_user"
	AuditUpdateUser AuditAction = "update_user"
	AuditDeleteUser AuditAction = "delete_user"
)

// AuditLog is a single audit trail entry.
type AuditLog struct {
	ID         int64
	ActorID    int64  // 0 when the original actor has been deleted
	ActorLogin string // snapshot of the actor's login at the time of the event
	Action     AuditAction
	EntityID   int64          // ID of the affected entity (user ID, etc.)
	Meta       map[string]any // optional context-specific fields
	DTCreated  time.Time
}

func (a *AuditLog) GetDTCreated() string {
	if a.DTCreated.IsZero() {
		return "—"
	}

	return a.DTCreated.Local().Format("2006-01-02 15:04:05")
}

// AuditLogFilter narrows the result set returned by AuditLogRepository.Search.
type AuditLogFilter struct {
	Action AuditAction
	Limit  int
	Offset int
}

// AuditLogRepository persists and retrieves audit log entries.
// It is optional: set Config.AuditLog to nil to disable audit logging.
type AuditLogRepository interface {
	Create(ctx context.Context, entry *AuditLog) (*AuditLog, error)
	Search(ctx context.Context, filter *AuditLogFilter) ([]*AuditLog, error)
	Count(ctx context.Context, filter *AuditLogFilter) (int64, error)
}
