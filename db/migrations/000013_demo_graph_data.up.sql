-- 000013_demo_graph_data.up.sql
-- Rich demo dataset for Knowledge Graph visualization
-- 34 documents + ~80 edges forming a realistic tech company knowledge base

-- ───────────────────────────────────────────────────────────────
-- 1. Extend doc_type CHECK to include spec, incident, runbook
-- ───────────────────────────────────────────────────────────────
ALTER TABLE documents
  DROP CONSTRAINT IF EXISTS documents_doc_type_check;

ALTER TABLE documents
  ADD CONSTRAINT documents_doc_type_check
  CHECK (doc_type IN (
    'report','adr','postmortem','snippet','glossary','guide',
    'policy','other','spec','incident','runbook'
  ));

-- ───────────────────────────────────────────────────────────────
-- 2. Documents (34 total)
--    UUID pattern: d1000000-0000-0000-0000-0000000000XX
--    tenant: 00000000-0000-0000-0000-000000000001
-- ───────────────────────────────────────────────────────────────

-- Policies (6)
INSERT INTO documents (tenant_id, doc_id, stable_key, title, doc_type, status, confidence, owners, tags, source) VALUES
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000001', 'demo/policy/data-retention',       'Data Retention Policy',         'policy',   'published', 'high', '["compliance-team"]',  '["compliance","data","gdpr"]',       '{"origin":"confluence","url":"https://wiki.internal/policy/data-retention"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000002', 'demo/policy/access-control',       'Access Control Policy',         'policy',   'published', 'high', '["security-team"]',    '["security","iam","rbac"]',          '{"origin":"confluence","url":"https://wiki.internal/policy/access-control"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000003', 'demo/policy/incident-response',    'Incident Response Policy',      'policy',   'published', 'high', '["sre-team"]',         '["incident","sre","oncall"]',        '{"origin":"confluence","url":"https://wiki.internal/policy/incident-response"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000004', 'demo/policy/code-review',          'Code Review Standards',         'policy',   'published', 'med',  '["engineering"]',      '["code-review","quality","pr"]',     '{"origin":"confluence","url":"https://wiki.internal/policy/code-review"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000005', 'demo/policy/deployment',           'Deployment Policy',             'policy',   'published', 'high', '["platform-team"]',    '["deploy","ci-cd","release"]',       '{"origin":"confluence","url":"https://wiki.internal/policy/deployment"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000006', 'demo/policy/security-baseline',    'Security Baseline Standards',   'policy',   'published', 'high', '["security-team"]',    '["security","compliance","baseline"]','{"origin":"confluence","url":"https://wiki.internal/policy/security-baseline"}')
ON CONFLICT (tenant_id, stable_key) DO NOTHING;

-- Specs (6)
INSERT INTO documents (tenant_id, doc_id, stable_key, title, doc_type, status, confidence, owners, tags, source) VALUES
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000007', 'demo/spec/api-v2',                'API v2 Specification',          'spec',     'published', 'high', '["api-team"]',         '["api","rest","openapi"]',           '{"origin":"github","url":"https://github.com/openclaw/api-spec"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000008', 'demo/spec/auth-system',           'Authentication System Design',  'spec',     'published', 'high', '["identity-team"]',    '["auth","oauth","jwt"]',             '{"origin":"github","url":"https://github.com/openclaw/auth-spec"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000009', 'demo/spec/search-engine',         'Search Engine Architecture',    'spec',     'published', 'med',  '["search-team"]',      '["search","bm25","vector"]',         '{"origin":"notion","url":"https://notion.so/search-arch"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000a', 'demo/spec/notification',          'Notification Service Spec',     'spec',     'inbox',     'med',  '["platform-team"]',    '["notification","events","email"]',  '{"origin":"notion","url":"https://notion.so/notif-spec"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000b', 'demo/spec/payment-gateway',       'Payment Gateway Integration',   'spec',     'published', 'high', '["payments-team"]',    '["payment","stripe","billing"]',     '{"origin":"confluence","url":"https://wiki.internal/spec/payment"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000c', 'demo/spec/data-pipeline',         'Data Pipeline Design',          'spec',     'published', 'med',  '["data-team"]',        '["etl","pipeline","streaming"]',     '{"origin":"notion","url":"https://notion.so/data-pipeline"}')
ON CONFLICT (tenant_id, stable_key) DO NOTHING;

-- Incidents (5)
INSERT INTO documents (tenant_id, doc_id, stable_key, title, doc_type, status, confidence, owners, tags, source) VALUES
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000d', 'demo/incident/db-outage-2024',     'Database Outage - Jan 2024',    'incident', 'published', 'high', '["sre-team"]',         '["incident","database","outage"]',   '{"origin":"pagerduty","url":"https://incidents.internal/INC-2024-001"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000e', 'demo/incident/api-latency-spike',  'API Latency Spike P1',          'incident', 'published', 'med',  '["api-team","sre-team"]','["incident","latency","api"]',      '{"origin":"pagerduty","url":"https://incidents.internal/INC-2024-002"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000f', 'demo/incident/auth-bypass',        'Authentication Bypass CVE',     'incident', 'published', 'high', '["security-team"]',    '["incident","security","cve"]',      '{"origin":"jira","url":"https://jira.internal/SEC-2024-007"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000010', 'demo/incident/payment-errors',     'Payment Processing Failures',   'incident', 'published', 'med',  '["payments-team"]',    '["incident","payment","errors"]',    '{"origin":"pagerduty","url":"https://incidents.internal/INC-2024-003"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000011', 'demo/incident/dns-failure',        'DNS Resolution Failure',        'incident', 'archived',  'high', '["infra-team"]',       '["incident","dns","networking"]',    '{"origin":"pagerduty","url":"https://incidents.internal/INC-2023-019"}')
ON CONFLICT (tenant_id, stable_key) DO NOTHING;

-- Runbooks (5)
INSERT INTO documents (tenant_id, doc_id, stable_key, title, doc_type, status, confidence, owners, tags, source) VALUES
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000012', 'demo/runbook/db-failover',         'Database Failover Procedure',   'runbook',  'published', 'high', '["sre-team","dba"]',   '["runbook","database","failover"]',  '{"origin":"confluence","url":"https://wiki.internal/runbook/db-failover"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000013', 'demo/runbook/scale-api',           'API Horizontal Scaling',        'runbook',  'published', 'med',  '["sre-team"]',         '["runbook","scaling","kubernetes"]',  '{"origin":"confluence","url":"https://wiki.internal/runbook/scale-api"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000014', 'demo/runbook/rotate-secrets',      'Secret Rotation Runbook',       'runbook',  'published', 'high', '["security-team"]',    '["runbook","secrets","rotation"]',    '{"origin":"confluence","url":"https://wiki.internal/runbook/rotate-secrets"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000015', 'demo/runbook/deploy-hotfix',       'Hotfix Deployment Procedure',   'runbook',  'published', 'med',  '["platform-team"]',    '["runbook","deploy","hotfix"]',       '{"origin":"confluence","url":"https://wiki.internal/runbook/deploy-hotfix"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000016', 'demo/runbook/restore-backup',      'Backup Restore Procedure',      'runbook',  'published', 'high', '["sre-team","dba"]',   '["runbook","backup","restore"]',      '{"origin":"confluence","url":"https://wiki.internal/runbook/restore-backup"}')
ON CONFLICT (tenant_id, stable_key) DO NOTHING;

-- Guides (5)
INSERT INTO documents (tenant_id, doc_id, stable_key, title, doc_type, status, confidence, owners, tags, source) VALUES
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000017', 'demo/guide/onboarding',           'New Engineer Onboarding',       'guide',    'published', 'high', '["engineering"]',      '["guide","onboarding","setup"]',      '{"origin":"confluence","url":"https://wiki.internal/guide/onboarding"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000018', 'demo/guide/dev-environment',      'Dev Environment Setup',         'guide',    'published', 'med',  '["platform-team"]',    '["guide","dev","environment"]',       '{"origin":"confluence","url":"https://wiki.internal/guide/dev-setup"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000019', 'demo/guide/monitoring',           'Monitoring & Alerting Guide',   'guide',    'published', 'high', '["sre-team"]',         '["guide","monitoring","alerting"]',   '{"origin":"notion","url":"https://notion.so/monitoring-guide"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000001a', 'demo/guide/testing-strategy',     'Testing Strategy Guide',        'guide',    'published', 'med',  '["engineering"]',      '["guide","testing","qa"]',            '{"origin":"confluence","url":"https://wiki.internal/guide/testing"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000001b', 'demo/guide/ci-cd-pipeline',       'CI/CD Pipeline Guide',          'guide',    'published', 'high', '["platform-team"]',    '["guide","ci-cd","pipeline"]',        '{"origin":"confluence","url":"https://wiki.internal/guide/ci-cd"}')
ON CONFLICT (tenant_id, stable_key) DO NOTHING;

-- ADRs (4)
INSERT INTO documents (tenant_id, doc_id, stable_key, title, doc_type, status, confidence, owners, tags, source) VALUES
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000001c', 'demo/adr/use-postgres',           'ADR-001: Use PostgreSQL',       'adr',      'published', 'high', '["architecture"]',     '["adr","database","postgres"]',       '{"origin":"github","url":"https://github.com/openclaw/adr/001"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000001d', 'demo/adr/event-driven',           'ADR-002: Event-Driven Arch',    'adr',      'published', 'high', '["architecture"]',     '["adr","events","architecture"]',     '{"origin":"github","url":"https://github.com/openclaw/adr/002"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000001e', 'demo/adr/graphql-vs-rest',        'ADR-003: GraphQL vs REST',      'adr',      'published', 'med',  '["api-team"]',         '["adr","api","graphql","rest"]',       '{"origin":"github","url":"https://github.com/openclaw/adr/003"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000001f', 'demo/adr/kubernetes-adoption',    'ADR-004: Kubernetes Adoption',  'adr',      'published', 'high', '["infra-team"]',       '["adr","kubernetes","infra"]',         '{"origin":"github","url":"https://github.com/openclaw/adr/004"}')
ON CONFLICT (tenant_id, stable_key) DO NOTHING;

-- Reports (2) + Glossary (1)
INSERT INTO documents (tenant_id, doc_id, stable_key, title, doc_type, status, confidence, owners, tags, source) VALUES
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000020', 'demo/report/q1-review',           'Q1 2024 Engineering Review',    'report',   'published', 'med',  '["engineering"]',      '["report","quarterly","metrics"]',    '{"origin":"slides","url":"https://slides.internal/q1-review"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000021', 'demo/report/security-audit',      'Annual Security Audit Report',  'report',   'published', 'high', '["security-team"]',    '["report","security","audit"]',       '{"origin":"sharepoint","url":"https://sharepoint.internal/security-audit-2024"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000022', 'demo/glossary/engineering',       'Engineering Glossary',          'glossary', 'published', 'med',  '["engineering"]',      '["glossary","terms","definitions"]',  '{"origin":"confluence","url":"https://wiki.internal/glossary"}')
ON CONFLICT (tenant_id, stable_key) DO NOTHING;

-- ───────────────────────────────────────────────────────────────
-- 3. Edges (~82 total)
--    Hub nodes: security-baseline(06), api-v2(07), auth-system(08)
--    Each hub has 8+ connections
-- ───────────────────────────────────────────────────────────────

-- Helper: tenant shorthand
-- t = '00000000-0000-0000-0000-000000000001'
-- d(XX) = 'd1000000-0000-0000-0000-0000000000XX'

-- ═══════════════════════════════════════════════════════════════
-- Security Baseline (06) — Hub #1: 10 outgoing edges
-- ═══════════════════════════════════════════════════════════════
INSERT INTO edges (tenant_id, from_doc_id, to_doc_id, relation, evidence) VALUES
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000006', 'd1000000-0000-0000-0000-000000000001', 'references', '{"section":"data-at-rest encryption"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000006', 'd1000000-0000-0000-0000-000000000002', 'references', '{"section":"rbac requirements"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000006', 'd1000000-0000-0000-0000-000000000003', 'references', '{"section":"incident classification"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000006', 'd1000000-0000-0000-0000-000000000008', 'references', '{"section":"auth requirements"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000006', 'd1000000-0000-0000-0000-00000000000f', 'links_to',   '{"reason":"CVE triggered baseline update"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000006', 'd1000000-0000-0000-0000-000000000014', 'links_to',   '{"reason":"secret rotation requirement"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000006', 'd1000000-0000-0000-0000-000000000021', 'references', '{"section":"audit compliance mapping"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000006', 'd1000000-0000-0000-0000-000000000005', 'references', '{"section":"deployment security gates"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000006', 'd1000000-0000-0000-0000-00000000000b', 'references', '{"section":"PCI-DSS payment requirements"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000006', 'd1000000-0000-0000-0000-000000000022', 'links_to',   '{"reason":"security terminology"}')
ON CONFLICT DO NOTHING;

-- ═══════════════════════════════════════════════════════════════
-- API v2 Spec (07) — Hub #2: 10 outgoing edges
-- ═══════════════════════════════════════════════════════════════
INSERT INTO edges (tenant_id, from_doc_id, to_doc_id, relation, evidence) VALUES
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000007', 'd1000000-0000-0000-0000-000000000008', 'references', '{"section":"auth endpoints"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000007', 'd1000000-0000-0000-0000-000000000009', 'references', '{"section":"search API"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000007', 'd1000000-0000-0000-0000-00000000000a', 'references', '{"section":"webhook notifications"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000007', 'd1000000-0000-0000-0000-00000000000b', 'references', '{"section":"payment endpoints"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000007', 'd1000000-0000-0000-0000-00000000000e', 'links_to',   '{"reason":"latency spike caused API redesign"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000007', 'd1000000-0000-0000-0000-00000000001e', 'references', '{"section":"REST decision rationale"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000007', 'd1000000-0000-0000-0000-000000000013', 'links_to',   '{"reason":"scaling procedures for API"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000007', 'd1000000-0000-0000-0000-000000000004', 'references', '{"section":"API code review checklist"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000007', 'd1000000-0000-0000-0000-00000000000c', 'references', '{"section":"data ingestion endpoints"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000007', 'd1000000-0000-0000-0000-000000000022', 'links_to',   '{"reason":"API terminology definitions"}')
ON CONFLICT DO NOTHING;

-- ═══════════════════════════════════════════════════════════════
-- Auth System (08) — Hub #3: 8 outgoing edges
-- ═══════════════════════════════════════════════════════════════
INSERT INTO edges (tenant_id, from_doc_id, to_doc_id, relation, evidence) VALUES
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000008', 'd1000000-0000-0000-0000-000000000002', 'references', '{"section":"RBAC implementation"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000008', 'd1000000-0000-0000-0000-00000000000f', 'links_to',   '{"reason":"CVE in auth flow"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000008', 'd1000000-0000-0000-0000-000000000014', 'links_to',   '{"reason":"token secret rotation"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000008', 'd1000000-0000-0000-0000-00000000001c', 'references', '{"section":"why postgres for session store"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000008', 'd1000000-0000-0000-0000-000000000017', 'links_to',   '{"reason":"auth setup for new engineers"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000008', 'd1000000-0000-0000-0000-000000000006', 'references', '{"section":"security baseline compliance"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000008', 'd1000000-0000-0000-0000-00000000001d', 'references', '{"section":"event-driven token events"}'),
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000008', 'd1000000-0000-0000-0000-000000000019', 'links_to',   '{"reason":"auth monitoring alerts"}')
ON CONFLICT DO NOTHING;

-- ═══════════════════════════════════════════════════════════════
-- Incident → Runbook links (natural connections)
-- ═══════════════════════════════════════════════════════════════
INSERT INTO edges (tenant_id, from_doc_id, to_doc_id, relation, evidence) VALUES
-- DB outage → DB failover runbook
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000d', 'd1000000-0000-0000-0000-000000000012', 'links_to',   '{"reason":"failover used during incident"}'),
-- DB outage → Restore backup runbook
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000d', 'd1000000-0000-0000-0000-000000000016', 'links_to',   '{"reason":"backup restore attempted"}'),
-- DB outage → Monitoring guide
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000d', 'd1000000-0000-0000-0000-000000000019', 'links_to',   '{"reason":"alert gaps identified"}'),
-- DB outage → Incident response policy
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000d', 'd1000000-0000-0000-0000-000000000003', 'references', '{"section":"P1 escalation path"}'),
-- API latency → Scale API runbook
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000e', 'd1000000-0000-0000-0000-000000000013', 'links_to',   '{"reason":"horizontal scaling during incident"}'),
-- API latency → Monitoring guide
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000e', 'd1000000-0000-0000-0000-000000000019', 'links_to',   '{"reason":"latency thresholds review"}'),
-- Auth bypass → Rotate secrets runbook
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000f', 'd1000000-0000-0000-0000-000000000014', 'links_to',   '{"reason":"emergency secret rotation"}'),
-- Auth bypass → Access control policy
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000f', 'd1000000-0000-0000-0000-000000000002', 'links_to',   '{"reason":"policy gap identified"}'),
-- Auth bypass → Incident response policy
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000f', 'd1000000-0000-0000-0000-000000000003', 'references', '{"section":"security incident handling"}'),
-- Payment errors → Payment gateway spec
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000010', 'd1000000-0000-0000-0000-00000000000b', 'links_to',   '{"reason":"integration bug analysis"}'),
-- Payment errors → Deploy hotfix runbook
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000010', 'd1000000-0000-0000-0000-000000000015', 'links_to',   '{"reason":"hotfix deployed for fix"}'),
-- DNS failure → Kubernetes ADR
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000011', 'd1000000-0000-0000-0000-00000000001f', 'links_to',   '{"reason":"CoreDNS configuration issue"}'),
-- DNS failure → Monitoring guide
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000011', 'd1000000-0000-0000-0000-000000000019', 'links_to',   '{"reason":"DNS monitoring gap"}')
ON CONFLICT DO NOTHING;

-- ═══════════════════════════════════════════════════════════════
-- Spec → ADR references (design rationale)
-- ═══════════════════════════════════════════════════════════════
INSERT INTO edges (tenant_id, from_doc_id, to_doc_id, relation, evidence) VALUES
-- Search engine → Use PostgreSQL ADR
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000009', 'd1000000-0000-0000-0000-00000000001c', 'references', '{"section":"pg_search decision"}'),
-- Search engine → Event-driven ADR
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000009', 'd1000000-0000-0000-0000-00000000001d', 'references', '{"section":"async indexing pipeline"}'),
-- Data pipeline → Event-driven ADR
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000c', 'd1000000-0000-0000-0000-00000000001d', 'references', '{"section":"event sourcing for data flow"}'),
-- Data pipeline → Use PostgreSQL ADR
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000c', 'd1000000-0000-0000-0000-00000000001c', 'references', '{"section":"warehouse storage"}'),
-- Notification spec → Event-driven ADR
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000a', 'd1000000-0000-0000-0000-00000000001d', 'references', '{"section":"event bus for notifications"}'),
-- Payment gateway → Use PostgreSQL ADR
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000b', 'd1000000-0000-0000-0000-00000000001c', 'references', '{"section":"transaction storage"}')
ON CONFLICT DO NOTHING;

-- ═══════════════════════════════════════════════════════════════
-- Guide cross-references
-- ═══════════════════════════════════════════════════════════════
INSERT INTO edges (tenant_id, from_doc_id, to_doc_id, relation, evidence) VALUES
-- Onboarding → Dev environment setup
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000017', 'd1000000-0000-0000-0000-000000000018', 'links_to',   '{"reason":"step 2 of onboarding"}'),
-- Onboarding → Code review standards
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000017', 'd1000000-0000-0000-0000-000000000004', 'links_to',   '{"reason":"PR process introduction"}'),
-- Onboarding → CI/CD guide
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000017', 'd1000000-0000-0000-0000-00000000001b', 'links_to',   '{"reason":"deploy pipeline overview"}'),
-- Onboarding → Testing strategy
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000017', 'd1000000-0000-0000-0000-00000000001a', 'links_to',   '{"reason":"testing expectations"}'),
-- Onboarding → Glossary
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000017', 'd1000000-0000-0000-0000-000000000022', 'links_to',   '{"reason":"company terminology"}'),
-- Dev environment → Kubernetes ADR
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000018', 'd1000000-0000-0000-0000-00000000001f', 'references', '{"section":"local k8s setup"}'),
-- CI/CD guide → Deployment policy
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000001b', 'd1000000-0000-0000-0000-000000000005', 'references', '{"section":"deploy gates compliance"}'),
-- CI/CD guide → Testing strategy
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000001b', 'd1000000-0000-0000-0000-00000000001a', 'links_to',   '{"reason":"test stage in pipeline"}'),
-- Testing strategy → Code review standards
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000001a', 'd1000000-0000-0000-0000-000000000004', 'links_to',   '{"reason":"test coverage in reviews"}'),
-- Monitoring guide → Incident response policy
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000019', 'd1000000-0000-0000-0000-000000000003', 'references', '{"section":"alert escalation paths"}'),
-- Monitoring guide → Data pipeline
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000019', 'd1000000-0000-0000-0000-00000000000c', 'links_to',   '{"reason":"pipeline health metrics"}')
ON CONFLICT DO NOTHING;

-- ═══════════════════════════════════════════════════════════════
-- Policy cross-references
-- ═══════════════════════════════════════════════════════════════
INSERT INTO edges (tenant_id, from_doc_id, to_doc_id, relation, evidence) VALUES
-- Code review → Testing strategy
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000004', 'd1000000-0000-0000-0000-00000000001a', 'links_to',   '{"reason":"test requirements in reviews"}'),
-- Deployment policy → CI/CD guide
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000005', 'd1000000-0000-0000-0000-00000000001b', 'links_to',   '{"reason":"implementation guide"}'),
-- Deployment policy → Hotfix runbook
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000005', 'd1000000-0000-0000-0000-000000000015', 'links_to',   '{"reason":"emergency deploy procedure"}'),
-- Data retention → Data pipeline
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000000c', 'references', '{"section":"data lifecycle in pipeline"}'),
-- Access control → Auth system
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000002', 'd1000000-0000-0000-0000-000000000008', 'links_to',   '{"reason":"implementation spec"}'),
-- Incident response → Monitoring guide
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000003', 'd1000000-0000-0000-0000-000000000019', 'links_to',   '{"reason":"monitoring as first response"}')
ON CONFLICT DO NOTHING;

-- ═══════════════════════════════════════════════════════════════
-- Runbook cross-references
-- ═══════════════════════════════════════════════════════════════
INSERT INTO edges (tenant_id, from_doc_id, to_doc_id, relation, evidence) VALUES
-- DB failover → Use PostgreSQL ADR
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000012', 'd1000000-0000-0000-0000-00000000001c', 'references', '{"section":"replication architecture"}'),
-- DB failover → Restore backup
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000012', 'd1000000-0000-0000-0000-000000000016', 'links_to',   '{"reason":"fallback if failover fails"}'),
-- Scale API → Kubernetes ADR
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000013', 'd1000000-0000-0000-0000-00000000001f', 'references', '{"section":"HPA configuration"}'),
-- Scale API → API v2 spec
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000013', 'd1000000-0000-0000-0000-000000000007', 'references', '{"section":"rate limit thresholds"}'),
-- Rotate secrets → Security baseline
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000014', 'd1000000-0000-0000-0000-000000000006', 'references', '{"section":"rotation schedule requirements"}'),
-- Deploy hotfix → Deployment policy
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000015', 'd1000000-0000-0000-0000-000000000005', 'references', '{"section":"emergency deploy approval"}'),
-- Deploy hotfix → Code review standards
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000015', 'd1000000-0000-0000-0000-000000000004', 'references', '{"section":"expedited review process"}'),
-- Restore backup → Data retention
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000016', 'd1000000-0000-0000-0000-000000000001', 'references', '{"section":"backup retention window"}')
ON CONFLICT DO NOTHING;

-- ═══════════════════════════════════════════════════════════════
-- Report/Quarterly links
-- ═══════════════════════════════════════════════════════════════
INSERT INTO edges (tenant_id, from_doc_id, to_doc_id, relation, evidence) VALUES
-- Q1 review → API v2 spec
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000020', 'd1000000-0000-0000-0000-000000000007', 'references', '{"section":"API v2 launch metrics"}'),
-- Q1 review → DB outage incident
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000020', 'd1000000-0000-0000-0000-00000000000d', 'references', '{"section":"incident impact analysis"}'),
-- Q1 review → Search engine spec
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000020', 'd1000000-0000-0000-0000-000000000009', 'references', '{"section":"search feature rollout"}'),
-- Security audit → Security baseline
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000021', 'd1000000-0000-0000-0000-000000000006', 'references', '{"section":"compliance status"}'),
-- Security audit → Auth bypass incident
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000021', 'd1000000-0000-0000-0000-00000000000f', 'references', '{"section":"CVE timeline"}'),
-- Security audit → Access control policy
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000021', 'd1000000-0000-0000-0000-000000000002', 'references', '{"section":"access control gaps"}'),
-- Security audit → Rotate secrets runbook
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000021', 'd1000000-0000-0000-0000-000000000014', 'references', '{"section":"rotation compliance"}')
ON CONFLICT DO NOTHING;

-- ═══════════════════════════════════════════════════════════════
-- ADR cross-references
-- ═══════════════════════════════════════════════════════════════
INSERT INTO edges (tenant_id, from_doc_id, to_doc_id, relation, evidence) VALUES
-- Kubernetes ADR → Deployment policy
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000001f', 'd1000000-0000-0000-0000-000000000005', 'references', '{"section":"container deployment model"}'),
-- Event-driven ADR → Data pipeline spec
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000001d', 'd1000000-0000-0000-0000-00000000000c', 'links_to',   '{"reason":"event bus implementation"}'),
-- GraphQL vs REST ADR → API v2 spec
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000001e', 'd1000000-0000-0000-0000-000000000007', 'links_to',   '{"reason":"REST chosen per ADR"}'),
-- Use PostgreSQL ADR → Data retention policy
('00000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-00000000001c', 'd1000000-0000-0000-0000-000000000001', 'references', '{"section":"storage lifecycle"}')
ON CONFLICT DO NOTHING;
