.PHONY: test test-verbose test-coverage test-unit test-integration clean

# Jalankan semua test
test:
	go test ./...

# Jalankan dengan verbose
test-verbose:
	go test ./... -v

# Jalankan dengan coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

# Jalankan coverage dengan HTML output
test-coverage-html:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Test per layer
test-repository:
	go test ./internal/repository/... -v

test-service:
	go test ./internal/service/... -v

test-controller:
	go test ./internal/controller/... -v

# Test dengan race detector
test-race:
	go test ./... -race -v

# Test dengan timeout
test-timeout:
	go test ./... -timeout=30s -v

# Bersihkan file test cache
clean:
	go clean -testcache
	rm -f coverage.out coverage.html

# Benchmark
benchmark:
	go test ./... -bench=. -benchmem