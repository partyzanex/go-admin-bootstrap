CREATE SCHEMA IF NOT EXISTS goadmin;

CREATE TYPE goadmin.USER_STATUS AS ENUM ('new', 'active', 'blocked');
CREATE TYPE goadmin.USER_ROLE AS ENUM ('owner', 'root', 'user');

CREATE TABLE IF NOT EXISTS goadmin."user"
(
    id             BIGSERIAL                   NOT NULL,
    login          CHARACTER VARYING(128)      NOT NULL,
    password       CHARACTER(64)               NOT NULL,
    status         goadmin.USER_STATUS         NOT NULL DEFAULT 'new',
    name           CHARACTER VARYING(255)      NOT NULL,
    role           goadmin.USER_ROLE           NOT NULL,
    dt_created     TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    dt_updated     TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    dt_last_logged TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    CONSTRAINT user_pkey PRIMARY KEY (id),
    CONSTRAINT user_login_ukey UNIQUE (login)
);

CREATE TYPE goadmin.TOKEN_TYPE AS ENUM ('auth');

CREATE TABLE IF NOT EXISTS goadmin.auth_token
(
    id         BIGSERIAL                   NOT NULL,
    user_id    BIGINT                      NOT NULL,
    token      CHARACTER(64)               NOT NULL,
    type       goadmin.TOKEN_TYPE          NOT NULL,
    dt_expired TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    dt_created TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT auth_token_pkey PRIMARY KEY (id),
    CONSTRAINT auth_token_ukey UNIQUE (token),
    CONSTRAINT user_auth_token_fkey FOREIGN KEY (user_id) REFERENCES goadmin."user" (id)
        ON DELETE CASCADE
);
