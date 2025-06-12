# Lint Strategy for PR-Only Changed Files

This repository now has **three complementary approaches** for running lints only on changed files in PRs:

## 🎯 Approach 1: Modified Existing Workflow (ENABLED)

**File:** `.github/workflows/test.yml`

**What changed:**
- Enabled linting job for PRs only (`if: github.event_name == 'pull_request'`)
- Added `--new-from-rev=${{ github.event.pull_request.base.sha }}` flag to golangci-lint
- Enabled coverage checks for PRs only

**Benefits:**
- ✅ Integrates with existing workflow
- ✅ Only lints files changed in the PR
- ✅ Includes security scanning with gosec
- ✅ No additional configuration needed

## 🔍 Approach 2: Dedicated PR Lint Workflow (AVAILABLE)

**File:** `.github/workflows/lint-pr.yml`

**Features:**
- Runs only on PRs affecting Go files
- Detects changed files using git diff
- Posts PR comments on lint failures
- Skips when no Go files are changed

**Benefits:**
- ✅ More granular file detection
- ✅ Better PR feedback with comments
- ✅ Lighter weight than full workflow
- ✅ Clear separation of concerns

## 💬 Approach 3: Reviewdog Integration (AVAILABLE)

**File:** `.github/workflows/lint-reviewdog.yml`

**Features:**
- Uses reviewdog for inline code comments
- Posts lint issues directly on changed lines
- GitHub PR review integration
- Filter mode set to 'added' (only new code)

**Benefits:**
- ✅ Inline comments on problematic lines
- ✅ Best developer experience
- ✅ Only flags newly introduced issues
- ✅ Professional code review feel

## 🚀 Current Status

**ACTIVE:** Approach 1 (Modified existing workflow)
- Linting now runs on PRs only
- Only checks files changed in the PR
- Integrated with existing CI/CD pipeline

**DISABLED:** Approaches 2 & 3 are available but not active
- Can be enabled by creating PRs
- Ready to use if you want different behavior

## 🛠 Key Technical Details

### Changed Files Detection
```bash
# Get only the changed .go files in the PR
CHANGED_FILES=$(git diff --name-only --diff-filter=AM $BASE_SHA...$HEAD_SHA | grep '\.go$' | grep '^api/' | sed 's|^api/||' | tr '\n' ' ')

# Run golangci-lint only on those specific files
golangci-lint run --timeout=5m $CHANGED_FILES
```

**Key Filters:**
- `--diff-filter=AM` - Only Added/Modified files (not deleted)
- `grep '\.go$'` - Only Go source files
- `grep '^api/'` - Only files in the API directory
- `sed 's|^api/||'` - Remove api/ prefix for relative paths

### Path Filtering
```yaml
paths:
  - 'api/**/*.go'           # Only Go files in API
  - '.github/workflows/'    # Workflow changes
  - 'api/.golangci.yml'     # Lint config changes
```

### Performance Benefits
- ⚡ **Faster feedback** (seconds vs minutes)
- 💰 **Lower CI costs** (less compute time)
- 🎯 **Focused results** (only relevant to PR changes)
- 🔄 **Better iteration** (quick fix cycles)

## 📝 Usage

### For Developers
1. Create a PR targeting `master`, `main`, or `develop`
2. Linting automatically runs on your changed Go files
3. Fix any issues and push updates
4. Lint re-runs automatically on new commits

### Example Workflow
```bash
# If your PR changes these files:
api/internal/service/user.go          ✅ Will be linted
api/internal/http/middleware.go       ✅ Will be linted
web/src/components/Header.tsx         ❌ Ignored (not Go)
api/cmd/migrate/main.go              ✅ Will be linted
README.md                            ❌ Ignored (not Go)

# golangci-lint will only run on:
golangci-lint run --timeout=5m internal/service/user.go internal/http/middleware.go cmd/migrate/main.go
```

### For Maintainers
- All three approaches use the same `.golangci.yml` configuration
- Lint rules are consistent across all workflows
- Can switch between approaches by enabling/disabling workflows
- Security scanning (gosec) is included in the main workflow

## 🔧 Customization

### To enable different approaches:
```bash
# Enable reviewdog workflow
git mv .github/workflows/lint-reviewdog.yml.disabled .github/workflows/lint-reviewdog.yml

# Enable dedicated PR lint workflow  
git mv .github/workflows/lint-pr.yml.disabled .github/workflows/lint-pr.yml
```

### To modify lint rules:
Edit `api/.golangci.yml` - changes apply to all approaches.

### To adjust file filtering:
Modify the `paths:` section in each workflow file.

---

**Recommendation:** Start with Approach 1 (already enabled). If you want inline comments, add Approach 3 (reviewdog) later.