.PHONY: backup dashboard docker-build docker-backup docker-dashboard docker-stop

# Running backup
backup:
	go run main.go backup

# Running dashboard with optional port
dashboard:
	go run main.go dashboard $(if $(port),--port=$(port))

# Docker: build image
docker-build:
	docker compose build

# Docker: running backup inside container and remove container after done
docker-backup:
	docker compose run --rm backup

# Docker: running dashboard inside container and remove container after done
docker-dashboard:
	docker compose run --rm -p $(if $(port),$(port),8080):8080 dashboard

# Docker: stop all running containers
docker-stop:
	docker compose stop

# Docker: stop and remove all containers, networks, and images defined in the file
docker-down:
	docker compose down