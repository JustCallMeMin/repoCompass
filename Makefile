.PHONY: help fmt vet test migrate-up migrate-down migrate-status db-up db-down

help:
	@printf "Available targets:\n"
	@printf "  make fmt\n"
	@printf "  make vet\n"
	@printf "  make test\n"
	@printf "  make migrate-up\n"
	@printf "  make migrate-down\n"
	@printf "  make migrate-status\n"
	@printf "  make db-up\n"
	@printf "  make db-down\n"

fmt:
	./backend/scripts/dev/fmt.sh

vet:
	./backend/scripts/dev/vet.sh

test:
	./backend/scripts/dev/test.sh

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
