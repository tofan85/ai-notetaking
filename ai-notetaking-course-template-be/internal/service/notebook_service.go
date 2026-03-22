package service

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/interfaces"
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.elastic.co/apm"
)

type notebookService struct {
	notebookRepository interfaces.INotebookRepository
	db                 *pgxpool.Pool
}

func NewNotebookService(notebookRepository interfaces.INotebookRepository, db *pgxpool.Pool) interfaces.INotebookService {
	return &notebookService{
		notebookRepository: notebookRepository,
		db:                 db,
	}
}

func (c *notebookService) GetAll(ctx context.Context) ([]*dto.GetAllNotebookResponse, error) {
	span, spanCtx := apm.StartSpan(ctx, "CreateNotebook", "Service")
	defer span.End()

	notebooks, err := c.notebookRepository.GetAll(spanCtx)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.GetAllNotebookResponse, 0)
	for _, notebook := range notebooks {
		res := dto.GetAllNotebookResponse{
			ID:        notebook.ID,
			Name:      notebook.Name,
			ParentID:  notebook.ParentId,
			CreateDAt: notebook.CreatedAt,
		}

		result = append(result, &res)
	}
	log.Printf("[SERVICE] GetAll - SUCCESS ")
	return result, nil
}

func (c *notebookService) CreateNotebook(ctx context.Context, req *dto.CreateNotebookRequest) (*dto.CreateNotebookResponse, error) {
	span, spanCtx := apm.StartSpan(ctx, "CreateNotebook", "Service")
	defer span.End()

	notebook := entity.Notebook{
		ID:        uuid.New(),
		Name:      req.Name,
		ParentId:  req.ParentID,
		CreatedAt: time.Now(),
	}
	err := c.notebookRepository.Create(spanCtx, &notebook)
	if err != nil {
		return nil, err
	}
	log.Printf("[SERVICE] CreateNotebook - SUCCESS | id=%s", notebook.ID)
	return &dto.CreateNotebookResponse{
		ID: notebook.ID,
	}, nil
}

func (c *notebookService) Show(ctx context.Context, id uuid.UUID) (*dto.ShowNotebookResponse, error) {
	span, spanCtx := apm.StartSpan(ctx, "ShowNotebook", "Service")
	defer span.End()
	notebook, err := c.notebookRepository.GetByID(spanCtx, id)
	if err != nil {
		return nil, err
	}

	return &dto.ShowNotebookResponse{
		ID:        notebook.ID,
		Name:      notebook.Name,
		ParentID:  notebook.ParentId,
		CreateDAt: notebook.CreatedAt,
		UpdatedAt: notebook.UpdatedAt,
	}, nil
}

func (c *notebookService) UpdateNotebook(ctx context.Context, req *dto.UpdateNotebookRequest) (*dto.UpdateNotebookResponse, error) {
	span, spanCtx := apm.StartSpan(ctx, "UpdateNotebook", "Service")
	defer span.End()
	notebook, err := c.notebookRepository.GetByID(spanCtx, req.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	notebook.Name = req.Name
	notebook.UpdatedAt = &now
	err = c.notebookRepository.UpdateByID(spanCtx, notebook)
	if err != nil {
		return nil, err
	}
	return &dto.UpdateNotebookResponse{
		ID:   notebook.ID,
		Name: notebook.Name,
	}, nil
}

func (c *notebookService) Delete(ctx context.Context, id uuid.UUID) error {
	span, spanCtx := apm.StartSpan(ctx, "Delete", "Service")
	defer span.End()
	_, err := c.notebookRepository.GetByID(spanCtx, id)
	if err != nil {
		return err
	}
	tx, err := c.db.BeginTx(spanCtx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(spanCtx)
	noteBookRepo := c.notebookRepository.UsingTx(spanCtx, tx)
	err = noteBookRepo.Delete(spanCtx, id)
	if err != nil {
		return err
	}

	err = noteBookRepo.NullifyParentById(spanCtx, id)

	err = tx.Commit(spanCtx)
	if err != nil {
		return err
	}

	return nil
}

func (c *notebookService) MoveNotebook(ctx context.Context, req *dto.MoveNotebookRequest) (*dto.MoveNotebookResponse, error) {
	span, spanCtx := apm.StartSpan(ctx, "MoveNotebook", "Service")
	defer span.End()
	_, err := c.notebookRepository.GetByID(spanCtx, req.ID)
	if err != nil {
		return nil, err
	}
	if req.ParentID != nil {
		_, err := c.notebookRepository.GetByID(spanCtx, *req.ParentID)
		if err != nil {
			return nil, err
		}
	}
	err = c.notebookRepository.UpdateParentID(spanCtx, req.ID, req.ParentID)
	if err != nil {
		return nil, err
	}

	return &dto.MoveNotebookResponse{
		ID: req.ID,
	}, nil

}
