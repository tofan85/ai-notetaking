package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/interfaces"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/pkg/database"
	"context"
	"errors"

	"ai-notetaking-be/internal/helpers"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type notebookRepository struct {
	db database.DatabaseQueryer
}

func NewNotebookRepository(db *pgxpool.Pool) interfaces.INotebookRepository {
	return &notebookRepository{
		db: db,
	}
}

func (n *notebookRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) interfaces.INotebookRepository {
	return &notebookRepository{
		db: tx,
	}
}

func (n *notebookRepository) Create(ctx context.Context, notebook *entity.Notebook) error {
	_, err := n.db.Exec(
		ctx,
		`INSERT INTO notebook (id, name, parent_id, created_at, updated_at, deleted_at, is_deleted) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		notebook.ID,
		notebook.Name,
		notebook.ParentId,
		notebook.CreatedAt,
		notebook.UpdatedAt,
		notebook.DeletedAt,
		notebook.IsDeleted,
	)

	if err != nil {
		return err
	}

	return nil
}

func (n *notebookRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Notebook, error) {
	var notebook entity.Notebook
	err := n.db.QueryRow(
		ctx,
		`SELECT id, name, parent_id, created_at, updated_at, deleted_at, is_deleted FROM notebook WHERE id = $1 AND is_deleted = false`,
		id,
	).Scan(
		&notebook.ID,
		&notebook.Name,
		&notebook.ParentId,
		&notebook.CreatedAt,
		&notebook.UpdatedAt,
		&notebook.DeletedAt,
		&notebook.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helpers.Logger.Warn("Notebook not found - ID: %s", id)
			return nil, serverutils.ErrNotFound
		}
		return nil, err
	}
	return &notebook, nil
}

func (n *notebookRepository) UpdateByID(ctx context.Context, notebook *entity.Notebook) error {
	_, err := n.db.Exec(
		ctx,
		`UPDATE notebook SET name = $1, updated_at = $2 WHERE id = $3 AND is_deleted = false`,
		notebook.Name,
		notebook.UpdatedAt,
		notebook.ID,
	)
	if err != nil {
		return err
	}
	return nil
}
