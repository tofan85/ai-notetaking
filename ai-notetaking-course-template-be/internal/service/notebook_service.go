package service

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/interfaces"
	"context"
	"time"

	"github.com/google/uuid"
)

type notebookService struct {
	notebookRepository interfaces.INotebookRepository
}

func NewNotebookService(notebookRepository interfaces.INotebookRepository) interfaces.INotebookService {
	return &notebookService{
		notebookRepository: notebookRepository,
	}
}

func (c *notebookService) CreateNotebook(ctx context.Context, req *dto.CreateNotebookRequest) (*dto.CreateNotebookResponse, error) {

	notebook := entity.Notebook{
		ID:        uuid.New(),
		Name:      req.Name,
		ParentId:  req.ParentID,
		CreatedAt: time.Now(),
	}
	err := c.notebookRepository.Create(ctx, &notebook)
	if err != nil {
		return nil, err
	}

	return &dto.CreateNotebookResponse{
		ID: notebook.ID,
	}, nil
}

func (c *notebookService) Show(ctx context.Context, id uuid.UUID) (*dto.ShowNotebookResponse, error) {
	notebook, err := c.notebookRepository.GetByID(ctx, id)
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
