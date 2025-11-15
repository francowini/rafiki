# ==============================================================================
# Rafiki Partner Service - Makefile
# ==============================================================================

# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

# Production server configuration
PROD_SERVER := root@178.156.170.37
PROD_PATH := /opt/rafiki

# ==============================================================================
# Development Commands
# ==============================================================================

run:
	go run api/services/partner/main.go | go run api/tooling/logfmt/main.go

help:
	go run api/services/partner/main.go --help

version:
	go run api/services/partner/main.go --version

# ==============================================================================
# Docker Compose - Local Development
# ==============================================================================

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f partner-service

logs-all:
	docker compose logs -f

build:
	docker compose build

rebuild:
	docker compose build --no-cache

restart:
	docker compose restart partner-service

ps:
	docker compose ps

# ==============================================================================
# Health Check Commands
# ==============================================================================

curl-ready:
	curl -i http://localhost:3000/v1/readiness

curl-live:
	curl -i http://localhost:3000/v1/liveness

health:
	@echo "🔍 Checking API health..."
	@curl -sf http://localhost:3000/v1/readiness > /dev/null && echo "✅ API is ready" || echo "❌ API is not ready"
	@curl -sf http://localhost:3000/v1/liveness > /dev/null && echo "✅ API is alive" || echo "❌ API is not alive"

# ==============================================================================
# Production Deployment Commands
# ==============================================================================

deploy:
	@echo "🚀 Deploying to production..."
	@echo "📍 Server: $(PROD_SERVER)"
	@echo "📂 Path: $(PROD_PATH)"
	@echo ""
	@ssh $(PROD_SERVER) 'cd $(PROD_PATH) && git pull origin main && ./devops/deploy.sh'

deploy-logs:
	@echo "📋 Fetching production logs..."
	@ssh $(PROD_SERVER) 'cd $(PROD_PATH) && docker compose logs -f partner-service'

deploy-status:
	@echo "📊 Production status..."
	@ssh $(PROD_SERVER) 'cd $(PROD_PATH) && docker compose ps'

deploy-health:
	@echo "🔍 Checking production health..."
	@ssh $(PROD_SERVER) 'curl -sf http://localhost:3000/v1/readiness && echo "✅ Production is healthy" || echo "❌ Production health check failed"'

deploy-restart:
	@echo "🔄 Restarting production service..."
	@ssh $(PROD_SERVER) 'cd $(PROD_PATH) && docker compose restart partner-service'

ssh:
	@echo "🔐 Connecting to production server..."
	@ssh $(PROD_SERVER)

# ==============================================================================
# Database Commands
# ==============================================================================

db-shell:
	docker exec -it rafiki-postgres psql -U rafiki -d rafiki

db-shell-prod:
	ssh $(PROD_SERVER) 'docker exec -it rafiki-postgres psql -U rafiki -d rafiki'

# ==============================================================================
# User Management
# ==============================================================================

create-user:
	@echo "👤 Create user (local):"
	@echo "Usage: ./zarf/create-user.sh <email> <password> <role> <name>"
	@echo "Example: ./zarf/create-user.sh admin@rafiki.lat secret123 ADMIN 'Admin User'"

create-user-prod:
	@echo "👤 Create user (production):"
	@echo "Run on server: ssh $(PROD_SERVER)"
	@echo "Then: cd $(PROD_PATH) && ./zarf/create-user.sh <email> <password> <role> <name>"

# ==============================================================================
# Dependencies
# ==============================================================================

tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	go get -u -v ./...
	go mod tidy
	go mod vendor

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

deps-list:
	go list -m -u -mod=readonly all

# ==============================================================================
# Cleanup
# ==============================================================================

clean:
	@echo "🧹 Cleaning up..."
	go clean -cache
	docker compose down -v
	rm -rf vendor/

clean-all: clean
	@echo "🧹 Deep clean (removes all Docker artifacts)..."
	docker system prune -af --volumes

# ==============================================================================
# Help
# ==============================================================================

.PHONY: run help version up down logs logs-all build rebuild restart ps \
        curl-ready curl-live health \
        deploy deploy-logs deploy-status deploy-health deploy-restart ssh \
        db-shell db-shell-prod create-user create-user-prod \
        tidy deps-upgrade deps-reset deps-list \
        clean clean-all
