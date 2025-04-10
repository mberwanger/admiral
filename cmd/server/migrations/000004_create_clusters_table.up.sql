CREATE TABLE IF NOT EXISTS clusters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL CHECK (name ~ '^[a-z0-9]{2,}(-[a-z0-9]*[a-z0-9])?$'),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_cluster_name ON clusters (name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_clusters_active ON clusters(id) WHERE deleted_at IS NULL;
