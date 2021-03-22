build:
	go build main.go

run:
	go run main.go

swag:
    swag init -g cmd/main.go
