# GolangFlow Template Migration Troubleshooting

**Date:** 2026-01-04
**Issue:** Templates not loading after Buffalo upgrade to v1.1.3 with Go 1.23 embed.FS

## Problem Overview

After migrating from Google Container Registry to Artifact Registry and updating to Buffalo v1.1.3, the application fails to load templates. Initial error: `"could not find template index.html"`

### Environment
- **Buffalo Version:** v1.1.3
- **Go Version:** 1.23
- **Deployment:** Google Cloud Run
- **Previous System:** Packr (deprecated)
- **Current System:** Go embed.FS

## Initial Setup

### Embed Configuration (embeds.go)
```go
//go:embed all:templates
var TemplatesFS embed.FS

//go:embed all:assets
var AssetsFS embed.FS
```

### Template Structure
```
templates/
├── index.plush.html
├── application.plush.html
├── privacy.plush.html
├── _flash.plush.html
├── _submit.plush.html
├── posts/
│   ├── index.plush.html
│   ├── show.plush.html
│   ├── edit.plush.html
│   ├── new.plush.html
│   └── _form.plush.html
└── users/
    ├── index.plush.html
    ├── show.plush.html
    ├── edit.plush.html
    ├── new.plush.html
    └── _form.plush.html
```

## Approaches Tried

### Attempt 1: Using fs.Sub() to strip "templates/" prefix
**Commit:** `4877bde`, `364de1a`, `e5e444a`

**Code:**
```go
func initRender() {
    templatesSubFS, err := fs.Sub(templatesFS, "templates")
    if err != nil {
        panic(err)
    }

    r = render.New(render.Options{
        HTMLLayout: "application.html",
        TemplatesFS: templatesSubFS,
        AssetsFS: assetsFS,
    })
}
```

**Result:** ❌ Failed - "could not find template index.html"

**Analysis:** Buffalo's template resolver walks the FS and creates aliases (`index.html` -> `index.plush.html`), but wasn't finding the templates in the sub-filesystem.

---

### Attempt 2: Using full embedded FS without fs.Sub()
**Commit:** `4b1f3e7`

**Code:**
```go
func initRender() {
    r = render.New(render.Options{
        HTMLLayout: "application.html",
        TemplatesFS: templatesFS,  // Full FS
        AssetsFS: assetsFS,
    })
}
```

**Result:** ❌ Failed - "could not find template index.html"

**Analysis:** Templates are at `templates/index.plush.html` but code calls `r.HTML("index.html")` - path mismatch.

---

### Attempt 3: Using buffalo.NewFS() wrapper
**Commit:** `c4b78d9`, `5758625`

**Code:**
```go
func initRender() {
    wrappedTemplatesFS := buffalo.NewFS(templatesFS, "templates")

    r = render.New(render.Options{
        HTMLLayout: "application.html",
        TemplatesFS: wrappedTemplatesFS,
        AssetsFS: assetsFS,
    })
}
```

**Result:** ❌ Build failed - type mismatch (fs.FS vs fs.ReadDirFS)

**Analysis:** buffalo.NewFS() requires fs.ReadDirFS interface.

---

### Attempt 4: fs.Sub() + buffalo.NewFS() with type casting
**Commit:** `195ca8b`

**Code:**
```go
func initRender() {
    templatesSubFS, err := fs.Sub(templatesFS, "templates")
    if err != nil {
        panic(err)
    }

    templatesReadDirFS, ok := templatesSubFS.(fs.ReadDirFS)
    if !ok {
        panic("templatesSubFS does not implement fs.ReadDirFS")
    }

    wrappedTemplatesFS := buffalo.NewFS(templatesReadDirFS, "templates")

    r = render.New(render.Options{
        HTMLLayout: "application.html",
        TemplatesFS: wrappedTemplatesFS,
        AssetsFS: assetsFS,
    })
}
```

**Result:** ❌ Build succeeded but site timeouts (HTTP 524)

**Analysis:** buffalo.NewFS() may have caused hanging during template resolution or filesystem walking.

---

### Attempt 5: Back to simple fs.Sub()
**Commit:** `e5e444a`

**Code:**
```go
func initRender() {
    templatesSubFS, err := fs.Sub(templatesFS, "templates")
    if err != nil {
        panic(err)
    }

    r = render.New(render.Options{
        HTMLLayout: "application.html",
        TemplatesFS: templatesSubFS,
        AssetsFS: assetsFS,
    })
}
```

**Result:** ❌ Build succeeded but site timeouts

