FROM golang:1.25.6-bookworm AS builder

WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

RUN CGO_ENABLED=0 go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/repocompass-server ./cmd/server

FROM alpine/git:2.49.1 AS runtime

WORKDIR /app

COPY --from=builder /out/repocompass-server /usr/local/bin/repocompass-server
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY backend/db/migrations ./db/migrations
COPY backend/testdata ./testdata
COPY backend/scripts/docker/server-entrypoint.sh /usr/local/bin/server-entrypoint.sh

RUN chmod +x /usr/local/bin/server-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["server-entrypoint.sh"]
