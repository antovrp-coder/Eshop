# StyleHome E-Commerce Platform

> Production-ready, enterprise-scale e-commerce platform built with Go (Standard Library `net/http` + gRPC), CQRS, PostgreSQL (`Eshop`), Elasticsearch 8, Redis 7, Apache Kafka, and React 18 (TypeScript + Vite).

---

## 🏛️ Architecture Overview

- **Backend**: Go 1.26 without external web frameworks (`net/http` + custom middleware pipeline)
- **Database**: PostgreSQL (`Eshop`) with isolated domain schemas
- **Inter-Service Communication**: gRPC (Protobuf v3) over HTTP/2
- **Client API**: GraphQL Backend-for-Frontend (BFF)
- **Search & Discovery**: Elasticsearch 8 (BM25 + Edge N-Gram Autocomplete + Facets)
- **Event Bus**: Apache Kafka with Transactional Outbox Pattern
- **Cache & Rate Limiting**: Redis 7
- **Frontend**: React 18 SPA (TypeScript + React Query + React Router)

---

## 🚀 Quick Start

### 1. Run Backend Server (Go `net/http`)
```bash
go run ./cmd/bff-server
```
- GraphQL BFF: `http://localhost:8080/graphql`
- Health Probe: `http://localhost:8080/healthz`

### 2. Run Frontend SPA (Vite)
```bash
cd frontend
npm install
npm run dev
```
- Frontend Web App: `http://localhost:3000`

### 3. Run Full Stack with Docker Compose
```bash
docker-compose up --build
```

---

## 📖 Complete Documentation
For full architectural details, technology trade-offs, and commands, see:
- [`STYLEHOME_SYSTEM_DOCUMENTATION_AND_COMMANDS.txt`](file:///c:/Users/Rehna/Downloads/Go-lab/Eshop/STYLEHOME_SYSTEM_DOCUMENTATION_AND_COMMANDS.txt)
- [`docs/ARCHITECTURE_BLUEPRINT.md`](file:///c:/Users/Rehna/Downloads/Go-lab/Eshop/docs/ARCHITECTURE_BLUEPRINT.md)
