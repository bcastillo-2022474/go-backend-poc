-- Casbin rule table for storing authorization policies and role assignments
-- 
-- IMPORTANT: This table supports both policy rules ('p') and grouping rules ('g'),
-- but our current implementation ONLY uses it for role assignments ('g' records).
-- Policy rules ('p' records) are intentionally managed in memory from YAML configuration
-- to maintain version control and consistency across tenants.
--
-- Future consideration: If needed, this table can store policy rules for tenant-specific
-- customizations while maintaining the current YAML-based approach as the default.
CREATE TABLE casbin_rule (
    id SERIAL PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL,  -- Policy type: 'p'/'p2'/etc for policies, 'g'/'g2'/etc for groupings
    v0 VARCHAR(100),              -- Subject (user_id for 'g', role for 'p')
    v1 VARCHAR(100),              -- Object (role for 'g', resource for 'p') 
    v2 VARCHAR(100),              -- Domain (tenant_id for both 'g' and 'p', action for 'p')
    v3 VARCHAR(100),              -- Reserved (tenant_id for 'p' policies)
    v4 VARCHAR(100),              -- Reserved for future extensions
    v5 VARCHAR(100),              -- Reserved for future extensions
    CONSTRAINT casbin_rule_unique UNIQUE (ptype, v0, v1, v2, v3)
);

-- Indexes for efficient authorization queries
CREATE INDEX idx_casbin_rule_ptype ON casbin_rule(ptype);
CREATE INDEX idx_casbin_rule_v0_v1_v2 ON casbin_rule(v0, v1, v2);      -- Common lookup pattern
CREATE INDEX idx_casbin_rule_v1_v2 ON casbin_rule(v1, v2);            -- Role/resource + tenant
CREATE INDEX idx_casbin_rule_ptype_v0 ON casbin_rule(ptype, v0);       -- Type + subject/role