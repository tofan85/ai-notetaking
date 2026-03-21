package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateNotebookRequest struct {
	Name     string     `json:"name" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id"`
}

type CreateNotebookResponse struct {
	ID uuid.UUID `json:"id"`
}

type ShowNotebookResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	CreateDAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
