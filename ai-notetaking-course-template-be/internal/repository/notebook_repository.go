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
	"go.elastic.co/apm"
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

func (n *notebookRepository) GetAll(ctx context.Context) ([]*entity.Notebook, error) {
	span, _ := apm.StartSpan(ctx, "GetAll", "Repository")
	defer span.End()
	start := time.Now()
	memBefore := helpers.TrackMemory()

	rows, err := n.db.Query(
		ctx,
		`SELECT id, name, parent_id, created_at, updated_at, is_deleted FROM notebook WHERE  is_deleted = false`,
	)

	if err != nil {
		return nil, err
	}

	result := make([]*entity.Notebook, 0)
	for rows.Next() {
		notebook := entity.Notebook{}
		err = rows.Scan(
			&notebook.ID,
			&notebook.Name,
			&notebook.ParentId,
			&notebook.CreatedAt,
			&notebook.UpdatedAt,
			&notebook.IsDeleted,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, &notebook)
	}
	defer helpers.LogExecution(
		n.Logger,
		start,
		&err,
		memBefore,
		"Repository: Get Notebook BY ID",
		map[string]interface{}{},
	)
	return result, nil
}

func (n *notebookRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) interfaces.INotebookRepository {
	return &notebookRepository{
		db:     tx,
		Logger: n.Logger,
	}
}

func (n *notebookRepository) Create(ctx context.Context, notebook *entity.Notebook) error {
	span, _ := apm.StartSpan(ctx, "Create", "Repository")
	defer span.End()
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
	span, _ := apm.StartSpan(ctx, "GetByID", "Repository")
	defer span.End()
	start := time.Now()
	memBefore := helpers.TrackMemory()
	notebook := &entity.Notebook{}
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

	defer func() {
		helpers.LogExecution(
			n.Logger, // atau helpers.Logger - konsisten!
			start,
			&err, // pointer ke err agar defer bisa lihat nilai final
			memBefore,
			"Repository: Get Notebook BY ID",
			map[string]interface{}{
				"notebook_id": id, // ✅ Gunakan id parameter, bukan notebook.ID (bisa nil)
				"found":       err == nil,
			},
		)
	}()

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// ✅ Cek logger nil untuk test
			if helpers.Logger != nil {
				helpers.Logger.Warn("Notebook not found ")
			}
			return nil, serverutils.ErrNotFound
		}
		return nil, err
	}
	return notebook, nil
}

func (n *notebookRepository) UpdateByID(ctx context.Context, notebook *entity.Notebook) error {
	span, _ := apm.StartSpan(ctx, "UpdateByID", "Repository")
	defer span.End()
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

func (n *notebookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	span, _ := apm.StartSpan(ctx, "Delete", "Repository")
	defer span.End()
	start := time.Now()
	memBefore := helpers.TrackMemory()
	_, err := n.db.Exec(
		ctx,
		`UPDATE notebook SET is_deleted = true, deleted_at = $1 WHERE id = $2 AND is_deleted = false`,
		time.Now(),
		id,
	)
	if err != nil {
		return err
	}
	defer helpers.LogExecution(
		n.Logger,
		start,
		&err,
		memBefore,
		"Repository: Delete Notebook",
		map[string]interface{}{
			"notebook_id": id,
		},
	)
	return nil
}

func (n *notebookRepository) NullifyParentById(ctx context.Context, parentID uuid.UUID) error {
	span, _ := apm.StartSpan(ctx, "NullifyParentById", "Repository")
	defer span.End()
	start := time.Now()
	memBefore := helpers.TrackMemory()
	_, err := n.db.Exec(
		ctx,
		`UPDATE notebook SET parent_id = null, updated_at = $1 WHERE parent_id = $2`,
		time.Now(),
		parentID,
	)
	if err != nil {
		return err
	}
	defer helpers.LogExecution(
		n.Logger,
		start,
		&err,
		memBefore,
		"Repository: Nullify Parent ID",
		map[string]interface{}{
			"parent_id": parentID,
		},
	)
	return nil
}

func (n *notebookRepository) UpdateParentID(ctx context.Context, id uuid.UUID, parentID *uuid.UUID) error {
	span, _ := apm.StartSpan(ctx, "UpdateParentID", "Repository")
	defer span.End()
	start := time.Now()
	memBefore := helpers.TrackMemory()
	_, err := n.db.Exec(
		ctx,
		`UPDATE notebook SET parent_id = $1, updated_at = $2 WHERE id = $3 AND is_deleted = false`,
		parentID,
		time.Now(),
		id,
	)
	defer helpers.LogExecution(
		n.Logger,
		start,
		&err,
		memBefore,
		"Repository: Update Parent ID",
		map[string]interface{}{
			"notebook_id": id,
			"parent_id":   parentID,
		},
	)
	return nil
}
