package models

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"

	"github.com/phoenix-industries/omega-server/pkg/database"
)

type Media struct {
	Model
	Name string `json:"name" db:"name"`
	Type string `json:"type" db:"type"`
	Hash string `json:"hash" db:"hash"`
}

func (m *Media) Validate() error {
	if m.ID == "" {
		return errors.New("id is required")
	}
	if m.Name == "" {
		return errors.New("name is required")
	}
	if m.Type == "" {
		return errors.New("type is required")
	}
	if m.Hash == "" {
		return errors.New("hash is required")
	}
	return nil
}

func MediaInsert(ctx context.Context, db database.DB, m *Media) error {
	if err := m.Validate(); err != nil {
		return err
	}
	query := `
		INSERT INTO media
		(id, name, type, hash)
		VALUES
		($1, $2, $3, $4)
	`
	_, err := db.Exec(ctx, query, m.ID, m.Name, m.Type, m.Hash)
	if err != nil {
		return err
	}
	return nil
}

func MediaUpdate(ctx context.Context, db database.DB, m *Media) error {
	if err := m.Validate(); err != nil {
		return err
	}
	query := `
		UPDATE media
		SET name = $2, type = $3, hash = $4
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := db.Exec(ctx, query, m.ID, m.Name, m.Type, m.Hash)
	if err != nil {
		return err
	}
	return nil
}

func MediaGetByID(ctx context.Context, db database.DB, id string) (*Media, error) {
	query := `
		SELECT *
		FROM media
		WHERE id = $1 AND deleted_at IS NULL
	`
	var media Media
	err := pgxscan.Get(ctx, db, &media, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}

func MediaGetByHash(ctx context.Context, db database.DB, hash string) (*Media, error) {
	query := `
		SELECT *
		FROM media
		WHERE hash = $1 AND deleted_at IS NULL
	`
	var media Media
	err := pgxscan.Get(ctx, db, &media, query, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}
