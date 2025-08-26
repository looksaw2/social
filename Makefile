export PATH := $(PATH):/home/user/go/bin

.PHONY: migrate_up
migrate_up:
	@migrate -path=./cmd/migrate/migrations -database="postgres://postgres:postgres@localhost:5432/social?sslmode=disable" up

.PHONY: migrate_down
migrate_down:
	@migrate -path=./cmd/migrate/migrations -database="postgres://postgres:postgres@localhost:5432/social?sslmode=disable" down

.PHONY: migrate_create
migrate_create:
	@if [ -z "$(name)" ]; then \
		echo "please input the migration file name"; \
		exit 1; \
	fi
	@migrate create -seq -ext=.sql -dir=./cmd/migrate/migrations $(name);

.PHONY: migrate_force
migrate_force:
	@if [ -z "$(version)" ]; then \
		read -p "Please enter the version you want to force: " version; \
	fi
	@migrate -path=./cmd/migrate/migrations -database="postgres://postgres:postgres@localhost:5432/social?sslmode=disable" force $(version)

.PHONY: seed
seed:
	@go run ./cmd/migrate/seed/main.go


.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt 


"e23aff79-5b92-479d-a67b-c23b5e54caf9"