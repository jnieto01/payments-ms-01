#!/bin/bash

# Test script with coverage report

set -e

echo "🧪 Running tests..."

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# Generate coverage report
echo ""
echo "📊 Generating coverage report..."
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

echo ""
echo "✅ Tests completed!"
echo "📈 Coverage report: coverage.html"
echo ""
echo "To view the coverage report in your browser:"
echo "  open coverage.html"
