# Makefile provides convenient shortcuts for infrastructure lifecycle.
# Use `make infra-up` before starting the server and worker.

.PHONY: infra-up infra-down

# Start all Docker services defined in docker-compose.yml (detached)
infra-up:
	docker compose up -d

# Stop and remove all Docker containers started by infra-up
infra-down:
	docker compose down
