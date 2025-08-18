-- Create "casbin_rule" table
CREATE TABLE "public"."casbin_rule" (
  "id" serial NOT NULL,
  "ptype" character varying(100) NOT NULL,
  "v0" character varying(100) NULL,
  "v1" character varying(100) NULL,
  "v2" character varying(100) NULL,
  "v3" character varying(100) NULL,
  "v4" character varying(100) NULL,
  "v5" character varying(100) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "casbin_rule_unique" UNIQUE ("ptype", "v0", "v1", "v2", "v3")
);
-- Create index "idx_casbin_rule_ptype" to table: "casbin_rule"
CREATE INDEX "idx_casbin_rule_ptype" ON "public"."casbin_rule" ("ptype");
-- Create index "idx_casbin_rule_ptype_v0" to table: "casbin_rule"
CREATE INDEX "idx_casbin_rule_ptype_v0" ON "public"."casbin_rule" ("ptype", "v0");
-- Create index "idx_casbin_rule_v0_v1_v2" to table: "casbin_rule"
CREATE INDEX "idx_casbin_rule_v0_v1_v2" ON "public"."casbin_rule" ("v0", "v1", "v2");
-- Create index "idx_casbin_rule_v1_v2" to table: "casbin_rule"
CREATE INDEX "idx_casbin_rule_v1_v2" ON "public"."casbin_rule" ("v1", "v2");
