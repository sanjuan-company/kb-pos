# Docker Build & Deployment Instructions

## Overview

This setup packages the Go CRUD app into a Docker image together with PostgreSQL.  
When you run the container, it automatically creates the database tables and stored procedures.

---

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed on your machine
- A [Docker Hub](https://hub.docker.com/) account (or any container registry)

---

## 1. File Structure

```
project-root/
├── main.go
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
├── db/
│   └── init.sql          # Tables + functions (auto-created on first run)
└── docker-build-instruction.md
```

---

## 2. Build the Docker Image

```bash
# Build the Go app image
docker build -t kbpos-17-app:latest .
```

Tag and push to your registry (example with Docker Hub):

```bash
docker tag kbpos-17-app:latest yourdockerhub/kbpos-17-app:latest
docker push yourdockerhub/kbpos-17-app:latest
```

---

## 3. Run with Docker Compose (Recommended)

```bash
docker-compose up -d
```

This starts:
- **PostgreSQL** on port `9041` (configurable)
- **Go API** on port `9040`

The first time PostgreSQL starts, it runs `db/init.sql` to create all tables, functions, and constraints automatically.

---

## 4. Run Manually (without Compose)

### 4.1 Create a network

```bash
docker network create kbpos-network
```

### 4.2 Start PostgreSQL

```bash
docker run -d \
  --name kbpos-postgres \
  --network kbpos-network \
  -e POSTGRES_USER=myuser \
  -e POSTGRES_PASSWORD=8013075 \
  -e POSTGRES_DB=postgres \
  -v $(pwd)/db/init.sql:/docker-entrypoint-initdb.d/init.sql \
  -p 9041:5432 \
  postgres:16-alpine
```

### 4.3 Run the App

```bash
docker run -d \
  --name kbpos-app \
  --network kbpos-network \
  -e POSTGRES_HOST=kbpos-postgres \
  -e POSTGRES_PORT=5432 \
  -e POSTGRES_USER=myuser \
  -e POSTGRES_PASSWORD=8013075 \
  -e POSTGRES_DB=postgres \
  -e APP_PORT=9040 \
  -p 9040:9040 \
  kbpos-17-app:latest
```

---

## 5. Pull & Run on a Different Machine

```bash
# Pull the image
docker pull yourdockerhub/kbpos-17-app:latest

# Create a local directory for the init script
mkdir -p db

# Save the init.sql from the FREELANCE(POSTGRE).txt content locally as db/init.sql,
# then run docker-compose (or the manual steps above)
docker-compose up -d
```

Or pull everything directly:

```bash
docker pull yourdockerhub/kbpos-17-app:latest
docker pull postgres:16-alpine
```

---

## 6. Environment Variables

| Variable          | Default       | Description                  |
|-------------------|---------------|------------------------------|
| `POSTGRES_USER`   | `myuser`      | PostgreSQL user              |
| `POSTGRES_PASSWORD` | `8013075`   | PostgreSQL password          |
| `POSTGRES_DB`     | `postgres`    | Database name                |
| `POSTGRES_HOST`   | `localhost`   | PostgreSQL hostname          |
| `POSTGRES_PORT`   | `5432`        | PostgreSQL internal port     |
| `APP_PORT`        | `9040`        | Go API server port           |

---

## 7. API Endpoints

| Method | Endpoint             | Description       |
|--------|----------------------|-------------------|
| POST   | `/addusers`          | Add a new user    |
| PUT    | `/updateuser/:id`    | Update user by ID |
| DELETE | `/deleteuser/:id`    | Delete user by ID |
| GET    | `/health`            | Health check      |

---

## 8. Verify Setup

```bash
# Health check
curl http://localhost:9040/health

# Add a user
curl -X POST http://localhost:9040/addusers \
  -H "Content-Type: application/json" \
  -d '{
    "fullname": "Juan Dela Cruz",
    "fname": "Juan",
    "lname": "Cruz",
    "mname": "Dela",
    "staffid": "STF12345",
    "contactnumber": "+639171234567",
    "email": "juan.cruz@example.com",
    "telephonenumber": "0287654321",
    "mainaddress": "123 Mabini Street, Manila",
    "secondaryaddress": "Unit 5, Sunshine Apartments",
    "lastaddress": "Old Residence, Quezon City",
    "birthdate": "1990-05-15T00:00:00Z",
    "userroleid": 1,
    "userpassword": "SecurePass123!"
  }'
```

---

## 9. Stop Containers

```bash
docker-compose down
# or for manual setup:
docker stop kbpos-app kbpos-postgres
docker rm kbpos-app kbpos-postgres
```
