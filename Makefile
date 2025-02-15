EXECUTABLE_NAME=merch_store

build:
	go build -o ${EXECUTABLE_NAME} cmd/main.go

run: build
	./${EXECUTABLE_NAME}

clean:
	go clean
	rm ${EXECUTABLE_NAME}

migration:
	go run cmd/migration/main.go

deploy:
	docker compose up --build

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix