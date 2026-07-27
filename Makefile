.PHONY: dev build wire test test-cover fmt clean swag docker-up docker-down

# 本地开发（Air 热重载）
dev:
	air

# 构建
build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

# Wire 依赖注入代码生成
wire:
	go generate ./...

# 测试
test:
	@echo "Running tests..."
	go test ./...

test-cover:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 格式化
fmt:
	go fmt ./...

# Swagger 文档生成
swag:
	swag init -g cmd/api/main.go --output docs/swagger --outputTypes go,json,yaml --parseInternal --parseDependency

# Docker 基础设施
docker-up:
	@echo "Starting infrastructure services..."
	docker-compose up -d

docker-down:
	@echo "Stopping infrastructure services..."
	docker-compose down

# 清理
clean:
	rm -rf bin/ coverage.out coverage.html tmp/