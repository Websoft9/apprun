package config

// validTestConfigYAML returns a complete valid configuration for testing
const validTestConfigYAML = `app:
  name: "test-service"
  version: "1.0.0"
  timezone: "Asia/Shanghai"
database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "testuser"
  password: "testpassword123"
  db_name: "testdb"
cache:
  host: "localhost"
  port: "6379"
  password: ""
  db: 0
  pool_size: 10
  timeout: 2s
logger:
  level: "info"
  output:
    targets: ["stdout"]
i18n:
  default_language: "en-US"
  supported_languages: ["en-US", "zh-CN"]
  translations_path: "./testdata"
auth:
  jwt:
    secret: "test-secret-must-be-at-least-32-characters-long!"
    access_token_expiration: 24h
    refresh_token_expiration: 168h
    issuer: "test-platform"
    audience: "test-api"
    whitelist_paths:
      - /health
  security:
    bcrypt_cost: 10
    failed_login_cache_ttl: 5m
    max_failed_attempts: 5
`
