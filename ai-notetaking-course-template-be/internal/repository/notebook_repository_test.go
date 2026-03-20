package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"

	"ai-notetaking-be/internal/entity"
)

func TestNotebookRepository_Create_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &notebookRepository{
		db: mock,
	}

	now := time.Now()

	notebook := &entity.Notebook{
		ID:        uuid.New(),
		Name:      "Test Notebook",
		ParentId:  nil,
		CreatedAt: now,
		UpdatedAt: &now,
		DeletedAt: nil,
		IsDeleted: false,
	}

	mock.ExpectExec(`INSERT INTO notebook`).
		WithArgs(
			notebook.ID,
			notebook.Name,
			notebook.ParentId,
			notebook.CreatedAt,
			notebook.UpdatedAt,
			notebook.DeletedAt,
			notebook.IsDeleted,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Create(context.Background(), notebook)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotebookRepository_Create_DBError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &notebookRepository{
		db: mock,
	}

	now := time.Now()

	notebook := &entity.Notebook{
		ID:        uuid.New(),
		Name:      "Test Notebook",
		ParentId:  nil,
		CreatedAt: now,
		UpdatedAt: &now,
		DeletedAt: nil,
		IsDeleted: false,
	}

	mock.ExpectExec(`INSERT INTO notebook`).
		WithArgs(
			notebook.ID,
			notebook.Name,
			notebook.ParentId,
			notebook.CreatedAt,
			notebook.UpdatedAt,
			notebook.DeletedAt,
			notebook.IsDeleted,
		).
		WillReturnError(assert.AnError)

	err = repo.Create(context.Background(), notebook)

	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
