
.PHONY: migrate migrate-down migrate-status migrate-reset migrate-check migrate-create
run:
	go run cmd/server/main.go -command=up
# Run all pending migrations
migrate:
	go run cmd/migrate/main.go -command=up

# Rollback the last migration
migrate-down:
	go run cmd/migrate/main.go -command=down

# Show migration status
migrate-status:
	go run cmd/migrate/main.go -command=status

# Reset database (WARNING: deletes all data)
migrate-reset:
	go run cmd/migrate/main.go -command=reset

# Check database status
migrate-check:
	go run cmd/migrate/main.go -command=check

# Create a new migration template
migrate-create:
	@read -p "Enter migration name: " name; \
	go run cmd/migrate/main.go -command=create -name=$$name