# How the Backend Works: Layers & Request Flow

This guide explains what happens when an API is called, which part of our code
runs first, and what each layer is for. Every example points to real code in this repo.

Example endpoint used throughout: **`POST /admin/driver/{id}/payroll`**

---

## Part 1: Startup (runs once, when the server starts)

Nothing here handles a request yet. It **builds** everything that requests will use.

```
cmd/main.go  main()
 │
 ├─ 1. read env vars           DM_DB_DSN, SECRET_KEY, PORT, CLOUDINARY_*   (panics if missing)
 ├─ 2. connect to Postgres     pgxpool.New(...)  → "pool" = reusable DB connections
 ├─ 3. queries := dm_svc_db.New(pool)        ← sqlc-generated code, wraps the pool
 │
 ├─ 4. wiring.NewDependencyContainer(config)        internal/wiring/di.go
 │       builds the layers bottom-up:
 │         driverSvc     := driver.NewDriverService(logger, queries, cfg)   ← service gets the DB
 │         DriverHandler := driver.NewDriverHandler(logger, driverSvc)      ← handler gets the service
 │
 ├─ 5. routes := transport.NewRoutes(di)            internal/transport/routes.go
 │       builds a "map" of URL → function:
 │         POST /admin/driver/{id}/payroll  →  di.DriverHandler.AddPayroll
 │
 └─ 6. http.Server{Handler: routes}.ListenAndServe()
         ← now it waits for requests forever
```

**Dependency injection (step 4):** each layer receives what it needs from outside.
The handler is *given* a service, and the service is *given* the DB.
That's why the handler can call `h.service.AddPayrollService(...)` without creating the service itself.

---

## Part 2: A request arrives

The frontend sends:

```
POST /admin/driver/5/payroll
Cookie: accessToken=eyJhbGc...
{"entry_type":"SALARY_CREDIT","amount":3000,"notes":"Nov"}
```

The request goes **down** through the layers, then the response comes back **up**:

```
 ┌──────────────────────────────────────────────────────────────────┐
 │ ① Go's http.Server     accepts the connection, makes r + w       │
 │ ② chi router           matches URL → finds the route             │
 │ ③ CORS middleware      routes.go                                 │
 │ ④ AuthMiddleware       internal/auth/middleware.go               │
 │ ⑤ Handler              internal/driver/handler.go  AddPayroll    │
 │ ⑥ Service              internal/driver/service.go  AddPayroll... │
 │ ⑦ sqlc queries         internal/data/dm_svc_db/   DriverLedger() │
 │ ⑧ Postgres             INSERT INTO driver_ledger ...             │
 └──────────────────────────────────────────────────────────────────┘
          response travels back up and is written out
```

Go creates two objects that are passed through every layer:

- **`r *http.Request`**: everything about the incoming request (URL, headers, cookies, body)
- **`w http.ResponseWriter`**: where you write the response

That's why every handler looks like `func(w http.ResponseWriter, r *http.Request)`.

---

## Part 3: The layers

### Overview

| # | Layer | File in the repo | Its one job | Knows about | Must NOT know about |
|---|---|---|---|---|---|
| 1 | **Router** | `internal/transport/routes.go` | Send each URL to the right function | URLs, HTTP methods, which middleware guards which route | What the endpoint does inside |
| 2 | **Middleware** | `internal/auth/middleware.go` | Guard the door: let the request through or stop it | Cookies, JWT tokens, user roles | Which endpoint it's protecting, the database |
| 3 | **Handler** | `internal/driver/handler.go` | Translate HTTP into Go and Go back into HTTP | URL params, JSON body, status codes (200/400/500) | SQL, business rules |
| 4 | **Service** | `internal/driver/service.go` | Apply business rules | What "adding payroll" *means* (ledger + transaction) | HTTP, JSON, cookies, status codes |
| 5 | **Queries (sqlc)** | `internal/data/postgres/query.sql` → `dm_svc_db/` | Run SQL safely from Go | SQL, tables, columns | Why the data is needed |
| 6 | **Database** | `migrations/*.sql` (structure) + Postgres | Store data and enforce hard rules | Column types, foreign keys, NOT NULL | Everything above it |

---

### 1. Router: the receptionist

**File:** `internal/transport/routes.go`

It reads the address on the request and sends it to the right place. It doesn't open the request.

```go
r.Post("/{id}/payroll", di.DriverHandler.AddPayroll)
```

"If someone POSTs to `/admin/driver/{id}/payroll`, send it to `AddPayroll`."
chi also remembers `id = "5"`, which the handler later reads with `chi.URLParam(r, "id")`.

- **Used for:** defining URLs, grouping routes (`/admin/...`), attaching middleware to groups
  (`adminRoute.Use(auth.AuthMiddleware)`).
- **Go here when:** you add a new endpoint, or want to protect a group of routes with a role.

---

### 2. Middleware: the security guard

**File:** `internal/auth/middleware.go`

It runs **before** the handler on every route it's attached to, and does one of two things:

- **stops** the request: writes an error (401/403) and `return`s
- **passes** it on: `next.ServeHTTP(w, r)`

