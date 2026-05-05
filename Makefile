.PHONY: help fmt vet test test-postgres server migrate-up migrate-down migrate-status db-up db-down db-reset db-seed db-status

help:
	@printf "Available targets:\n"
	@printf "  make fmt\n"
	@printf "  make vet\n"
	@printf "  make test\n"
	@printf "  make test-postgres\n"
	@printf "  make server\n"
	@printf "  make migrate-up\n"
	@printf "  make migrate-down\n"
	@printf "  make migrate-status\n"
	@printf "  make db-up\n"
	@printf "  make db-down\n"
	@printf "  make db-reset\n"
	@printf "  make db-seed\n"
	@printf "  make db-status\n"

fmt:
	./backend/scripts/dev/fmt.sh

vet:
	./backend/scripts/dev/vet.sh

test:
	./backend/scripts/dev/test.sh

test-postgres:
	sh ./backend/scripts/dev/test-postgres.sh

server:
	set -a; [ ! -f backend/.env ] || . backend/.env; set +a; cd backend && go run ./cmd/server

migrate-up:
	./backend/scripts/dev/migrate-up.sh

migrate-down:
	./backend/scripts/dev/migrate-down.sh

migrate-status:
	./backend/scripts/dev/migrate-status.sh

db-up:
	./backend/scripts/dev/db-up.sh

db-down:
	./backend/scripts/dev/db-down.sh

db-reset:
	./backend/scripts/dev/db-reset.sh

db-seed:
	./backend/scripts/dev/db-seed.sh

db-status:
	./backend/scripts/dev/db-status.sh
