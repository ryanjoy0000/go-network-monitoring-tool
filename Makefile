DATA_COLLECTOR_BINARY=dataCollectorApp
CENTRAL_BINARY=centralApp

#--------------------------------------------------------------------------
#build: stops docker-compose, builds all projects and starts docker-compose
build: build_d build_c
	@echo "Stopping docker containers if any..."
	docker-compose -f ./docker-compose.yml down
	@echo "Building & starting all services ..."
	docker-compose -f ./docker-compose.yml up --build

# down : stop docker-compose
down:
	@echo "Stopping docker-compose..."
	docker-compose -f ./docker-compose.yml down

#--------------------------------------------------------------------------
#build: stops docker-compose, builds all projects and starts docker-compose
build_all: build_kafka build_d build_c
	@echo "Stopping docker containers if any..."
	docker-compose -f ./docker-compose.yml down
	@echo "Building & starting all services ..."
	docker-compose -f ./docker-compose.yml up --build

# down : stop docker-compose
down_all:
	@echo "Stopping docker-compose..."
	docker-compose -f ./docker-compose.yml down
	docker-compose -f ./kafka.yml down
# ---------------------------------------------------------------------------


# build_kafka : build all kafka containers in background
build_kafka:
	@echo "Stopping kafka containers if any..."
	docker-compose -f ./kafka.yml down
	@echo "Starting all kafka services in background..."
	docker-compose -f ./kafka.yml up --build


# build_data-collector: Build the data collector service -> binary as a linux executable
build_d:
	@echo "Building the data-collector service -> binary..."
	cd data-collector-service && env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ${DATA_COLLECTOR_BINARY} .
	@echo "data-collector service build exec ready..."

# build_central: Build the central service -> binary as a linux executable
build_c:
	@echo "Building the central service -> binary..."
	cd central-service && env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ${CENTRAL_BINARY} .
	@echo "central-service build exec ready..."
