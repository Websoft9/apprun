# apprun

A lightweight BaaS (Backend as a Service) framework built with Go following the [BMad Method](https://github.com/bmad-code-org/BMAD-METHOD).

---

## 🚀 What is apprun?

**apprun** is a modular, cloud-neutral BaaS platform that provides:

- **Authentication & Authorization**: User management and RBAC
- **Data Management**: RESTful APIs with PostgreSQL and Ent ORM
- **Storage Service**: File storage with pluggable backends
- **Workflow Engine**: Business process automation
- **Real-time Features**: WebSocket support
- **Multi-tenant**: Project-based resource isolation

**Key Features**:
- 🔒 Security-first design
- 🌍 Internationalization support
- 🔌 Plugin architecture
- ☁️ Cloud-neutral deployment
- 📦 Production-ready

---

## 📦 Deployment

### Prerequisites
- Docker 20.10+
- Docker Compose 2.0+

### Quick Start

```bash
# Clone repository
git clone https://github.com/Websoft9/apprun.git
cd apprun

# Start services
docker-compose up -d

# Access application
# API: http://localhost:8080
# Swagger UI: http://localhost:8080/api/docs/index.html
```

### Production Deployment

```bash
# Use production configuration
docker-compose -f docker-compose.yml up -d

# Check service health
curl http://localhost:8080/health
```

**Configuration**:
- Environment variables: See `docker-compose.yml`
- Secrets: Use Docker secrets or external secret manager
- Database: PostgreSQL 14+ (included in docker-compose)

---

## 🤝 Contributing

We follow the **BMad Method** for AI-assisted development.

**Get started**:
1. Read [CONTRIBUTING.md](./CONTRIBUTING.md)
2. Check [Sprint Artifacts](./docs/sprint-artifacts/)
3. Follow [Standards](./docs/standards/)
4. Submit Pull Request

**For maintainers**: See [OWNER.md](./OWNER.md)

---

## 📚 Documentation

- [Product Requirements](./docs/prd.md)
- [Architecture](./docs/architecture/)
- [Technical Standards](./docs/standards/)
- [Sprint Artifacts](./docs/sprint-artifacts/)

---

## 📄 License

[MIT License](./LICENSE)

---

**Maintained by**: [Websoft9](https://www.websoft9.com)