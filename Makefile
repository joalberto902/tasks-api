build: create vet fmt
	go build -o ./bin/tasks-api ./cmd/api

create:
	mkdir bin

vet:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin
