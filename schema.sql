-- =========================
--  Enhanced Schema for Security Detector V2
--  Adds persistent monitoring state, audit trail, and AI analysis support
-- =========================

-- Original tables (runtime, framework, app) remain unchanged
-- Enhanced dependencies table for better tracking
CREATE TABLE IF NOT EXISTS dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    owner TEXT NOT NULL,
    repo TEXT NOT NULL,
    last_commit_sha TEXT,
    last_commit_at TIMESTAMPTZ,
    last_tag TEXT,
    last_tag_at TIMESTAMPTZ,
    default_branch TEXT DEFAULT 'main',
    repository_url TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(owner, repo)
);

-- Enhanced app_dependencies for monitoring state
CREATE TABLE IF NOT EXISTS app_dependencies (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id           UUID        NOT NULL REFERENCES app(id),
    dependency_id    UUID        NOT NULL REFERENCES dependencies(id),
    used_commit_sha  VARCHAR(64),
    used_version     VARCHAR(128) NOT NULL,
    used_tag         VARCHAR(128),
    
    -- Monitoring configuration
    is_monitored     BOOLEAN     NOT NULL DEFAULT FALSE,
    monitoring_enabled BOOLEAN   NOT NULL DEFAULT TRUE,
    polling_interval_minutes INT NOT NULL DEFAULT 60,
    
    -- Persistent state for monitoring
    last_checked_commit_sha TEXT,
    last_checked_tag TEXT,
    last_checked_at TIMESTAMPTZ,
    last_monitored_at TIMESTAMPTZ,
    monitor_status   VARCHAR(32) DEFAULT 'ready', -- ready, running, paused, error
    
    -- Audit and tracking
    total_checks_count INT DEFAULT 0,
    last_security_detection_at TIMESTAMPTZ,
    last_security_score INT DEFAULT 0,
    
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE(app_id, dependency_id)
);

-- Dependency Versions table (tracks commits/tags seen over time)
CREATE TABLE IF NOT EXISTS dependency_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dependency_id UUID NOT NULL REFERENCES dependencies(id) ON DELETE CASCADE,
    commit_sha TEXT NOT NULL,
    commit_at TIMESTAMPTZ NOT NULL,
    tag TEXT,
    branch TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- AI Analysis Results table - optimized for AnalysisResult struct storage
CREATE TABLE IF NOT EXISTS ai_analysis_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dependency_id UUID NOT NULL REFERENCES dependencies(id),
    app_id UUID REFERENCES app(id), -- Optional reference
    
    -- Analysis context
    filename TEXT NOT NULL,       -- The identifier (tag/version) being analyzed
    repository TEXT NOT NULL,     -- Repository name for reference

    -- AI response data
    response TEXT NOT NULL,       -- Full AI response
    duration BIGINT NOT NULL,     -- Analysis duration in nanoseconds
    success BOOLEAN NOT NULL,     -- Whether analysis succeeded
    
    -- Parsed summary fields for efficient querying
    classification TEXT,          -- From AnalysisSummary
    severity TEXT,               -- From AnalysisSummary  
    confidence TEXT,             -- From AnalysisSummary
    key_findings JSONB,          -- JSON array of key findings
    summary TEXT,                -- Summary text
    risk_level TEXT,             -- Risk assessment
    action_required TEXT,        -- Recommended action
    processing_time TEXT,        -- Duration as human-readable string
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- AI Processing Queue
CREATE TABLE IF NOT EXISTS ai_processing_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    analysis_id UUID NOT NULL REFERENCES ai_analysis_results(id),
    status TEXT NOT NULL DEFAULT 'queued', -- queued, processing, completed, failed
    priority INT NOT NULL DEFAULT 5,        -- 1=critical, 2=high, 3=medium, 4=low, 5=minimal
    queued_at TIMESTAMPTZ DEFAULT NOW(),
    processing_started_at TIMESTAMPTZ,
    processed_at TIMESTAMPTZ,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3
);

-- Monitoring Jobs table (for tracking active monitoring processes)
CREATE TABLE IF NOT EXISTS monitoring_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_type TEXT NOT NULL,      -- scheduled, manual, on_demand
    status TEXT NOT NULL,        -- running, completed, failed, cancelled
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    
    -- Job configuration
    app_ids TEXT,               -- JSON array of app UUIDs (database-agnostic)
    dependency_ids TEXT,        -- JSON array of dependency UUIDs (database-agnostic)
    polling_interval_minutes INT DEFAULT 60,
    max_concurrent_checks INT DEFAULT 10,
    
    -- Progress tracking
    total_checks_planned INT DEFAULT 0,
    checks_completed INT DEFAULT 0,
    checks_failed INT DEFAULT 0,
    security_detections INT DEFAULT 0,
    
    -- Results
    results_summary JSONB,
    error_log TEXT,
    
    created_by TEXT,            -- user or system identifier
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Audit Trail table
CREATE TABLE IF NOT EXISTS audit_trail (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type TEXT NOT NULL,   -- app, dependency, monitoring_job, etc.
    entity_id UUID NOT NULL,
    action TEXT NOT NULL,        -- created, updated, deleted, monitored, etc.
    old_values JSONB,
    new_values JSONB,
    performed_by TEXT,           -- user or system identifier
    performed_at TIMESTAMPTZ DEFAULT NOW(),
    context JSONB,               -- additional context (IP, user agent, etc.)
    
    -- For security-specific events
    security_relevant BOOLEAN DEFAULT FALSE,
    risk_level TEXT              -- for security events
);

