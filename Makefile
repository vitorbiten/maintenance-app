## Development:
env: ## Copies .env.example files
	cp .env.example .env
	cp api/.env.example api/.env
	cp worker/.env.example worker/.env
	cp kubernetes/secrets-example.yaml kubernetes/secrets.yaml

run: ## Starts development server with hot-reload
	docker-compose -p dev up -d --build

stop: ## Stops all containers
	docker-compose -p dev down --volumes
	docker-compose -f docker-compose.test.yml -p test down --volumes

lint: ## Uses golangci-lint to lint
	docker run -t --rm -v "$(CURDIR)/api:/app" -w /app golangci/golangci-lint:v1.50.1 golangci-lint run -v
	docker run -t --rm -v "$(CURDIR)/worker:/app" -w /app golangci/golangci-lint:v1.50.1 golangci-lint run -v

## Migrations:
migrate-up:
	sql-migrate up

migrate-down:
	sql-migrate down

create:
	@read -p  "What is the name of migration?" NAME; \
	sql-migrate new $$NAME 

## Test:
test: ## Run the tests on the project
	docker-compose -f docker-compose.test.yml -p test up --build -d
	CGO_ENABLED=0 GOOS=linux go test -coverpkg ./api/... ./api/app/...
	docker-compose -f docker-compose.test.yml -p test down
	
coverage: ## Run tests and open coverage report of the project
	docker-compose -f docker-compose.test.yml -p test up --build -d
	CGO_ENABLED=0 GOOS=linux go test -coverprofile ./api/cover.out -coverpkg ./api/... ./api/app/...
	go tool cover -html ./api/cover.out -o ./api/cover.html
	open ./api/cover.html
	docker-compose -f docker-compose.test.yml -p test down

perf: ## Run performance tests
	docker-compose -f ./load/docker-compose.yaml up -d --build grafana influxdb
	docker-compose -f ./load/docker-compose.yaml run k6 run /scripts/k6.js 

## Docs:
swagger: ## Generates api docs
	swag init -g api/app/main.go -o api/app/docs

## Kubernetes:
deploy: ## Deploys on minikube
	minikube addons enable metrics-server
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.7.0/aio/deploy/recommended.yaml
	kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml
	kubectl apply -f https://github.com/rabbitmq/cluster-operator/releases/latest/download/cluster-operator.yml
	kubectl apply -f kubernetes/kubernetes-dashboard.yaml
	docker build . -t vitorbiten/maintenance-api
	sed -i'' -e 's/Always/Never/g' ./kubernetes/api.yaml
	kubectl apply -f kubernetes/api.yaml
	kubectl apply -f kubernetes/api-hpa.yaml
	sed -i'' -e 's/Never/Always/g' ./kubernetes/api.yaml
	docker build . -f Dockerfile.worker -t vitorbiten/maintenance-worker
	sed -i'' -e 's/Always/Never/g' ./kubernetes/worker.yaml
	kubectl apply -f kubernetes/worker.yaml
	kubectl apply -f kubernetes/worker-hpa.yaml
	sed -i'' -e 's/Never/Always/g' ./kubernetes/worker.yaml
	kubectl apply -f kubernetes/mysql-configmap.yaml
	kubectl apply -f kubernetes/mysql.yaml
	kubectl apply -f kubernetes/secrets.yaml
	kubectl apply -f kubernetes/rabbitmq.yaml

dashboard: ## Generates a token and opens docker dashboard
	@echo 'URL:'
	@echo 'http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/'
	@echo ''
	@echo 'Token:'
	kubectl -n kubernetes-dashboard create token admin-user
	@echo ''
	kubectl proxy

api-url: ## Get url for api service
	minikube service maintenance-api --url

mysql-url: ## Get url for mysql service
	minikube service mysql --url

rabbitmq: ## Forwards rabbitmq port to localhost
	kubectl port-forward "service/rabbitmq" 15672

kill-dashboard: ## Kills the dashboard proxy process
	pkill -9 -f "kubectl proxy"

## Help:
help: ## Show this help.
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    %-20s%s\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  %s\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)