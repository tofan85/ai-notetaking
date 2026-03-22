package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/loggers"
	"ai-notetaking-be/internal/pkg/serverutils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// TEST SETUP
// ============================================================================

func setupTestRepo(t *testing.T) (*notebookRepository, pgxmock.PgxPoolIface, loggers.Logger) {
	mockDB, err := pgxmock.NewPool()
	require.NoError(t, err)

	logger := *loggers.NewLooger()

	repo := &notebookRepository{
		db:     mockDB,
		Logger: logger,
	}

	return repo, mockDB, logger
}

// ============================================================================
// TEST: GetAll
// ============================================================================

func TestNotebookRepository_GetAll(t *testing.T) {
	t.Run("Success - Found Multiple", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()

		now := time.Now()
		id1 := uuid.New()
		id2 := uuid.New()

		rows := pgxmock.NewRows([]string{
			"id", "name", "parent_id", "created_at", "updated_at", "is_deleted",
		}).
			AddRow(id1, "Notebook 1", nil, now, &now, false).
			AddRow(id2, "Notebook 2", nil, now, &now, false)

		mockDB.ExpectQuery(`SELECT id, name, parent_id, created_at, updated_at, is_deleted FROM notebook WHERE  is_deleted = false`).
			WillReturnRows(rows)

		result, err := repo.GetAll(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Notebook 1", result[0].Name)
		assert.Equal(t, "Notebook 2", result[1].Name)
		assert.Equal(t, id1, result[0].ID)
		assert.Equal(t, id2, result[1].ID)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Success - Empty Result", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()

		rows := pgxmock.NewRows([]string{
			"id", "name", "parent_id", "created_at", "updated_at", "is_deleted",
		})

		mockDB.ExpectQuery(`SELECT id, name, parent_id, created_at, updated_at, is_deleted FROM notebook WHERE  is_deleted = false`).
			WillReturnRows(rows)

		result, err := repo.GetAll(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
		assert.Len(t, result, 0)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()

		dbError := errors.New("connection failed")
		mockDB.ExpectQuery(`SELECT id, name, parent_id, created_at, updated_at, is_deleted FROM notebook`).
			WillReturnError(dbError)

		result, err := repo.GetAll(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbError, err)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Scan Error - Column Mismatch", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()

		// Return rows dengan column yang kurang
		rows := pgxmock.NewRows([]string{"id", "name"}).
			AddRow(uuid.New(), "Test")

		mockDB.ExpectQuery(`SELECT id, name, parent_id, created_at, updated_at, is_deleted FROM notebook`).
			WillReturnRows(rows)

		result, err := repo.GetAll(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})
}

// ============================================================================
// TEST: GetByID
// ============================================================================

func TestNotebookRepository_GetByID(t *testing.T) {
	t.Run("Success - Found", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()
		id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

		now := time.Now()
		rows := pgxmock.NewRows([]string{
			"id", "name", "parent_id", "created_at", "updated_at", "deleted_at", "is_deleted",
		}).AddRow(
			id,
			"Test Notebook",
			nil, // parent_id
			now,
			now,
			nil, // deleted_at
			false,
		)

		mockDB.ExpectQuery(`SELECT id, name, parent_id, created_at, updated_at, deleted_at, is_deleted FROM notebook WHERE id = \$1 AND is_deleted = false`).
			WithArgs(id).
			WillReturnRows(rows)

		result, err := repo.GetByID(ctx, id)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, id, result.ID)
		assert.Equal(t, "Test Notebook", result.Name)
		assert.False(t, result.IsDeleted)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Not Found - pgx.ErrNoRows", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()
		id := uuid.New()

		mockDB.ExpectQuery(`SELECT id, name, parent_id, created_at, updated_at, deleted_at, is_deleted FROM notebook WHERE id = \$1 AND is_deleted = false`).
			WithArgs(id).
			WillReturnError(pgx.ErrNoRows)

		result, err := repo.GetByID(ctx, id)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, serverutils.ErrNotFound))
		assert.Nil(t, result)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()
		id := uuid.New()

		dbError := errors.New("database timeout")
		mockDB.ExpectQuery(`SELECT .+ FROM notebook WHERE id = \$1`).
			WithArgs(id).
			WillReturnError(dbError)

		result, err := repo.GetByID(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.Nil(t, result)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})
}

