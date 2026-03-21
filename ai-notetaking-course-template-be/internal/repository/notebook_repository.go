package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/interfaces"
	"ai-notetaking-be/internal/loggers"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/pkg/database"
	"context"
	"errors"
	"log"
	"time"

	"ai-notetaking-be/internal/helpers"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type notebookRepository struct {
	db     database.DatabaseQueryer
	Logger loggers.Logger
}

func NewNotebookRepository(db *pgxpool.Pool, logger loggers.Logger) interfaces.INotebookRepository {
	return &notebookRepository{
		db:     db,
		Logger: logger,
	}
}

func (n *notebookRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) interfaces.INotebookRepository {
	return &notebookRepository{
		db:     tx,
		Logger: n.Logger,
	}
}

func (n *notebookRepository) Create(ctx context.Context, notebook *entity.Notebook) error {

	start := time.Now()
	memBefore := helpers.TrackMemory()

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
	defer helpers.LogExecution(
		n.Logger,
		start,
		&err,
		memBefore,
		"Repository: CreateNotebook",
		map[string]interface{}{
			"notebook_id": notebook.ID,
		},
	)
	log.Printf("[REPOSITORY] CreateNotebook - SUCCESS | id=%s", notebook.ID)
	return nil
}

func (n *notebookRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Notebook, error) {
	start := time.Now()
	memBefore := helpers.TrackMemory()
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
	defer helpers.LogExecution(
		n.Logger,
		start,
		&err,
		memBefore,
		"Repository: Get Notebook BY ID",
		map[string]interface{}{
			"notebook_id": notebook.ID,
		},
	)
	return &notebook, nil
}

func (n *notebookRepository) UpdateByID(ctx context.Context, notebook *entity.Notebook) error {
	start := time.Now()
	memBefore := helpers.TrackMemory()
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
	defer helpers.LogExecution(
		n.Logger,
		start,
		&err,
		memBefore,
		"Repository: Update Notebook BY ID",
		map[string]interface{}{
			"notebook_id": notebook.ID,
		},
	)
	return nil
}
