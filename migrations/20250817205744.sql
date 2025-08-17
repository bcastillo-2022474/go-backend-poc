-- Create "users" table
CREATE TABLE "public"."users" (
  "id" uuid NOT NULL,
  "name" character varying(255) NOT NULL,
  "email" character varying(255) NOT NULL,
  "password_hash" character varying(255) NOT NULL,
  "created_at" timestamptz NULL DEFAULT now(),
  "updated_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_key" UNIQUE ("email")
);
-- Create index "idx_users_created_at" to table: "users"
CREATE INDEX "idx_users_created_at" ON "public"."users" ("created_at");
-- Create index "idx_users_email" to table: "users"
CREATE INDEX "idx_users_email" ON "public"."users" ("email");
