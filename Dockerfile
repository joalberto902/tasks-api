FROM golang:1.27 AS build

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /usr/local/bin/tasks-api ./cmd/api

FROM alpine:3.24

COPY --from=build /usr/local/bin/tasks-api /usr/local/bin/tasks-api

USER nobody

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/tasks-api"]