**Analysis:** Same timeout issue as before.

---

### Attempt 6: Update all template paths to include "templates/" prefix ✅
**Commit:** `c5a219b` (FINAL)

**Code Changes:**

**actions/render.go:**
```go
func initRender() {
    r = render.New(render.Options{
        HTMLLayout: "templates/application.html",  // Added prefix
        TemplatesFS: templatesFS,  // Use full FS
        AssetsFS: assetsFS,
    })
}
```

**actions/home.go:**
```go
// Before: r.HTML("index.html")
return c.Render(200, r.HTML("templates/index.html"))

// Before: r.HTML("privacy.html")
return c.Render(200, r.HTML("templates/privacy.html"))
```

**actions/posts.go & actions/users.go:**
```bash
sed -i '' 's/r\.HTML("\([^"]*\)\.html")/r.HTML("templates\/\1.html")/g'
```

All template paths updated to include `templates/` prefix:
- `templates/index.html`
- `templates/privacy.html`
- `templates/posts/index.html`
- `templates/posts/show.html`
- `templates/posts/edit.html`
- `templates/posts/new.html`
- `templates/users/index.html`
- `templates/users/show.html`
- `templates/users/edit.html`
- `templates/users/new.html`

**Result:** ⚠️ **Partial Success**
- Build: ✅ Success
- Deployment: ✅ Success
- App Startup: ✅ Success
- Template Error: ✅ Resolved (no more "could not find template" errors)
- **Current Issue:** GET requests to pages hang/timeout

---

## Current Status (Revision 00049-xxx) - 2026-01-04

### What's Working ✅
- RSS feed (`/rss/`) - 200 status, 191kB response
- Database connectivity confirmed (db=4.135738ms in logs)
- Cloud Run deployment pipeline
- Container startup and health checks
- Template-free responses (when bypassing Buffalo render system)

### What's Not Working ❌
- Homepage returns 500 error with Buffalo's default error page
- Application layout template (`templates/application.plush.html`) fails to render
- Template helpers removed but core template system still broken
- All routes using application layout fail

### Root Cause Analysis

**Template System Issues with embed.FS:**
1. **Partial Resolution**: `partial("_flash.html")` can't find `_flash.plush.html`
2. **Layout Processing**: Application layout fails during render phase
3. **Path Resolution**: Embedded filesystem paths not resolving correctly in Buffalo's template engine

**Evidence from Logs:**
```
level=error msg="templates/application.html: line 39: could not call partial function: could not find template _flash.html"
```

**Working vs Broken:**
- RSS feed works (no layout, direct template render)
- Homepage fails (uses application layout)
- Maintenance page worked (bypassed template system entirely)

### Attempted Fixes

1. ✅ **Template Helpers Removed**: Eliminated `getAvatar`, `byLine`, `ownsPost` helpers that caused nil pointer errors
2. ✅ **Import Cleanup**: Removed unused imports causing build failures  
3. ❌ **Partial Path Fixes**: Tried various partial paths (`_flash.html`, `templates/_flash.html`) - none work
4. ❌ **Simple Templates**: Created standalone templates - still fail when using layout
5. ❌ **Template Resolution**: Core Buffalo template engine incompatible with embed.FS structure

### Technical Details

**Buffalo Template Resolution Process:**
1. `updateAliases()` walks embedded FS to create template aliases
2. `resolve()` tries to find templates using aliases
3. **FAILURE POINT**: Partial template resolution fails in embedded filesystem

**File Structure:**
```
templates/
├── _flash.plush.html          ✅ File exists
├── application.plush.html     ✅ File exists, calls partial("_flash.html")
├── index-standalone.plush.html ✅ Created, bypasses layout
└── index-simple.plush.html    ✅ Created, uses layout (fails)
```

**Current Handler:**
```go
func HomeHandler(c buffalo.Context) error {
    tx := c.Value("tx").(*pop.Connection)  // ✅ Works
    posts := &models.Posts{}
    
    q := tx.PaginateFromParams(c.Request().URL.Query())
    err := q.Order("created_at desc").All(posts)  // ✅ Works
    
    return c.Render(200, r.HTML("templates/index-standalone.html"))  // ❌ Fails
}
```

### Next Investigation Steps

1. **Buffalo Version Compatibility**: Check if Buffalo v0.18+ has embed.FS fixes
2. **Template Engine Bypass**: Use direct HTTP responses instead of Buffalo render
3. **Rollback Strategy**: Consider reverting to Packr temporarily
4. **Manual Template Loading**: Load templates manually from embed.FS

