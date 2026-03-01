-- 000010_seed.up.sql
-- Bootstrap seed data: default tenant for development

INSERT INTO tenants (tenant_id, name) VALUES
    ('00000000-0000-0000-0000-000000000001', 'default')
ON CONFLICT (tenant_id) DO NOTHING;
