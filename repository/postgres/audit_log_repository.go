package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
)

type auditLogRepository struct {
	db *bun.DB
}

// NewAuditLogRepository returns a PostgreSQL-backed AuditLogRepository.
func NewAuditLogRepository(db *bun.DB) goadmin.AuditLogRepository {
	return &auditLogRepository{db: db}
}

// auditLogModel is the bun ORM representation of goadmin.audit_log.
type auditLogModel struct {
	bun.BaseModel `bun:"table:goadmin.audit_log,alias:al"`

	ID         int64               `bun:"id,pk,autoincrement"`
	ActorID    *int64              `bun:"actor_id"`
	ActorLogin string              `bun:"actor_login,notnull"`
	Action     goadmin.AuditAction `bun:"action,notnull"`
	EntityID   int64               `bun:"entity_id,notnull"`
	Meta       json.RawMessage     `bun:"meta,type:jsonb,notnull"`
	DTCreated  time.Time           `bun:"dt_created,notnull,default:now()"`
}

func auditLogToModel(a *goadmin.AuditLog) (*auditLogModel, error) {
	meta := a.Meta
	if meta == nil {
		meta = map[string]any{}
	}

	raw, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("marshal audit meta: %w", err)
	}

	m := &auditLogModel{
		ActorLogin: a.ActorLogin,
		Action:     a.Action,
		EntityID:   a.EntityID,
		Meta:       raw,
	}

	if a.ActorID != 0 {
		m.ActorID = &a.ActorID
	}

	return m, nil
}

func modelToAuditLog(m *auditLogModel) *goadmin.AuditLog {
	a := &goadmin.AuditLog{
		ID:         m.ID,
		ActorLogin: m.ActorLogin,
		Action:     m.Action,
		EntityID:   m.EntityID,
		DTCreated:  m.DTCreated,
	}

	if m.ActorID != nil {
		a.ActorID = *m.ActorID
	}

	if len(m.Meta) > 0 {
		_ = json.Unmarshal(m.Meta, &a.Meta)
	}

	return a
}

func (r *auditLogRepository) Create(ctx context.Context, entry *goadmin.AuditLog) (*goadmin.AuditLog, error) {
	model, err := auditLogToModel(entry)
	if err != nil {
		return nil, err
	}

	model.DTCreated = time.Now().UTC()

	if _, err = r.db.NewInsert().Model(model).Exec(ctx); err != nil {
		return nil, fmt.Errorf("insert audit log: %w", err)
	}

	return modelToAuditLog(model), nil
}

func (r *auditLogRepository) Search(ctx context.Context, filter *goadmin.AuditLogFilter) ([]*goadmin.AuditLog, error) {
	var models []auditLogModel

	q := r.db.NewSelect().Model(&models).OrderExpr("al.dt_created DESC")
	q = applyAuditLogFilter(q, filter)

	if err := q.Scan(ctx); err != nil {
		return nil, fmt.Errorf("search audit log: %w", err)
	}

	result := make([]*goadmin.AuditLog, len(models))
	for i := range models {
		result[i] = modelToAuditLog(&models[i])
	}

	return result, nil
}

func (r *auditLogRepository) Count(ctx context.Context, filter *goadmin.AuditLogFilter) (int64, error) {
	q := r.db.NewSelect().Model((*auditLogModel)(nil))
	q = applyAuditLogFilter(q, filter)

	count, err := q.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count audit log: %w", err)
	}

	return int64(count), nil
}

func applyAuditLogFilter(q *bun.SelectQuery, f *goadmin.AuditLogFilter) *bun.SelectQuery {
	if f == nil {
		return q
	}

	if f.Action != "" {
		q = q.Where("al.action = ?", f.Action)
	}

	if f.Limit > 0 {
		q = q.Limit(f.Limit).Offset(f.Offset)
	}

	return q
}