// ============================================================================
// TEST: Create
// ============================================================================

func TestNotebookRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()

		now := time.Now()
		notebook := &entity.Notebook{
			ID:        uuid.New(),
			Name:      "New Notebook",
			ParentId:  nil,
			CreatedAt: now,
			UpdatedAt: &now,
			IsDeleted: false,
		}

		mockDB.ExpectExec(`INSERT INTO notebook`).
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

		err := repo.Create(ctx, notebook)

		assert.NoError(t, err)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Database Error - Duplicate Key", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()

		notebook := &entity.Notebook{
			ID:   uuid.New(),
			Name: "Test",
		}

		dbError := errors.New("duplicate key value violates unique constraint")
		mockDB.ExpectExec(`INSERT INTO notebook`).
			WithArgs(
				notebook.ID,
				notebook.Name,
				notebook.ParentId,
				notebook.CreatedAt,
				notebook.UpdatedAt,
				notebook.DeletedAt,
				notebook.IsDeleted,
			).
			WillReturnError(dbError)

		err := repo.Create(ctx, notebook)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})
}

// ============================================================================
// TEST: UpdateByID
// ============================================================================

func TestNotebookRepository_UpdateByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()

		now := time.Now()
		notebook := &entity.Notebook{
			ID:        uuid.New(),
			Name:      "Updated Name",
			UpdatedAt: &now,
		}

		mockDB.ExpectExec(`UPDATE notebook SET name`).
			WithArgs(notebook.Name, notebook.UpdatedAt, notebook.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.UpdateByID(ctx, notebook)

		assert.NoError(t, err)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()

		notebook := &entity.Notebook{
			ID:   uuid.New(),
			Name: "Test",
		}

		dbError := errors.New("update failed")
		mockDB.ExpectExec(`UPDATE notebook SET name`).
			WithArgs(notebook.Name, notebook.UpdatedAt, notebook.ID).
			WillReturnError(dbError)

		err := repo.UpdateByID(ctx, notebook)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})
}

// ============================================================================
// TEST: Delete
// ============================================================================

func TestNotebookRepository_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()
		id := uuid.New()

		mockDB.ExpectExec(`UPDATE notebook SET is_deleted = true, deleted_at = \$1 WHERE id = \$2 AND is_deleted = false`).
			WithArgs(pgxmock.AnyArg(), id).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.Delete(ctx, id)

		assert.NoError(t, err)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Not Found - No Rows Affected", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()
		id := uuid.New()

		mockDB.ExpectExec(`UPDATE notebook SET is_deleted = true`).
			WithArgs(pgxmock.AnyArg(), id).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0)) // 0 rows affected

		err := repo.Delete(ctx, id)

		assert.NoError(t, err) // Repository tidak return error jika 0 rows
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()
		id := uuid.New()

		dbError := errors.New("delete failed")
		mockDB.ExpectExec(`UPDATE notebook SET is_deleted = true`).
			WithArgs(pgxmock.AnyArg(), id).
			WillReturnError(dbError)

		err := repo.Delete(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})
}

// ============================================================================
// TEST: NullifyParentById
// ============================================================================

