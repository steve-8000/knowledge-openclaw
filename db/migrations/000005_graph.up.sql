-- 000005_graph.up.sql
-- Knowledge graph edges between documents

CREATE TABLE edges (
    tenant_id       uuid        NOT NULL,
    edge_id         uuid        NOT NULL DEFAULT uuid_generate_v4(),
    from_doc_id     uuid        NOT NULL,
    to_doc_id       uuid,
    to_external_key text,
    relation        text        NOT NULL CHECK (relation IN ('links_to','references','supersedes','duplicates','contradicts')),
    evidence        jsonb       NOT NULL DEFAULT '{}',
    created_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, edge_id),
    FOREIGN KEY (tenant_id, from_doc_id) REFERENCES documents(tenant_id, doc_id) ON DELETE CASCADE,
    -- to_doc_id is nullable (external links use to_external_key instead)
    CHECK (to_doc_id IS NOT NULL OR to_external_key IS NOT NULL)
);

-- Index for graph traversal queries
CREATE INDEX idx_edges_from ON edges (tenant_id, from_doc_id, relation);
CREATE INDEX idx_edges_to   ON edges (tenant_id, to_doc_id, relation) WHERE to_doc_id IS NOT NULL;

-- Supersedes chain index (for finding latest document in chain)
CREATE INDEX idx_edges_supersedes ON edges (tenant_id, to_doc_id)
    WHERE relation = 'supersedes';
