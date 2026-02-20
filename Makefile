.PHONY: backup dashboard build.local build.run.backup build.run.dashboard docker.build docker.backup docker.dashboard docker-stop

# Running backup
backup:
	@echo "Starting backup from source code..."
	go run main.go backup

# Running dashboard with optional port
dashboard:
	@echo "Starting dashboard from source code..."
	go run main.go dashboard $(if $(port),--port=$(port))

# Building binary locally
build.local:
	@echo "Building binary..."
	@if not exist "bin" mkdir "bin"
	@go build -o bin/go-backup main.go
	@echo "Build successful! Binary located at ./bin/go-backup"

# Compile binary and directly run BACKUP
build.run.backup: build.local
	@echo "Running compiled backup binary..."
	./bin/go-backup backup

# Compile binary and directly run DASHBOARD with optional port
build.run.dashboard: build.local
	@echo "Running compiled dashboard binary..."
	./bin/go-backup dashboard $(if $(port),--port=$(port))

# Docker: build image
docker.build:
	@echo "Building Docker images defined in docker-compose.yml..."
	docker compose build
	@echo "Docker build completed."

# Docker: running backup inside container
docker.backup:
	@echo "Running backup service in a temporary Docker container..."
	docker compose run --rm backup
	@echo "Backup container finished and removed."

# Docker: running dashboard inside container
docker.dashboard:
	@echo "Starting dashboard service in Docker on port $(if $(port),$(port),8080)..."
	docker compose run --rm -p $(if $(port),$(port),8080):8080 dashboard

# Docker: stop all running containers
docker.stop:
	@echo "Stopping all running containers for this project..."
	docker compose stop
	@echo "Containers stopped."

# Docker: stop and remove all resources
docker.down:
	@echo "Shutting down and removing all containers, networks, and images..."
	docker compose down
	@echo "Cleanup completed."
