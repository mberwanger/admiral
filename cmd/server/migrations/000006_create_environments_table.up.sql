CREATE TABLE IF NOT EXISTS environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    cluster_id UUID REFERENCES clusters(id) ON DELETE RESTRICT,
    name VARCHAR(255) NOT NULL CHECK (name ~ '^[a-z0-9]{2,}(-[a-z0-9]*[a-z0-9])?$'),
    namespace VARCHAR(63) CHECK (namespace ~ '^[a-z0-9]{2,}(-[a-z0-9]*[a-z0-9])?$'),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_environment_name ON environments (application_id, name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_environments_application_id ON environments(application_id);
CREATE INDEX IF NOT EXISTS idx_environments_cluster_id ON environments(cluster_id);
CREATE INDEX IF NOT EXISTS idx_environments_namespace ON environments (namespace);
CREATE INDEX IF NOT EXISTS idx_environments_active ON environments(id) WHERE deleted_at IS NULL;