```go
AuthMiddleware   → "do you have a valid login token?"   no → 401 Unauthorized
RequireRole(...) → "is your role allowed here?"         no → 403 Forbidden
```

It also puts the user's ID and role into the request **context**, so later code can ask
who is calling (`auth.GetUserIDFromContext`).

CORS (in `routes.go`) is also a middleware: it checks the browser's `Origin` header against
`AllowedOrigins`. It only matters for browsers; Postman and curl ignore it.

- **Used for:** checks that apply to **many endpoints at once**: login, roles, CORS, logging.
- **Go here when:** working on "manager can't see finance". That's a `RequireRole` on the
  finance routes, not code inside each handler.

> ⚠️ Currently the payroll routes have **no `RequireRole`**, so any logged-in user can add payroll.

---

### 3. Handler: the translator

**File:** `internal/driver/handler.go`

It turns the raw HTTP request into clean Go values, calls the service, and turns the result
back into an HTTP response.

```go
driverID := chi.URLParam(r, "id")           // HTTP → Go   (read URL)
json.Decode(r.Body, &req)                    // HTTP → Go   (read body into req)
// validate input here                       // reject bad input → 400
err := h.service.AddPayrollService(...)      // hand over to business logic
responses.SuccessJSONResponse(..., 200, w)   // Go → HTTP   (write response)
```

`req` is just a variable holding the decoded JSON body. `req.EntryType` means
"the `entry_type` field the client sent".

- **Used for:** reading input, **validating its shape** (empty? too long? negative?),
  picking the right status code.
- **Go here when:** adding a request field, adding input validation, or changing what JSON goes back.
- **Warning sign:** SQL or business calculations in a handler are in the wrong layer.

---

### 4. Service: the brain

**File:** `internal/driver/service.go`

It knows what the business **means**. "Adding payroll" isn't one action. It's two:

```go
s.queries.DriverLedger(...)        // 1. record it on the driver's ledger
s.queries.CreateTransaction(...)   // 2. record it as a company expense
```

A handler couldn't know that; a service does.

- **Used for:** rules, calculations, combining several queries, and decisions like
  "a trip can't be closed if it's unpaid".
- **Go here when:** implementing **tariff × weight = trip income**. That's a business
  calculation, so it lives here.
- **Why no HTTP here:** a service takes plain Go values (`driverID int32`), so it can be called
  from a background job, a CLI tool or a test, not only from an HTTP request.

---

### 5. Queries (sqlc): the SQL messenger

**Files:** you write `internal/data/postgres/query.sql`; `sqlc generate` writes `internal/data/dm_svc_db/*.go`

```sql
-- name: DriverLedger :one
INSERT INTO driver_ledger (driver_id, entry_type, amount, related_trip_id, notes)
VALUES ($1, $2, $3, $4, $5) RETURNING *;
```

becomes a Go function: `queries.DriverLedger(ctx, DriverLedgerParams{...})`.

- **Used for:** type-safe database access. If you pass the wrong type, it won't compile, and the
  `$1` placeholders prevent SQL injection.
- **Go here when:** you need a new query (e.g. "trips by vehicle"). Write the SQL, run
  `sqlc generate`, then call the function from a service.
- **Rule:** never hand-edit files in `dm_svc_db/`. The next `sqlc generate` overwrites them.

---

### 6. Database: the vault (last line of defense)

**Files:** `migrations/*.sql` define the structure; Postgres stores the data

It enforces **hard rules** no matter what the Go code does:

```sql
driver_id  INTEGER NOT NULL REFERENCES driver(id)   -- must be a real driver
entry_type VARCHAR(50) NOT NULL                     -- at most 50 chars, not NULL
```

This is why `driver_id = 0` failed: the database refused it, even though the Go code let it through.

- **Used for:** keeping data correct even if a bug slips through the upper layers.
- **Go here when:** adding or removing columns (e.g. dropping `co_driver_name`), always with a
  **new migration** (goose). Never edit an old one.

---

## Part 4: Where each kind of check belongs

| Check | Example | Layer | Why there |
|---|---|---|---|
| Are you logged in? | missing cookie | **Middleware** | Same check for many routes |
| Is your role allowed? | manager hitting finance | **Middleware** | Same rule for a whole group |
| Is the input well-formed? | empty `entry_type`, 51 chars, negative amount | **Handler** | Reject early with a clear **400** |
| Does it make business sense? | paying more than what's owed | **Service** | Needs business knowledge |
| Is the data physically valid? | driver doesn't exist, too long | **Database** | Safety net: catches anything missed above |

**Current gap:** the `entry_type` length rule only exists in the **database**. The handler doesn't
check it, so the DB catches it, and the client gets an ugly **500** (with the raw Postgres error)
instead of a clean **400**. The fix is to add the check one layer up, in the handler.

---

## Try it yourself

Add a log line at the start of each layer:

```go
h.logger.Info("handler reached")   // in AddPayroll (handler.go)
s.logger.Info("service reached")   // in AddPayrollService (service.go)
```

Run the server locally, call the endpoint, and watch the order the lines print.
Then call it **without** the cookie: the handler line never prints, because the middleware stopped it.
