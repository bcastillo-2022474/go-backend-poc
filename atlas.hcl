env "local" {
  src = [
    "file://infra/user/sql/schema.sql",
    "file://infra/auth/sql/schema.sql"
  ]
  url = "postgres://postgres:postgres@localhost:5437/edoo_class?sslmode=disable"
  dev = "docker://postgres/15/dev"
  migration {
    dir = "file://migrations"
  }
}

env "dev" {
  src = [
    "file://infra/user/sql/schema.sql",
    "file://infra/auth/sql/schema.sql"
  ]
  url = env("DATABASE_URL")
  dev = "docker://postgres/15/dev"
  migration {
    dir = "file://migrations"
  }
}