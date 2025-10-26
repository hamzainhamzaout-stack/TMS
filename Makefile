MIGRATIONS_DIR = ./migrations

build:
	docker compose up --build -d

down:
	docker compose down -v

l:
	echo $(s)
	docker compose logs -f $(s)

gs:
	goose -s -dir $(MIGRATIONS_DIR) create $(message) sql