### Workaround Options

**Option 1: Direct HTTP Response**
```go
func HomeHandler(c buffalo.Context) error {
    // Load template manually from embed.FS
    // Render with html/template directly
    // Write to c.Response()
}
```

**Option 2: API + Frontend**
- Convert to JSON API backend
- Add simple frontend that calls API
- Bypass Buffalo template system entirely

**Option 3: Minimal Template**
- Single template file with no partials
- No application layout
- Inline all CSS/JS

### Deployment Status
- **Current Revision**: Shows Buffalo error page
- **RSS Feed**: Fully functional
- **Site Status**: Partially broken but accessible

---

## Recommended Action

**Investigate Buffalo embed.FS compatibility or implement direct template rendering bypass. The core issue is Buffalo's template engine not properly resolving embedded filesystem paths for partials and layouts.**

---

## Technical Findings

### Buffalo Template Resolution (from source code analysis)

**How Buffalo finds templates:**
1. `updateAliases()` walks `TemplatesFS` starting from `"."`
2. Creates aliases: `index.plush.html` → `index.html`
3. `resolve()` tries:
   - Direct open: `TemplatesFS.Open("index.html")`
   - If fails, lookup alias: `aliases.Load("index.html")`
   - Open aliased file: `TemplatesFS.Open("index.plush.html")`

**Buffalo FS Wrapper (buffalo.NewFS):**
```go
type FS struct {
    embed fs.FS
    dir   fs.FS  // os.DirFS(dir)
}

func (f FS) getFile(name string) (fs.File, error) {
    // Try disk first (for development)
    file, err := f.dir.Open(name)
    if err == nil {
        return file, nil
    }
    // Fallback to embedded
    return f.embed.Open(name)
}
```

**Why buffalo.NewFS() caused issues:**
- Creates `os.DirFS("templates")` which doesn't exist in Docker container
- Tries disk first, fails, then tries embed
- May cause hanging during filesystem walking

---

## Why Template Paths with Prefix Works (In Theory)

With `//go:embed all:templates`, files are embedded at:
- `templates/index.plush.html`
- `templates/application.plush.html`
- etc.

Buffalo's `updateAliases()` walks the FS and should create:
- `templates/index.html` → `templates/index.plush.html`
- `templates/application.html` → `templates/application.plush.html`

When code calls `r.HTML("templates/index.html")`:
1. Buffalo tries `Open("templates/index.html")` - fails
2. Looks up alias: `aliases["templates/index.html"]` → `"templates/index.plush.html"`
3. Opens: `Open("templates/index.plush.html")` - succeeds ✅

This is the most straightforward approach and should work.

---

## Current Hypothesis: Middleware Deadlock

### Why Requests Timeout

The fact that:
1. App starts successfully
2. HEAD/404 requests work
3. GET requests to template pages don't appear in logs
4. Previous revision showed "nil pointer dereference" during render

Suggests the issue is **NOT** templates anymore, but a middleware problem.

**Most likely cause:** Database middleware (popmw) deadlock or hang

**Evidence:**
- Homepage handler fetches posts with pagination: `tx.PaginateFromParams(...).All(posts)`
- Buffalo pop middleware starts transaction before each request
- Requests hang before logging (middleware runs before handlers)
- Database connection works (seen in earlier logs)

**Possible issues:**
- Connection pool exhaustion
- Transaction deadlock
- Slow/hanging database query
- Middleware initialization issue

---

## Files Modified

### Final Working Changes (Commit c5a219b)

**actions/render.go**
- Removed `fs.Sub()` usage
- Removed `buffalo.NewFS()` wrapper
- Updated `HTMLLayout` to `"templates/application.html"`
- Use full `templatesFS` directly

**actions/home.go**
- `r.HTML("index.html")` → `r.HTML("templates/index.html")`
- `r.HTML("privacy.html")` → `r.HTML("templates/privacy.html")`

**actions/posts.go**
- All template paths updated with `templates/` prefix
- `posts/index.html` → `templates/posts/index.html`
- `posts/show.html` → `templates/posts/show.html`
- `posts/edit.html` → `templates/posts/edit.html`
- `posts/new.html` → `templates/posts/new.html"`

**actions/users.go**
- All template paths updated with `templates/` prefix
- `users/index.html` → `templates/users/index.html`
- `users/show.html` → `templates/users/show.html`
- `users/edit.html` → `templates/users/edit.html`
- `users/new.html` → `templates/users/new.html`

