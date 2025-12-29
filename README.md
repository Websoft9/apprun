# apprun

A lightweight BaaS (Backend as a Service) framework for developers, SDD (spec-driven-development), built with Go following the [BMad Method](https://github.com/bmad-code-org/BMAD-METHOD).

---

## 🚀 What is apprun?

**apprun** is a modular, cloud-neutral BaaS platform that provides:

- **Authentication & Authorization**: Built-in user management with Ory Kratos integration
- **Data Management**: RESTful APIs with PostgreSQL and Ent ORM
- **Storage Service**: File storage with pluggable backends (Local/S3)
- **Workflow Engine**: Flexible business process automation
- **Real-time Features**: WebSocket support for live updates
- **Multi-tenant**: Project-based resource isolation

**Key Features**:
- 🔒 Security-first design with RBAC
- 🌍 i18n/l10n support (English, Chinese, Japanese)
- 🔌 Plugin architecture for extensibility
- ☁️ Cloud-neutral deployment
- 📦 Production-ready with monitoring & logging

---

## 📦 Deployment

### Prerequisites
- Go 1.21+
- PostgreSQL 14+
- Redis 7+ (optional, for caching)

### Quick Start

```bash
# Clone repository
git clone https://github.com/Websoft9/apprun.git
cd apprun/core

# Configure environment
cp config/default.yaml config/local.yaml
# Edit config/local.yaml with your settings

# Run server
make run

# Access API documentation
# Swagger UI: http://localhost:8080/api/docs/index.html
# OpenAPI spec: http://localhost:8080/api/docs/doc.json
```

### API Documentation

**Interactive API documentation is available via Swagger UI**:

- **URL**: `http://localhost:8080/api/docs/index.html`
- **Features**:
  - Browse all API endpoints with descriptions
  - Test APIs directly in the browser ("Try it out")
  - View request/response examples
  - See data models and validation rules

**Generating documentation**:
```bash
# After modifying API code, regenerate Swagger docs
make swagger
```

### Production Deployment
- Docker: TBD
- Kubernetes: TBD
- Cloud Providers: TBD

---

## 🤝 Contributing

We follow the **BMad Method** for development:

1. **Read Documentation**: Check [`docs/`](./docs/) for project standards
2. **Find Issues**: Look for issues tagged `good-first-issue`
3. **Follow Standards**: Read [`docs/standards/`](./docs/standards/) before coding
4. **Create PR**: Follow the [DevOps Process](./docs/standards/devops-process.md)

See [CONTRIBUTING.md](./CONTRIBUTING.md) for detailed guidelines.

---

## 📚 Documentation

- [Product Requirements](./docs/prd.md)
- [Architecture](./docs/architecture/)
- [API Standards](./docs/standards/api-design.md)
- [Sprint Artifacts](./docs/sprint-artifacts/)

---

## 📄 License

[MIT License](./LICENSE)

---

**Maintained by**: [Websoft9](https://www.websoft9.com)