DATA_COLLECTOR_BINARY=dataCollectorApp
CENTRAL_BINARY=centralApp

#build: stops docker-compose, builds all projects and starts docker-compose
build: build_kafka build_d build_c
	@echo "Stopping docker containers if any..."
	docker-compose -f ./docker-compose.yml down
	@echo "Building & starting all services ..."
	docker-compose -f ./docker-compose.yml up -d --build

# down : stop docker-compose
down:
	@echo "Stopping docker-compose..."
	docker-compose -f ./docker-compose.yml down
# ---------------------------------------------------------------------------

# build_kafka : build all kafka containers in background
build_kafka:
	@echo "Starting all kafka services in background..."
	docker-compose -f ./docker-compose.yml up -d --build zookeeper kafka-1 kafka-2

# build_: Build the data collector service -> binary as a linux executable
build_d:
	@echo "Building the data-collector service -> binary..."
	cd data-collector-service && env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ${DATA_COLLECTOR_BINARY} .
	@echo "data-collector service build exec ready..."
	@echo "Building & starting service..."
	docker-compose -f ./docker-compose.yml up -d --build data-collector-service

# down_d : stop data_collector
down_d:
	@echo "Stopping docker-compose..."
	docker-compose -f ./docker-compose.yml down data-collector-service

# build_c: Build the central service -> binary as a linux executable
build_c:
	@echo "Building the central service -> binary..."
	cd central-service && env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ${CENTRAL_BINARY} .
	@echo "central-service build exec ready..."
	@echo "Building & starting service..."
	docker-compose -f ./docker-compose.yml up --build central-service

# down_c : stop central-service 
down_c:
	@echo "Stopping central-service..."
	docker-compose -f ./docker-compose.yml down central-service 
