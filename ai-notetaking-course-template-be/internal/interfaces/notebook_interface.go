package interfaces

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/pkg/database"
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type INotebookRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) INotebookRepository
	Create(ctx context.Context, notebook *entity.Notebook) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Notebook, error)
	UpdateByID(ctx context.Context, notebook *entity.Notebook) error
	Delete(ctx context.Context, id uuid.UUID) error
	NullifyParentById(ctx context.Context, id uuid.UUID) error
	UpdateParentID(ctx context.Context, id uuid.UUID, parentID *uuid.UUID) error
	GetAll(cxt context.Context) ([]*entity.Notebook, error)
}

type INotebookService interface {
	CreateNotebook(ctx context.Context, req *dto.CreateNotebookRequest) (*dto.CreateNotebookResponse, error)
	Show(ctx context.Context, id uuid.UUID) (*dto.ShowNotebookResponse, error)
	UpdateNotebook(ctx context.Context, req *dto.UpdateNotebookRequest) (*dto.UpdateNotebookResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	MoveNotebook(ctx context.Context, req *dto.MoveNotebookRequest) (*dto.MoveNotebookResponse, error)
	GetAll(ctx context.Context) ([]*dto.GetAllNotebookResponse, error)
}

type INotebookController interface {
	RegisterRoutes(r fiber.Router)
	Create(ctx *fiber.Ctx) error
	Show(ctx *fiber.Ctx) error
	Delete(ctx *fiber.Ctx) error
	MoveNotebook(ctx *fiber.Ctx) error
	GetAllRoutes(ctx *fiber.Ctx) error
}
