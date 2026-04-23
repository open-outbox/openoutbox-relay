# benchmark/Makefile

.PHONY: produce-seed, produce-bench, produce-help, bench-kafka bench-nats

FILTERED_ARGS := $(filter-out --,$(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS)))
%:
	@:
# ==========================================
# Docker Compose Producer Commands
# ==========================================

produce-seed: ## Run producer seed (e.g., make produce-seed ARGS="--batch-size 5000")
	docker compose --profile tools run --rm producer seed  $(FILTERED_ARGS)

produce-bench: ## Run producer bench (e.g., make produce-seed ARGS="--batch-size 5000")
	docker compose --profile tools run --rm producer bench $(FILTERED_ARGS)

produce-help: ## Show producer help
	docker compose --profile tools run --rm producer --help

# ==========================================
# NATS benchmark
# ==========================================
.PHONY: bench-nats
bench-nats: ## Start stack with NATS JetStream profile
	@echo "🚀 Starting Relay with NATS profile..."
	docker-compose --profile nats up -d

# ==========================================
# Kafka benchmark
# ==========================================
.PHONY: bench-kafka
bench-kafka: ## Start stack with Kafka profile
	@echo "🚀 Starting Relay with Kafka profile..."
	docker-compose --profile kafka up -d

# ==========================================
# Redis benchmark
# ==========================================
.PHONY: bench-redis
bench-redis: ## Start stack with Redis profile
	@echo "🚀 Starting Relay with Redis profile..."
	docker-compose --profile redis up -d

# ==========================================
# Clean benchmark
# ==========================================
.PHONY: bench-clean
bench-clean: ## Stop and wipe everything
	docker-compose --profile nats --profile kafka --profile redis down -v

