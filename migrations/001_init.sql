-- RouteX 初始化 schema —— 仅作参考文档使用，不会被自动执行。
--
-- 实际建表由 Go 后端启动时的 GORM AutoMigrate 完成（见 server/internal/db/db.go）。
-- 这份文件是给运维/DBA 看的 schema 速览；如果某个字段在这里和 model 不一致，
-- 以 model 为准（GORM 是 source of truth）。

CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    username        VARCHAR(64)  UNIQUE NOT NULL,
    uuid            VARCHAR(128) UNIQUE NOT NULL,
    password        VARCHAR(128),
    status          INT          DEFAULT 1,
    traffic_limit   BIGINT       DEFAULT 0,
    used_traffic    BIGINT       DEFAULT 0,
    expire_time     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  DEFAULT now(),
    updated_at      TIMESTAMPTZ  DEFAULT now()
);

CREATE TABLE IF NOT EXISTS nodes (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(64),
    ip          VARCHAR(64),
    region      VARCHAR(64),
    status      INT          DEFAULT 1,
    cpu         DOUBLE PRECISION DEFAULT 0,
    memory      DOUBLE PRECISION DEFAULT 0,
    bandwidth   BIGINT       DEFAULT 0,
    created_at  TIMESTAMPTZ  DEFAULT now(),
    updated_at  TIMESTAMPTZ  DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_traffic (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    upload      BIGINT DEFAULT 0,
    download    BIGINT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_user_traffic_user_id    ON user_traffic(user_id);
CREATE INDEX IF NOT EXISTS idx_user_traffic_created_at ON user_traffic(created_at);

CREATE TABLE IF NOT EXISTS proxy_providers (
    id          BIGSERIAL PRIMARY KEY,
    type        VARCHAR(32),
    host        VARCHAR(128),
    port        INT,
    username    VARCHAR(128),
    password    VARCHAR(128),
    region      VARCHAR(64),
    status      INT          DEFAULT 1,
    created_at  TIMESTAMPTZ  DEFAULT now(),
    updated_at  TIMESTAMPTZ  DEFAULT now()
);
