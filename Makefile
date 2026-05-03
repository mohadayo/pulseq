.PHONY: up down test test-gateway test-worker test-dashboard lint build

up:
	docker compose up --build -d

down:
	docker compose down

build:
	docker compose build

test: test-gateway test-worker test-dashboard

test-gateway:
	cd gateway && pip install -r requirements.txt -q && pytest -v

test-worker:
	cd worker && go test ./... -v

test-dashboard:
	cd dashboard && npm ci --silent && npm test

lint: lint-gateway lint-worker lint-dashboard

lint-gateway:
	cd gateway && flake8 app.py test_app.py --max-line-length=120

lint-worker:
	cd worker && go vet ./...

lint-dashboard:
	cd dashboard && npm run lint

logs:
	docker compose logs -f

health:
	@echo "Gateway:" && curl -s http://localhost:5000/health | python3 -m json.tool
	@echo "Worker:" && curl -s http://localhost:5001/health | python3 -m json.tool
	@echo "Dashboard:" && curl -s http://localhost:5002/health | python3 -m json.tool
