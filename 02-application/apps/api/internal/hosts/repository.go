package hosts

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/soumabali/vexa/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, host *models.Host) error {
	query := `
		INSERT INTO hosts (id, owner_id, name, address, protocol, port, group_path, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		host.ID.String(), host.OwnerID.String(), host.Name, host.Address, string(host.Protocol), host.Port,
		host.GroupPath, host.Description, host.IsActive,
	)
	if err != nil {
		return fmt.Errorf("failed to create host: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) (*models.Host, error) {
	query := `
		SELECT id, name, address, protocol, port, is_active, created_at, updated_at
		FROM hosts WHERE id = $1 AND owner_id = $2
	`
	var host models.Host
	var createdAt, updatedAt string
	var isActive bool

	err := r.db.QueryRowContext(ctx, query, id.String(), ownerID.String()).Scan(
		&host.ID, &host.Name, &host.Address,
		&host.Protocol, &host.Port, &isActive,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("host not found")
		}
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	host.IsActive = isActive
	if isActive {
		host.Status = "active"
	} else {
		host.Status = "inactive"
	}
	return &host, nil
}

func (r *Repository) List(ctx context.Context, ownerID uuid.UUID, tagFilter, groupFilter string, limit, offset int) ([]*models.Host, error) {
	query := `
		SELECT id, owner_id, name, address, protocol, port, is_active, created_at, updated_at
		FROM hosts
		WHERE owner_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, ownerID.String(), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list hosts: %w", err)
	}
	defer rows.Close()

	var hosts []*models.Host
	for rows.Next() {
		var host models.Host
		var ownerIDStr string
		var createdAt, updatedAt string
		var isActive bool
		err := rows.Scan(
			&host.ID, &ownerIDStr, &host.Name, &host.Address, &host.Protocol, &host.Port,
			&isActive, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan host: %w", err)
		}
		host.OwnerID, _ = uuid.Parse(ownerIDStr)
		host.IsActive = isActive
		if isActive {
			host.Status = "active"
		} else {
			host.Status = "inactive"
		}
		hosts = append(hosts, &host)
	}
	return hosts, rows.Err()
}

func (r *Repository) Update(ctx context.Context, host *models.Host, ownerID uuid.UUID) error {
	query := `
		UPDATE hosts SET name = COALESCE($1, name),
						 address = COALESCE($2, address),
						 protocol = COALESCE($3, protocol),
						 port = COALESCE($4, port),
						 group_path = COALESCE($5, group_path),
						 description = COALESCE($6, description),
						 is_active = COALESCE($7, is_active),
						 updated_at = NOW()
		WHERE id = $8 AND owner_id = $9
	`

	var name, address, protocol interface{}
	var port interface{}
	var groupPath, description interface{}
	var isActive interface{}

	if host.Name != "" {
		name = host.Name
	}
	if host.Address != "" {
		address = host.Address
	}
	if host.Protocol != "" {
		protocol = host.Protocol
	}
	if host.Port > 0 {
		port = host.Port
	}
	if host.GroupPath != "" {
		groupPath = host.GroupPath
	}
	if host.Description != "" {
		description = host.Description
	}
	if host.Status != "" {
		isActive = host.Status == "active"
	}

	_, err := r.db.ExecContext(ctx, query, name, address, protocol, port, groupPath, description, isActive, host.ID.String(), ownerID.String())
	if err != nil {
		return fmt.Errorf("failed to update host: %w", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error {
	query := `DELETE FROM hosts WHERE id = $1 AND owner_id = $2`
	_, err := r.db.ExecContext(ctx, query, id.String(), ownerID.String())
	if err != nil {
		return fmt.Errorf("failed to delete host: %w", err)
	}
	return nil
}
