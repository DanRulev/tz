docker-up:
	docker compose -f docker-compose.yaml up -d

docker-stop:
	docker compose -f docker-compose.yaml stop

docker-build:
	docker compose -f docker-compose.yaml build

docker-logs:
	docker compose -f docker-compose.yaml logs -f

docker-down:
	docker compose -f docker-compose.yaml down || true

docker-clean:
	docker compose -f docker-compose.yaml down --volumes || true

.PHONY: docker-up, docker-build, docker-logs, docker-down