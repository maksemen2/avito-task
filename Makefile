EXECUTABLE_NAME=merch_store

build:
	go build -o ${EXECUTABLE_NAME} cmd/main.go

run: build
	./${EXECUTABLE_NAME}

clean:
	go clean
	rm ${EXECUTABLE_NAME}

deploy:
	docker compose up --build

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

unit-tests:
	go test -v --cover ./internal/dao

e2e-tests:
	go test -v --cover ./internal/handlers