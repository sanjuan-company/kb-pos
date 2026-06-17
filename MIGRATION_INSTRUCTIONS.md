# Migration via HTTP Endpoint

This document describes how to move the DB migration from `docker-compose.yml` (host-mounted `init.sql`) into the Go app as an HTTP endpoint, so other devs don't need `db/init.sql` locally.

---

## 1. Changes to `main.go`

**Add the `embed` import and embed the SQL file:**

```go
import (
    _ "embed"
    // ... existing imports
)

//go:embed db/init.sql
var initSQL string
```

> The `//go:embed` directive bakes `db/init.sql` into the binary at compile time. The file `db/init.sql` must exist relative to `main.go` during `go build`.

**Add a migration handler:**

```go
func (a *App) migrateHandler(c *gin.Context) {
    _, err := a.db.Exec(initSQL)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ApiResponse{
            Success: false,
            Message: "migration failed: " + err.Error(),
        })
        return
    }
    c.JSON(http.StatusOK, ApiResponse{
        Success: true,
        Message: "migration completed successfully",
    })
}
```

**Register the route in `main()`:**

```go
r.POST("/migrate", app.migrateHandler)
```

---

## 2. Changes to `docker-compose.yml`

**Remove the `init.sql` volume mount** from the `postgres` service:

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: kbpos-postgres
    environment:
      POSTGRES_USER: myuser
      POSTGRES_PASSWORD: "8013075"
      POSTGRES_DB: postgres
    ports:
      - "9041:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
      # REMOVE this line:
      # - ./db/init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U myuser -d postgres"]
      interval: 5s
      timeout: 5s
      retries: 5
```

---

## 3. Changes to `Dockerfile`

**No changes needed.** The existing `COPY . .` already copies `db/init.sql` into the builder stage, and Go's `//go:embed` includes it in the binary. The final image doesn't need the file at runtime.

---

## 4. Usage

After `docker-compose up`, run the migration once:

```bash
curl -X POST http://localhost:9040/migrate
```

Response on success:
```json
{"success": true, "message": "migration completed successfully"}
```

The `init.sql` is already idempotent (`CREATE TABLE IF NOT EXISTS`, `CREATE OR REPLACE FUNCTION`), so calling `/migrate` multiple times is safe.
