# Atlas Configuration for AppRun
# See: https://atlasgo.io/atlas-schema/projects

# Environment configuration
env "local" {
  # Database connection URL (override via ATLAS_URL env var)
  url = "postgres://apprun:dev_password_123@localhost:5432/apprun_dev?sslmode=disable"
  
  # Dev database for computing diffs (uses temp container)
  dev = "docker://postgres/15/dev?search_path=public"
  
  # Migrations directory
  migration {
    dir = "file://migrations"
  }
  
  # Schema source (Ent schema)
  src = "ent://ent/schema"
  
  # Exclude Atlas metadata tables from inspection
  # This prevents diff/inspect from reporting atlas_schema_revisions as drift
  schemas = ["public"]
  exclude = [
    # Pattern to exclude Atlas migration tracking table
    # Note: This table is managed by Atlas SDK, not by application schema
    "atlas_schema_revisions",
  ]
}

env "production" {
  url = getenv("DATABASE_URL")
  
  migration {
    dir = "file://migrations"
  }
  
  src = "ent://ent/schema"
  
  schemas = ["public"]
  exclude = [
    "atlas_schema_revisions",
  ]
}

# Diff policy - prevent destructive changes (global)
diff {
  # Skip dropping columns (data safety)
  skip {
    drop_column = true
    drop_table  = true
  }
}

# Lint policy
lint {
  # Destructive changes check
  destructive {
    error = true
  }
  
  # Data-dependent changes
  data_depend {
    error = true
  }
}
