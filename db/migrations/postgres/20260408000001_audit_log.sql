-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS goadmin.audit_log
(
    id          BIGSERIAL                   NOT NULL,
    actor_id    BIGINT,
    actor_login CHARACTER VARYING(128)      NOT NULL DEFAULT '',
    action      CHARACTER VARYING(50)       NOT NULL,
    entity_id   BIGINT                      NOT NULL DEFAULT 0,
    meta        JSONB                       NOT NULL DEFAULT '{}',
    dt_created  TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT audit_log_pkey PRIMARY KEY (id),
    CONSTRAINT audit_log_actor_fkey FOREIGN KEY (actor_id)
        REFERENCES goadmin."user" (id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_audit_log_actor_id   ON goadmin.audit_log (actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_action     ON goadmin.audit_log (action);
CREATE INDEX IF NOT EXISTS idx_audit_log_dt_created ON goadmin.audit_log (dt_created DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS goadmin.idx_audit_log_dt_created;
DROP INDEX IF EXISTS goadmin.idx_audit_log_action;
DROP INDEX IF EXISTS goadmin.idx_audit_log_actor_id;
DROP TABLE IF EXISTS goadmin.audit_log;
-- +goose StatementEnd
