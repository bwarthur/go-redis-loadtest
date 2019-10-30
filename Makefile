test: install
	go test -mod=readonly ./...

test-race: | test
	go test -race -mod=readonly ./...

vet: | test
	go vet ./...

install:
	go install -mod=readonly cmd/redis-load/redis-load.go

run:
	go run cmd/redis-load/redis-load.go -stop-nodes -metrics-address 127.0.0.1:8080 -redis-node redis-1=localhost:6381 -redis-node redis-2=localhost:6382 -redis-node redis-3=localhost:6383 -envoy-addr localhost:6378

run-loop:
	go run cmd/redis-load/redis-load.go -metrics-address 127.0.0.1:8080 -envoy-addr localhost:6378

build:
	go build -mod=readonly cmd/redis-load/redis-load.go

fmt:
	go fmt -mod=readonly ./...

download:
	go mod tidy
	go mod download

docker-build:
	docker build -t redis-load .

docker-run:
	docker run -it --rm --network redis-ha_redis-ha redis-load -envoy-addr=redis-envoy:6379

docker-push:
	docker tag redis-load kc0isg/redis-load
	docker push kc0isg/redis-load