package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestNotebookRepository_Show_DB(t *testing.T) {
	// Setup mock database
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	// Initialize repository dengan mock
	repo := &notebookRepository{
		db: mock,
	}

	// Prepare test data
	now := time.Now()
	id := uuid.New()
	ctx := context.Background()
	expectedNotebook := &entity.Notebook{
		ID:        id,
		Name:      "Test Notebook",
		ParentId:  nil, // root level
		CreatedAt: now,
		UpdatedAt: &now,
		DeletedAt: nil,
		IsDeleted: false,
	}

	// Setup mock expectation
	mock.ExpectQuery(`SELECT id, name, parent_id, created_at, updated_at, deleted_at, is_deleted FROM notebook WHERE id = \$1 AND is_deleted = false`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "parent_id", "created_at", "updated_at", "deleted_at", "is_deleted"}).
			AddRow(id, "Test Notebook", nil, now, &now, nil, false))

	// 🎯 EXECUTE: Panggil method yang di-test
	result, err := repo.GetByID(ctx, id) // atau GetByID, sesuaikan nama method Anda

	// ✅ ASSERTIONS
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedNotebook.ID, result.ID)
	assert.Equal(t, expectedNotebook.Name, result.Name)
	assert.Equal(t, expectedNotebook.ParentId, result.ParentId)
	assert.Equal(t, expectedNotebook.IsDeleted, result.IsDeleted)

	// ✅ Verifikasi semua mock expectation terpenuhi
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// Test case: Data tidak ditemukan
func TestNotebookRepository_Show_NotFound_DB(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &notebookRepository{db: mock}
	id := uuid.New()
	ctx := context.Background()

	// Return empty rows = not found
	mock.ExpectQuery(`SELECT id, name, parent_id, created_at, updated_at, deleted_at, is_deleted FROM notebook WHERE id = \$1 AND is_deleted = false`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "parent_id", "created_at", "updated_at", "deleted_at", "is_deleted"}))

	// Execute
	result, err := repo.GetByID(ctx, id)

	// Assert
	assert.Error(t, err)
	// assert.True(t, errors.Is(err, ErrNotFound)) // jika Anda punya error sentinel
	assert.Nil(t, result)

	mock.ExpectationsWereMet()
}

// Test case: Database error
func TestNotebookRepository_Show_DBError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &notebookRepository{db: mock}
	id := uuid.New()
	ctx := context.Background()

	// Simulate database error
	mock.ExpectQuery(`SELECT id, name, parent_id, created_at, updated_at, deleted_at, is_deleted FROM notebook WHERE id = \$1 AND is_deleted = false`).
		WithArgs(id).
		WillReturnError(context.DeadlineExceeded) // atau errors.New("connection refused")

	// Execute
	result, err := repo.GetByID(ctx, id)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	mock.ExpectationsWereMet()
}
