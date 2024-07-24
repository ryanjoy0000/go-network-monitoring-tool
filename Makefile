# ----------------------------
# up : start all containers in background without forcing build
up:
	@echo "Starting all services in background..."
	docker-compose -f ./docker-compose.yml up -d

# up_build: stops docker-compose, builds all projects and starts docker-compose
up_build:
	@echo "Stopping docker containers if any..."
	docker-compose -f ./docker-compose.yml down
	@echo "Building & starting all services ..."
	docker-compose -f ./docker-compose.yml up --build

# down : stop docker-compose
down:
	@echo "Stopping docker-compose..."
	docker-compose -f ./docker-compose.yml down
# ------------------------------