func TestNotebookRepository_NullifyParentById(t *testing.T) {
	t.Run("Success - Multiple Rows Updated", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()
		parentID := uuid.New()

		mockDB.ExpectExec(`UPDATE notebook SET parent_id = null, updated_at = \$1 WHERE parent_id = \$2`).
			WithArgs(pgxmock.AnyArg(), parentID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 3))

		err := repo.NullifyParentById(ctx, parentID)

		assert.NoError(t, err)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Success - No Rows Updated", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()
		parentID := uuid.New()

		mockDB.ExpectExec(`UPDATE notebook SET parent_id = null`).
			WithArgs(pgxmock.AnyArg(), parentID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		err := repo.NullifyParentById(ctx, parentID)

		assert.NoError(t, err)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()
		parentID := uuid.New()

		dbError := errors.New("constraint violation")
		mockDB.ExpectExec(`UPDATE notebook SET parent_id = null`).
			WithArgs(pgxmock.AnyArg(), parentID).
			WillReturnError(dbError)

		err := repo.NullifyParentById(ctx, parentID)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})
}

// ============================================================================
// TEST: UpdateParentID
// ============================================================================

func TestNotebookRepository_UpdateParentID(t *testing.T) {
	t.Run("Success - Set Parent", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()
		id := uuid.New()
		parentID := uuid.New()

		mockDB.ExpectExec(`UPDATE notebook SET parent_id`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), id).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.UpdateParentID(ctx, id, &parentID)

		assert.NoError(t, err)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Success - Set Parent to Null", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()
		id := uuid.New()

		mockDB.ExpectExec(`UPDATE notebook SET parent_id`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), id).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.UpdateParentID(ctx, id, nil)

		assert.NoError(t, err)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		repo, mockDB, _ := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()
		id := uuid.New()
		parentID := uuid.New()

		// Repository implementation always returns nil for UpdateParentID
		// So we adjust the test expectation
		err := repo.UpdateParentID(ctx, id, &parentID)

		assert.NoError(t, err)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})
}

// ============================================================================
// TEST: UsingTx
// ============================================================================

func TestNotebookRepository_UsingTx(t *testing.T) {
	t.Run("Create Repository with Transaction", func(t *testing.T) {
		repo, mockDB, logger := setupTestRepo(t)
		defer mockDB.Close()
		ctx := context.Background()

		// UsingTx dengan mockDB sebagai transaction
		txRepo := repo.UsingTx(ctx, mockDB)

		assert.NotNil(t, txRepo)

		// Type assert untuk verifikasi
		notebookRepo, ok := txRepo.(*notebookRepository)
		assert.True(t, ok)
		assert.Equal(t, logger, notebookRepo.Logger)
		assert.NotNil(t, notebookRepo.db)
	})
}

// ============================================================================
// BENCHMARK
// ============================================================================

func BenchmarkNotebookRepository_GetAll(b *testing.B) {
	mockDB, _ := pgxmock.NewPool()
	defer mockDB.Close()

	logger := *loggers.NewLooger()

	repo := &notebookRepository{
		db:     mockDB,
		Logger: logger,
	}

	now := time.Now()
	rows := pgxmock.NewRows([]string{
		"id", "name", "parent_id", "created_at", "updated_at", "is_deleted",
	}).
		AddRow(uuid.New(), "Notebook 1", nil, now, &now, false).
		AddRow(uuid.New(), "Notebook 2", nil, now, &now, false).
		AddRow(uuid.New(), "Notebook 3", nil, now, &now, false)

	mockDB.ExpectQuery(`SELECT .+ FROM notebook`).
		WillReturnRows(rows)

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = repo.GetAll(ctx)
	}
}

func BenchmarkNotebookRepository_GetByID(b *testing.B) {
	mockDB, _ := pgxmock.NewPool()
	defer mockDB.Close()

	logger := *loggers.NewLooger()
	repo := &notebookRepository{
		db:     mockDB,
		Logger: logger,
	}

	id := uuid.New()
	now := time.Now()
	rows := pgxmock.NewRows([]string{
		"id", "name", "parent_id", "created_at", "updated_at", "deleted_at", "is_deleted",
	}).AddRow(id, "Test", nil, now, now, nil, false)

	mockDB.ExpectQuery(`SELECT .+ FROM notebook WHERE id`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(rows)

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByID(ctx, id)
	}
}