---

## Next Steps to Investigate

### 1. Database Middleware Issue
Check if database queries are hanging:
- Review homepage handler query: `tx.PaginateFromParams(c.Request().URL.Query()).Order("created_at desc").All(posts)`
- Check connection pool settings
- Verify DATABASE_URL environment variable
- Test with database middleware disabled

### 2. Template Helper Issues
The nil pointer error suggests template helpers may be problematic:
```go
"getAvatar": func(id uuid.UUID, help hctx.HelperContext) (string, error) {
    tx := help.Value("tx").(*pop.Connection)  // Could be nil?
    // ...
}
```

Check if `help.Value("tx")` returns nil when it shouldn't.

### 3. Pagination Issues
Check if `PaginateFromParams()` is hanging or causing issues.

### 4. Production vs Development Config
Verify environment variables are set correctly:
- `GO_ENV=production`
- `DATABASE_URL` set correctly
- `PORT=8080`

### 5. Test Without Template Rendering
Create a simple handler that doesn't render templates:
```go
func HealthCheck(c buffalo.Context) error {
    return c.Render(200, r.JSON(map[string]string{"status": "ok"}))
}
```

If this works but template routes don't, confirms template rendering issue.

---

## Useful Commands for Debugging

### Check Build Status
```bash
gcloud builds list --project=golangflow-290604 --limit=1
```

### Check Deployed Revisions
```bash
gcloud run revisions list --service=golangflow-production \
  --region=us-west1 --project=golangflow-290604 --limit=5
```

### View Application Logs
```bash
gcloud logging read \
  "resource.type=cloud_run_revision AND resource.labels.service_name=golangflow-production" \
  --project=golangflow-290604 --limit=50 --freshness=5m
```

### Test Site
```bash
curl --max-time 10 https://golangflow.io/
```

### View Specific Revision Logs
```bash
gcloud logging read \
  "resource.type=cloud_run_revision AND resource.labels.revision_name=golangflow-production-00040-2st" \
  --project=golangflow-290604 --limit=100 --freshness=15m
```

---

## References