-- Monitoring Configuration table
CREATE TABLE IF NOT EXISTS monitoring_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key TEXT UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    config_type TEXT NOT NULL,   -- string, int, boolean, json
    description TEXT,
    is_system_config BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- =========================
--  Indexes for Performance
-- =========================

-- App Dependencies indexes
CREATE INDEX IF NOT EXISTS idx_app_dependencies_monitoring ON app_dependencies(is_monitored, monitoring_enabled);
CREATE INDEX IF NOT EXISTS idx_app_dependencies_status ON app_dependencies(monitor_status);
CREATE INDEX IF NOT EXISTS idx_app_dependencies_last_checked ON app_dependencies(last_checked_at);
CREATE INDEX IF NOT EXISTS idx_app_dependencies_polling ON app_dependencies(polling_interval_minutes, last_checked_at);

-- Dependency Versions indexes
CREATE INDEX IF NOT EXISTS idx_dependency_versions_dependency_id ON dependency_versions(dependency_id);
CREATE INDEX IF NOT EXISTS idx_dependency_versions_commit_at ON dependency_versions(commit_at DESC);
CREATE INDEX IF NOT EXISTS idx_dependency_versions_tag ON dependency_versions(tag);

-- AI Analysis Results indexes - optimized for dependency ID querying
CREATE INDEX IF NOT EXISTS idx_ai_results_dependency ON ai_analysis_results(dependency_id);
CREATE INDEX IF NOT EXISTS idx_ai_results_app ON ai_analysis_results(app_id);
CREATE INDEX IF NOT EXISTS idx_ai_results_classification ON ai_analysis_results(classification);
CREATE INDEX IF NOT EXISTS idx_ai_results_severity ON ai_analysis_results(severity);
CREATE INDEX IF NOT EXISTS idx_ai_results_risk ON ai_analysis_results(risk_level);
CREATE INDEX IF NOT EXISTS idx_ai_results_created ON ai_analysis_results(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_results_success ON ai_analysis_results(success);
CREATE INDEX IF NOT EXISTS idx_ai_results_repository ON ai_analysis_results(repository);

-- AI Processing Queue indexes
CREATE INDEX IF NOT EXISTS idx_ai_queue_status_priority ON ai_processing_queue(status, priority, queued_at);
CREATE INDEX IF NOT EXISTS idx_ai_queue_analysis ON ai_processing_queue(analysis_id);

-- Monitoring Jobs indexes
CREATE INDEX IF NOT EXISTS idx_monitoring_jobs_status ON monitoring_jobs(status);
CREATE INDEX IF NOT EXISTS idx_monitoring_jobs_started ON monitoring_jobs(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_monitoring_jobs_type ON monitoring_jobs(job_type);

-- Audit Trail indexes
CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit_trail(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_performed_at ON audit_trail(performed_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_security ON audit_trail(security_relevant, risk_level);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_trail(action);

-- Monitoring Config indexes
CREATE INDEX IF NOT EXISTS idx_monitoring_config_key ON monitoring_config(config_key);
CREATE INDEX IF NOT EXISTS idx_monitoring_config_system ON monitoring_config(is_system_config);

-- =========================
--  Default Configuration Values
-- =========================

INSERT INTO monitoring_config (config_key, config_value, config_type, description) VALUES
('default_polling_interval_minutes', '60', 'int', 'Default polling interval for new dependencies'),
('max_concurrent_monitoring_jobs', '10', 'int', 'Maximum number of concurrent monitoring jobs'),
('ai_processing_enabled', 'true', 'boolean', 'Enable AI processing of security detections'),
('ai_queue_batch_size', '20', 'int', 'Batch size for AI processing queue'),
('audit_retention_days', '365', 'int', 'Number of days to retain audit trail records'),
('notification_webhook_url', '', 'string', 'Webhook URL for security notifications'),
('notification_threshold_score', '60', 'int', 'Minimum security score for notifications'),
('max_retry_attempts', '3', 'int', 'Maximum retry attempts for failed operations'),
('monitoring_timeout_minutes', '30', 'int', 'Timeout for individual monitoring operations'),
('batch_processing_size', '50', 'int', 'Batch size for bulk operations')
ON CONFLICT (config_key) DO NOTHING;
