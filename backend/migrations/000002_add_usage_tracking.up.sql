-- Usage Events: Granular event log for metering
CREATE TABLE IF NOT EXISTS usage_events (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Who generated this usage
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    organization_id INTEGER REFERENCES organizations(id) ON DELETE SET NULL,

    -- What type of usage (api_call, storage, compute, etc.)
    event_type VARCHAR(50) NOT NULL,

    -- Resource identifier (endpoint path, feature name, etc.)
    resource VARCHAR(255) NOT NULL,

    -- Quantity consumed (1 for single events, bytes for storage, ms for compute)
    quantity BIGINT NOT NULL DEFAULT 1,

    -- Unit of measurement (count, bytes, ms, etc.)
    unit VARCHAR(20) NOT NULL DEFAULT 'count',

    -- Additional metadata as JSONB
    metadata JSONB DEFAULT '{}',

    -- Request context
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- Billing period this event belongs to
    billing_period_start DATE NOT NULL,
    billing_period_end DATE NOT NULL
);

-- Indexes for common queries
CREATE INDEX idx_usage_events_user_id ON usage_events(user_id);
CREATE INDEX idx_usage_events_org_id ON usage_events(organization_id);
CREATE INDEX idx_usage_events_event_type ON usage_events(event_type);
CREATE INDEX idx_usage_events_created_at ON usage_events(created_at);
CREATE INDEX idx_usage_events_billing_period ON usage_events(billing_period_start, billing_period_end);
CREATE INDEX idx_usage_events_user_period ON usage_events(user_id, billing_period_start, billing_period_end);

-- Usage Periods: Aggregated totals per billing period
CREATE TABLE IF NOT EXISTS usage_periods (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Who this period belongs to
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    organization_id INTEGER REFERENCES organizations(id) ON DELETE CASCADE,

    -- Billing period
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,

    -- Aggregated usage counts by type (stored as JSONB for flexibility)
    -- e.g., {"api_calls": 1500, "storage_bytes": 1073741824, "compute_ms": 360000}
    usage_totals JSONB NOT NULL DEFAULT '{}',

    -- Limit configuration for this period (from subscription plan)
    usage_limits JSONB DEFAULT '{}',

    -- Whether limits were exceeded
    limits_exceeded BOOLEAN DEFAULT FALSE,

    -- When limits were last checked/updated
    last_aggregated_at TIMESTAMPTZ,

    -- Unique constraint per user/org per period
    CONSTRAINT unique_user_period UNIQUE (user_id, period_start, period_end),
    CONSTRAINT unique_org_period UNIQUE (organization_id, period_start, period_end),

    -- Must have either user_id or organization_id
    CONSTRAINT usage_periods_owner_check CHECK (
        (user_id IS NOT NULL AND organization_id IS NULL) OR
        (user_id IS NULL AND organization_id IS NOT NULL)
    )
);

-- Indexes for usage periods
CREATE INDEX idx_usage_periods_user_id ON usage_periods(user_id);
CREATE INDEX idx_usage_periods_org_id ON usage_periods(organization_id);
CREATE INDEX idx_usage_periods_period ON usage_periods(period_start, period_end);

-- Usage Alerts: Notifications when approaching or exceeding limits
CREATE TABLE IF NOT EXISTS usage_alerts (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Who this alert is for
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    organization_id INTEGER REFERENCES organizations(id) ON DELETE CASCADE,

    -- Alert type (warning_80, warning_90, exceeded, etc.)
    alert_type VARCHAR(50) NOT NULL,

    -- Which usage type triggered this (api_calls, storage, etc.)
    usage_type VARCHAR(50) NOT NULL,

    -- Current usage and limit at time of alert
    current_usage BIGINT NOT NULL,
    usage_limit BIGINT NOT NULL,

    -- Percentage of limit used
    percentage_used INTEGER NOT NULL,

    -- Whether the alert has been acknowledged
    acknowledged BOOLEAN DEFAULT FALSE,
    acknowledged_at TIMESTAMPTZ,
    acknowledged_by INTEGER REFERENCES users(id),

    -- Billing period this alert is for
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,

    -- Prevent duplicate alerts for same type/period
    CONSTRAINT unique_user_alert UNIQUE (user_id, alert_type, usage_type, period_start),
    CONSTRAINT unique_org_alert UNIQUE (organization_id, alert_type, usage_type, period_start)
);

-- Indexes for usage alerts
CREATE INDEX idx_usage_alerts_user_id ON usage_alerts(user_id);
CREATE INDEX idx_usage_alerts_org_id ON usage_alerts(organization_id);
CREATE INDEX idx_usage_alerts_unacknowledged ON usage_alerts(user_id, acknowledged) WHERE acknowledged = FALSE;

-- Function to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_usage_periods_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update updated_at
CREATE TRIGGER trigger_usage_periods_updated_at
    BEFORE UPDATE ON usage_periods
    FOR EACH ROW
    EXECUTE FUNCTION update_usage_periods_updated_at();
