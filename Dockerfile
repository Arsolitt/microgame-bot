FROM harbor.arsolitt.tech/hub/golang:1.25.4-bookworm AS local
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GOCACHE=/src/cache
WORKDIR /src/app
RUN go install github.com/air-verse/air@latest
RUN go install github.com/go-task/task/v3/cmd/task@latest
COPY go.mod go.sum ./
RUN go mod download
RUN mkdir -p /src/cache && chown -R 1000:1000 /src/cache
RUN chmod 777 -R /go
EXPOSE 8080
USER 1000
CMD ["air"]

FROM harbor.arsolitt.tech/hub/golang:1.25.4-bookworm AS builder
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app ./cmd/main.go

FROM harbor.arsolitt.tech/hub/bookworm-slim AS production
ENV TZ=UTC
WORKDIR /app
COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"]
