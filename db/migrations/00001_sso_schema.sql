-- Gerege SSO's own schema — the four modules this distribution adds.
--
-- This history belongs to this repository, not to the platform. It is applied
-- with MIGRATIONS_DIR and a MIGRATIONS_TABLE of its own
-- (goose_db_version_sso), which is not optional: goose writes one row per
-- applied version in one table, so the core's 00001 and this 00001 would be
-- the same row and whichever ran second would be recorded as already applied.
--
-- Two rules hold for everything below.
--
-- 1. Every name is prefixed `sso_`. A table this repository owns must never be
--    able to collide with one a future core release adds.
--
-- 2. Nothing here touches a table the core owns. No ALTER, no policy, no
--    trigger on `sessions`, `devices` or the RBAC tables — the core's history
--    knows nothing about this file, and a column added from here is a column
--    the core's next migration would find already present and unexplained.
--
-- Row-level security is set per table rather than by the sweep in the core's
-- 00029: that loop ran once, over the tables that existed then, and cannot
-- reach a table created afterwards by another repository. The fork this
-- product replaced learned that the expensive way — its OAuth2 tables sat
-- outside the policies until somebody noticed.

-- +goose Up

-- ---------------------------------------------------------------- federation

-- An external identity provider this installation trusts.
--
-- Per tenant, because trust is: one organisation federating with its parent
-- ministry says nothing about what any other organisation on the same
-- installation should accept.
CREATE TABLE IF NOT EXISTS sso_federation_providers (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    -- The OIDC issuer. Identity is the issuer, not the display name: two
    -- tenants may call the same provider different things, and one tenant must
    -- not register the same issuer twice under two names and then wonder which
    -- of them a login came through.
    issuer       TEXT NOT NULL,
    client_id    TEXT NOT NULL,
    -- Encrypted at rest with SSO_FEDERATION_KEY and never returned by the API.
    -- A console that can read back the credentials of every federation an
    -- installation holds is a console worth stealing for that alone.
    client_secret_encrypted BYTEA NOT NULL,
    -- Requested scopes, and how the claims that come back map onto this
    -- platform's fields. Both are provider-specific and neither is worth a
    -- table of its own.
    scopes        TEXT NOT NULL DEFAULT 'openid profile email',
    attribute_map JSONB NOT NULL DEFAULT '{}'::jsonb,
    -- Disabled rather than deleted is the ordinary way to stop trusting a
    -- provider: the links below stay, so the record of who arrived through it
    -- survives the decision to stop accepting new arrivals.
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  UUID,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, issuer)
);

-- Who somebody is over there, and who they are here.
--
-- The subject is the provider's identifier for a person and is the only thing
-- that survives a name change on either side, so it is what the link is keyed
-- by. Deleting a provider takes its links with it: they mean nothing without
-- the issuer that produced them.
CREATE TABLE IF NOT EXISTS sso_federation_links (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    provider_id UUID NOT NULL REFERENCES sso_federation_providers(id) ON DELETE CASCADE,
    subject     TEXT NOT NULL,
    user_id     UUID NOT NULL,
    linked_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider_id, subject)
);

CREATE INDEX IF NOT EXISTS idx_sso_federation_providers_tenant
    ON sso_federation_providers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sso_federation_links_user
    ON sso_federation_links(tenant_id, user_id);

-- ------------------------------------------------------------------ sessions

