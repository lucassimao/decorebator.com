#!/bin/bash

# Test script to validate changed files detection logic
# Usage: ./test-changed-files.sh <base-sha> <head-sha>

BASE_SHA=${1:-"HEAD~1"}
HEAD_SHA=${2:-"HEAD"}

echo "🔍 Testing changed files detection logic"
echo "Base SHA: $BASE_SHA"
echo "Head SHA: $HEAD_SHA"
echo "----------------------------------------"

# Get list of changed .go files in the api directory
CHANGED_FILES=$(git diff --name-only --diff-filter=AM $BASE_SHA...$HEAD_SHA | grep '\.go$' | grep '^api/' | sed 's|^api/||' | tr '\n' ' ' || echo "")

echo "Raw git diff output:"
git diff --name-only --diff-filter=AM $BASE_SHA...$HEAD_SHA

echo ""
echo "Filtered .go files in api/:"
git diff --name-only --diff-filter=AM $BASE_SHA...$HEAD_SHA | grep '\.go$' | grep '^api/'

echo ""
echo "Final changed files for linting:"
echo "CHANGED_FILES='$CHANGED_FILES'"

# Check if we have any Go files to lint
if [ -z "$CHANGED_FILES" ]; then
    echo ""
    echo "❌ No Go files changed - linting will be skipped"
else
    echo ""
    echo "✅ Found Go files - linting will run on:"
    for file in $CHANGED_FILES; do
        echo "  - $file"
    done
    
    echo ""
    echo "📝 golangci-lint command that would run:"
    echo "golangci-lint run --timeout=5m $CHANGED_FILES"
fi