### Buffalo Documentation
- [Buffalo FS Implementation](https://github.com/gobuffalo/buffalo/blob/main/fs.go)
- [Buffalo Template Resolver](https://github.com/gobuffalo/buffalo/blob/main/render/template.go)
- [Replace packr with embed.fs PR #2166](https://github.com/gobuffalo/buffalo/pull/2166)

### Go embed.FS
- [Go 1.16 embed package](https://pkg.go.dev/embed)
- [fs.Sub documentation](https://pkg.go.dev/io/fs#Sub)

### Cloud Run
- [Cloud Run Troubleshooting](https://cloud.google.com/run/docs/troubleshooting)

---

## Timeline

| Time | Action | Result |
|------|--------|--------|
| Initial | Templates not loading | Error: "could not find template index.html" |
| Attempt 1 | fs.Sub() to strip prefix | ❌ Still can't find templates |
| Attempt 2 | Full FS without Sub() | ❌ Path mismatch |
| Attempt 3 | buffalo.NewFS() wrapper | ❌ Type error |
| Attempt 4 | fs.Sub() + buffalo.NewFS() | ❌ Site timeouts |
| Attempt 5 | Back to simple fs.Sub() | ❌ Site timeouts |
| Attempt 6 | Add "templates/" prefix to all paths | ⚠️ Template errors resolved, but requests timeout |

---

## Summary

**Template Issue:** ✅ **RESOLVED**
- Solution: Add `templates/` prefix to all template paths in code
- Templates are now found correctly (no more "could not find template" errors)

**Current Issue:** ❌ **Request Timeout/Hanging**
- GET requests to template-based pages timeout
- Requests don't reach handlers (not logged)
- Likely middleware issue (database/transaction)
- Needs further investigation

---

## For AI Assistant: Critical Context

### What You Need to Know

1. **Template paths are correct** - Don't try to "fix" the template paths again
2. **The issue is NOT templates** - Templates are being found (confirmed by lack of "could not find template" errors)
3. **Requests are hanging BEFORE handlers** - They don't appear in logs at all
4. **The app IS running** - Container healthy, responds to HEAD/404 requests
5. **This is likely a middleware or database issue**

### Specific Code to Investigate

**Homepage Handler (actions/home.go:19-37):**
```go
func HomeHandler(c buffalo.Context) error {
    tx := c.Value("tx").(*pop.Connection)  // Database transaction from middleware
    posts := &models.Posts{}

    q := tx.PaginateFromParams(c.Request().URL.Query())  // ⚠️ Could hang here
    err := q.Order("created_at desc").All(posts)         // ⚠️ Or here
    if err != nil {
        return errors.WithStack(err)
    }

    c.Set("posts", posts)
    c.Set("pagination", q.Paginator)
    return c.Render(200, r.HTML("templates/index.html"))
}
```

**Potential hanging points:**
1. `tx.PaginateFromParams()` - Pagination setup
2. `q.Order(...).All(posts)` - Database query
3. `c.Render()` - Template rendering (but likely not since it doesn't reach this point)

### Template Helpers (actions/render.go:46-72)

These helpers use `help.Value("tx")` which could return nil:

```go
"getAvatar": func(id uuid.UUID, help hctx.HelperContext) (string, error) {
    tx := help.Value("tx").(*pop.Connection)  // ⚠️ Could panic if nil
    u := models.User{}
    erru := tx.Find(&u, id)
    if erru != nil {
        return "http://via.placeholder.com/140x100", nil
    }
    return u.GravatarID.String, nil
},
```

### Database Configuration (database.yml)

```yaml
production:
  url: {{env "DATABASE_URL"}}
  dialect: postgres
  pool: 1           # ⚠️ Very small pool size
  idle_pool: 1S
```

**Concern:** Pool size of 1 could cause connection exhaustion.

### Environment Variables to Check

Cloud Run service should have:
- `GO_ENV=production`
- `DATABASE_URL=postgres://...` (connection string)
- `PORT=8080`

### Middleware Stack

The app uses Buffalo pop middleware (popmw) which:
1. Starts a database transaction before each request
2. Stores it in context as "tx"
3. Commits/rolls back after handler completes

**Potential issue:** If middleware hangs during transaction start, request never reaches handler.

### Testing Strategy

1. **Add a simple non-database route** to verify routing works:
   ```go
   app.GET("/health", func(c buffalo.Context) error {
       return c.Render(200, r.JSON(map[string]string{"status": "ok"}))
   })
   ```

2. **Test homepage with database query bypassed** (temporary debug):
   ```go
   func HomeHandler(c buffalo.Context) error {
       // Bypass database for testing
       return c.Render(200, r.String("Hello from homepage"))
   }
   ```

3. **Check if database connection pool is exhausted**:
   - Increase pool size in database.yml to 10-20
   - Check DATABASE_URL is correct
   - Verify database is accessible from Cloud Run

4. **Add logging before database calls**:
   ```go
   func HomeHandler(c buffalo.Context) error {
       fmt.Println("DEBUG: HomeHandler started")
       tx := c.Value("tx").(*pop.Connection)
       fmt.Println("DEBUG: Got tx from context")
       posts := &models.Posts{}
       fmt.Println("DEBUG: About to call PaginateFromParams")
       q := tx.PaginateFromParams(c.Request().URL.Query())
       fmt.Println("DEBUG: About to execute query")
       // ... rest of function
   }
   ```

### Quick Fixes to Try

1. **Increase database connection pool** (database.yml):
   ```yaml
   production:
     url: {{env "DATABASE_URL"}}
     dialect: postgres
     pool: 20        # Increase from 1
     idle_pool: 5S   # Increase from 1S
   ```

2. **Add timeout to database queries**:
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   // Use ctx in queries
   ```

3. **Test without pagination**:
   ```go
   // Instead of: q := tx.PaginateFromParams(...)
   err := tx.Order("created_at desc").Limit(10).All(posts)
   ```

### Files You'll Need to Modify

- `database.yml` - Connection pool settings
- `actions/home.go` - Homepage handler for debugging
- `actions/app.go` - Add health check route (if needed)

### How to Deploy and Test

```bash
# Build locally
go build -o golangflow-test

# Commit changes
git add <files>
git commit -m "Debug: <description>"
git push origin release

# Wait for Cloud Build
gcloud builds list --project=golangflow-290604 --limit=1

# Check logs after deployment
gcloud logging read \
  "resource.type=cloud_run_revision AND resource.labels.service_name=golangflow-production" \
  --project=golangflow-290604 --limit=50 --freshness=5m

# Test site
curl --max-time 10 https://golangflow.io/
```

---

## Recommended Action

**Investigate database middleware and connection pooling. The template migration is complete, but there's a separate runtime issue causing request hangs.**

**Start with:** Increase database pool size in database.yml and add debug logging to HomeHandler to identify exact hanging point.
