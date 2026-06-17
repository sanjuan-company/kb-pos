# Code Changes for POST /migrate

---

## `main.go`

### 1. Add import

Insert `_ "embed"` into the import block (line 3-13):

```go
import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)
```

### 2. Add embed directive

Insert after the import block (before `type App struct`):

```go
//go:embed db/init.sql
var initSQL string
```

### 3. Add migration handler

Insert after `envOrDefault` function (before closing `}` of file):

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
	c.JSON(http.StatusOK, ApiResponse{Success: true, Message: "migration completed successfully"})
}
```

### 4. Register route

Add inside `main()`, after `r.GET("/health", healthHandler)` (line 55):

```go
	r.POST("/migrate", app.migrateHandler)
```

---

## `docker-compose.yml`

### Remove volume mount

On line 15, remove `- ./db/init.sql:/docker-entrypoint-initdb.d/init.sql`:

```yaml
    volumes:
      - pgdata:/var/lib/postgresql/data
      # delete this line:
      # - ./db/init.sql:/docker-entrypoint-initdb.d/init.sql
```

---

## `Dockerfile`

**No changes.**
