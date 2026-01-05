# Atlas Configuration for AppRun
# See: https://atlasgo.io/atlas-schema/projects

# Environment configuration
env "local" {
  # Database connection URL (override via ATLAS_URL env var)
  url = getenv("ATLAS_URL", "postgres://postgres:password@localhost:5432/apprun?sslmode=disable")
  
  # Dev database for computing diffs (uses temp container)
  dev = "docker://postgres/15/dev?search_path=public"
  
  # Migrations directory
  migration {
    dir = "file://migrations"
  }
  
  # Schema source (Ent schema)
  src = "ent://ent/schema"
}

env "production" {
  url = getenv("ATLAS_URL")
  
  migration {
    dir = "file://migrations"
  }
  
  src = "ent://ent/schema"
}

# Diff policy - prevent destructive changes
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
