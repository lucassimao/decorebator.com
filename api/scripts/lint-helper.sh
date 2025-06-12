#!/bin/bash

# Lint Helper Script - Easy linting for developers
# Usage: ./scripts/lint-helper.sh [command]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

function print_usage() {
    echo -e "${BLUE}🔍 Lint Helper - Easy Go Linting${NC}"
    echo ""
    echo -e "${YELLOW}Usage:${NC} ./scripts/lint-helper.sh [command]"
    echo ""
    echo -e "${CYAN}Available commands:${NC}"
    echo "  help           Show this help message"
    echo "  install        Install golangci-lint"
    echo "  check          Quick check - lint only changed files"
    echo "  fix            Fix linting issues in changed files"
    echo "  staged         Lint only staged files (for pre-commit)"
    echo "  all            Lint entire codebase"
    echo "  format         Format code with gofmt and goimports"
    echo "  pre-commit     Run pre-commit checks"
    echo "  pre-push       Run pre-push checks"
    echo "  setup-hooks    Install git hooks for automated linting"
    echo ""
    echo -e "${CYAN}Examples:${NC}"
    echo "  ./scripts/lint-helper.sh check      # Lint only your changes"
    echo "  ./scripts/lint-helper.sh fix        # Auto-fix linting issues"
    echo "  ./scripts/lint-helper.sh staged     # Check staged files"
    echo ""
    echo -e "${CYAN}Git Integration:${NC}"
    echo "  make git-hooks-install              # Install git hooks"
    echo "  make lint-changed                   # Lint changes vs main"
    echo "  make lint-branch BASE=develop       # Lint changes vs develop"
}

function check_golangci_lint() {
    if ! command -v golangci-lint >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  golangci-lint not found. Installing...${NC}"
        make lint-install
    fi
}

function lint_changed_files() {
    echo -e "${BLUE}🔍 Linting changed files since main...${NC}"
    
    # Get changed files for display
    CHANGED_FILES=$(git diff --name-only --diff-filter=AM origin/main...HEAD | grep '\.go$' | grep '^api/' | sed 's|^api/||' || echo "")
    
    if [ -z "$CHANGED_FILES" ]; then
        echo -e "${GREEN}✅ No Go files changed since main branch${NC}"
        return 0
    fi
    
    echo -e "${CYAN}📁 Changed files:${NC} $CHANGED_FILES"
    
    # Use --new-from-rev for proper linting
    MERGE_BASE=$(git merge-base origin/main HEAD)
    if golangci-lint run --timeout=5m --new-from-rev=$MERGE_BASE; then
        echo -e "${GREEN}✅ All linting checks passed!${NC}"
    else
        echo -e "${RED}❌ Linting issues found. Run './scripts/lint-helper.sh fix' to auto-fix some issues.${NC}"
        return 1
    fi
}

function fix_changed_files() {
    echo -e "${BLUE}🔧 Fixing linting issues in changed files...${NC}"
    
    # Get changed files for display
    CHANGED_FILES=$(git diff --name-only --diff-filter=AM origin/main...HEAD | grep '\.go$' | grep '^api/' | sed 's|^api/||' || echo "")
    
    if [ -z "$CHANGED_FILES" ]; then
        echo -e "${GREEN}✅ No Go files changed since main branch${NC}"
        return 0
    fi
    
    echo -e "${CYAN}📁 Fixing files:${NC} $CHANGED_FILES"
    
    # Use --new-from-rev for proper linting with fix
    MERGE_BASE=$(git merge-base origin/main HEAD)
    if golangci-lint run --fix --timeout=5m --new-from-rev=$MERGE_BASE; then
        echo -e "${GREEN}✅ Auto-fix completed!${NC}"
        
        # Check if there are any remaining issues
        if ! golangci-lint run --timeout=5m --new-from-rev=$MERGE_BASE >/dev/null 2>&1; then
            echo -e "${YELLOW}⚠️  Some issues require manual fixing. Run 'make lint-changed' to see details.${NC}"
        fi
    else
        echo -e "${RED}❌ Some issues couldn't be auto-fixed. Please fix manually.${NC}"
        return 1
    fi
}

function lint_staged_files() {
    echo -e "${BLUE}🔍 Linting staged files...${NC}"
    
    # Get staged files for display
    STAGED_FILES=$(git diff --cached --name-only --diff-filter=AM | grep '\.go$' | grep '^api/' | sed 's|^api/||' || echo "")
    
    if [ -z "$STAGED_FILES" ]; then
        echo -e "${GREEN}✅ No Go files staged${NC}"
        return 0
    fi
    
    echo -e "${CYAN}📁 Staged files:${NC} $STAGED_FILES"
    
    # Use --new-from-rev for staged files (compare against HEAD~1)
    if golangci-lint run --timeout=5m --new-from-rev=HEAD~1; then
        echo -e "${GREEN}✅ All staged files pass linting!${NC}"
    else
        echo -e "${RED}❌ Linting issues found in staged files.${NC}"
        echo -e "${YELLOW}💡 Run './scripts/lint-helper.sh fix' to auto-fix before committing.${NC}"
        return 1
    fi
}

# Main command handling
case "${1:-help}" in
    "help"|"-h"|"--help")
        print_usage
        ;;
    "install")
        echo -e "${BLUE}📦 Installing golangci-lint...${NC}"
        make lint-install
        ;;
    "check")
        check_golangci_lint
        lint_changed_files
        ;;
    "fix")
        check_golangci_lint
        fix_changed_files
        ;;
    "staged")
        check_golangci_lint
        lint_staged_files
        ;;
    "all")
        echo -e "${BLUE}🔍 Linting entire codebase...${NC}"
        check_golangci_lint
        make lint
        ;;
    "format")
        echo -e "${BLUE}🎨 Formatting code...${NC}"
        make format
        ;;
    "pre-commit")
        echo -e "${BLUE}🔒 Running pre-commit checks...${NC}"
        check_golangci_lint
        make pre-commit
        ;;
    "pre-push")
        echo -e "${BLUE}🚀 Running pre-push checks...${NC}"
        check_golangci_lint
        make pre-push
        ;;
    "setup-hooks")
        echo -e "${BLUE}⚙️  Setting up git hooks...${NC}"
        make git-hooks-install
        echo -e "${GREEN}✅ Git hooks installed! They will run automatically on commit/push.${NC}"
        ;;
    *)
        echo -e "${RED}❌ Unknown command: $1${NC}"
        echo ""
        print_usage
        exit 1
        ;;
esac