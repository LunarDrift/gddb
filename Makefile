DB_USER := postgres
DB_NAME := deadabase
DB_URL := postgres://postgres:postgres@localhost:5432/$(DB_NAME)
BACKUP_DIR := backups
TIMESTAMP := $(shell date +%Y%m%d_%H%M%S)
MIGRATIONS_DIR := sql/schema
PGPASSWORD := postgres

.PHONY: backup backup-schema restore wipe reimport

backup:
	@mkdir -p $(BACKUP_DIR)
	PGPASSWORD=$(PGPASSWORD) pg_dump -h localhost -U $(DB_USER) -d $(DB_NAME) -F c -f $(BACKUP_DIR)/$(DB_NAME)_$(TIMESTAMP).dump
	@echo "Backup written to $(BACKUP_DIR)/$(DB_NAME)_$(TIMESTAMP).dump"

backup-schema:
	@mkdir -p $(BACKUP_DIR)
	PGPASSWORD=$(PGPASSWORD) pg_dump -h localhost -U $(DB_USER) -d $(DB_NAME) --schema-only -f $(BACKUP_DIR)/$(DB_NAME)_schema_$(TIMESTAMP).sql
	@echo "Schema-only backup written to $(BACKUP_DIR)/$(DB_NAME)_schema_$(TIMESTAMP).sql"

restore:
	PGPASSWORD=$(PGPASSWORD) pg_restore -h localhost -U $(DB_USER) -d $(DB_NAME) --clean $(FILE)

wipe:
	@echo "Dropping all tables and re-running migrations..."
	goose postgres -dir $(MIGRATIONS_DIR) "$(DB_URL)" reset
	goose postgres -dir $(MIGRATIONS_DIR) "$(DB_URL)" up
	@echo "Schema reset complete."

reimport: backup wipe
	@echo "Running importer against fresh schema..."
	go run ./cmd/importer
	@echo "Reimport complete."