-- What this module did to a session, and who asked for it.
--
-- The sessions themselves are the core's. This table holds the one thing the
-- core does not record: that an administrator reached in and ended somebody's
-- session, when, and why. Cutting a person off mid-work is an act somebody has
-- to be able to be asked about afterwards.
CREATE TABLE IF NOT EXISTS sso_session_events (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL,
    -- Not a foreign key. The core owns `sessions` and deletes rows from it as
    -- they expire; a reference would either block that or take the record of
    -- the revocation away with it, and the record is the entire point.
    session_id UUID,
    user_id    UUID NOT NULL,
    action     VARCHAR(32) NOT NULL,
    actor_id   UUID,
    reason     TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sso_session_events_tenant_time
    ON sso_session_events(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sso_session_events_user
    ON sso_session_events(tenant_id, user_id, created_at DESC);

-- ------------------------------------------------------------- access review

-- One round of asking whether the access people hold is still the access they
-- need.
CREATE TABLE IF NOT EXISTS sso_review_campaigns (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name      VARCHAR(255) NOT NULL,
    -- What is being reviewed: everything, one app's permissions, or one role's
    -- holders. Narrow campaigns are the ones that get finished.
    scope     VARCHAR(16) NOT NULL DEFAULT 'all',
    scope_ref VARCHAR(128) NOT NULL DEFAULT '',
    due_date  DATE,
    -- draft → open → closed, and no way back. Reopening a closed campaign
    -- would let a decision be changed after it was reported as final, which is
    -- the one thing an attestation record must not allow.
    status     VARCHAR(16) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    opened_at  TIMESTAMPTZ,
    closed_at  TIMESTAMPTZ,
    CONSTRAINT sso_review_campaigns_scope_check
        CHECK (scope IN ('all', 'app', 'role')),
    CONSTRAINT sso_review_campaigns_status_check
        CHECK (status IN ('draft', 'open', 'closed'))
);

-- One person's one permission, as it stood when the campaign opened.
--
-- Copied out of RBAC at open time rather than read live. A campaign that
-- queried the current state would change under the reviewer: a role edited on
-- Tuesday would silently rewrite what somebody attested to on Monday, and the
-- report at the end would describe a set of decisions nobody made.
CREATE TABLE IF NOT EXISTS sso_review_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    campaign_id UUID NOT NULL REFERENCES sso_review_campaigns(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL,
    user_email  VARCHAR(255) NOT NULL DEFAULT '',
    role_id     UUID,
    role_name   VARCHAR(255) NOT NULL DEFAULT '',
    permission_code VARCHAR(128) NOT NULL,
    status      VARCHAR(16) NOT NULL DEFAULT 'pending',
    reviewer_id UUID,
    decided_at  TIMESTAMPTZ,
    CONSTRAINT sso_review_items_status_check
        CHECK (status IN ('pending', 'kept', 'revoked')),
    UNIQUE (campaign_id, user_id, permission_code)
);

-- Every decision ever recorded, including the ones that were changed.
--
-- The item carries the current answer; this carries the history of answers. A
-- reviewer who keeps an access on Monday and revokes it on Thursday has made
-- two decisions, and an attestation trail that only remembers the second one
-- cannot answer why the first was made.
CREATE TABLE IF NOT EXISTS sso_review_decisions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    item_id     UUID NOT NULL REFERENCES sso_review_items(id) ON DELETE CASCADE,
    decision    VARCHAR(16) NOT NULL,
    reviewer_id UUID,
    note        TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT sso_review_decisions_decision_check
        CHECK (decision IN ('kept', 'revoked'))
);

CREATE INDEX IF NOT EXISTS idx_sso_review_campaigns_tenant
    ON sso_review_campaigns(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sso_review_items_campaign
    ON sso_review_items(campaign_id, status);
CREATE INDEX IF NOT EXISTS idx_sso_review_decisions_item
    ON sso_review_decisions(item_id, created_at DESC);

-- -------------------------------------------------------------- provisioning

-- A system this installation pushes user records into.
CREATE TABLE IF NOT EXISTS sso_scim_targets (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name      VARCHAR(255) NOT NULL,
    base_url  TEXT NOT NULL,
    -- Encrypted with the same key as a federation secret and, like it, never
    -- read back through the API.
    token_encrypted BYTEA NOT NULL,
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, name)
);

-- Work waiting to be pushed.
--
-- A queue rather than a synchronous call because the target is somebody else's
-- server: it can be down for an afternoon, and a user who could not be created
-- in this platform because a remote system was unreachable is a platform whose
-- availability is the worst of everyone it integrates with.
CREATE TABLE IF NOT EXISTS sso_scim_queue (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    target_id UUID NOT NULL REFERENCES sso_scim_targets(id) ON DELETE CASCADE,
    op        VARCHAR(16) NOT NULL,
    user_id   UUID NOT NULL,
    payload   JSONB NOT NULL DEFAULT '{}'::jsonb,
    attempts  INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT sso_scim_queue_op_check
        CHECK (op IN ('create', 'update', 'deactivate'))
);

-- What was sent and what came back.
CREATE TABLE IF NOT EXISTS sso_scim_log (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    target_id UUID NOT NULL REFERENCES sso_scim_targets(id) ON DELETE CASCADE,
    op        VARCHAR(16) NOT NULL,
    user_id   UUID,
    status_code INT NOT NULL DEFAULT 0,
    -- Truncated on write. A remote system that answers an error with a page of
    -- HTML should cost this database a line, not a page.
    response_excerpt TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- The claim order of the worker's query: due work, oldest first.
CREATE INDEX IF NOT EXISTS idx_sso_scim_queue_due
    ON sso_scim_queue(next_attempt_at, created_at);
CREATE INDEX IF NOT EXISTS idx_sso_scim_log_target_time
    ON sso_scim_log(target_id, created_at DESC);

-- ------------------------------------------------------- grants and policies

-- The application role is what the platform switches into for tenant-scoped
-- work (see the core's dbguard), so every table here needs the same grants the
-- core's own tables have — and the policies the core's sweep could not reach.
--
-- Background work is deliberately outside this: the SCIM worker runs with no
-- tenant in its context, which dbguard binds to the login role rather than to
-- gerege_nexus_app, so it sees the queue across every organisation. That is the
-- same path the platform's own housekeeping sweeps take.
-- +goose StatementBegin
DO $$
DECLARE
    target TEXT;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'gerege_nexus_app') THEN
        RETURN;
    END IF;
    FOREACH target IN ARRAY ARRAY[
        'sso_federation_providers', 'sso_federation_links',
        'sso_session_events',
        'sso_review_campaigns', 'sso_review_items', 'sso_review_decisions',
        'sso_scim_targets', 'sso_scim_queue', 'sso_scim_log'
    ]
    LOOP
        EXECUTE format('GRANT SELECT, INSERT, UPDATE, DELETE ON public.%I TO gerege_nexus_app', target);
        EXECUTE format('ALTER TABLE public.%I ENABLE ROW LEVEL SECURITY', target);
        EXECUTE format('ALTER TABLE public.%I FORCE ROW LEVEL SECURITY', target);
        EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON public.%I', target);
        EXECUTE format(
            'CREATE POLICY tenant_isolation ON public.%I TO gerege_nexus_app '
            'USING (tenant_id IS NULL OR tenant_id = NULLIF(current_setting(''app.current_tenant'', true), '''')::uuid) '
            'WITH CHECK (tenant_id = NULLIF(current_setting(''app.current_tenant'', true), '''')::uuid)',
            target);
    END LOOP;
END
$$;
-- +goose StatementEnd

-- +goose Down

DROP TABLE IF EXISTS sso_scim_log;
DROP TABLE IF EXISTS sso_scim_queue;
DROP TABLE IF EXISTS sso_scim_targets;
DROP TABLE IF EXISTS sso_review_decisions;
DROP TABLE IF EXISTS sso_review_items;
DROP TABLE IF EXISTS sso_review_campaigns;
DROP TABLE IF EXISTS sso_session_events;
DROP TABLE IF EXISTS sso_federation_links;
DROP TABLE IF EXISTS sso_federation_providers;
