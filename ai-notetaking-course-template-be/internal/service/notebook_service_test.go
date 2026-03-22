package service

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/interfaces"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/pkg/database"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// MOCK REPOSITORY IMPLEMENTATION
// ============================================================================

type mockNotebookRepository struct {
	mock.Mock
}

func (m *mockNotebookRepository) GetAll(ctx context.Context) ([]*entity.Notebook, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Notebook), args.Error(1)
}

func (m *mockNotebookRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Notebook, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Notebook), args.Error(1)
}

func (m *mockNotebookRepository) Create(ctx context.Context, notebook *entity.Notebook) error {
	args := m.Called(ctx, notebook)
	return args.Error(0)
}

func (m *mockNotebookRepository) UpdateByID(ctx context.Context, notebook *entity.Notebook) error {
	args := m.Called(ctx, notebook)
	return args.Error(0)
}

func (m *mockNotebookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockNotebookRepository) NullifyParentById(ctx context.Context, parentID uuid.UUID) error {
	args := m.Called(ctx, parentID)
	return args.Error(0)
}

func (m *mockNotebookRepository) UpdateParentID(ctx context.Context, id uuid.UUID, parentID *uuid.UUID) error {
	args := m.Called(ctx, id, parentID)
	return args.Error(0)
}

func (m *mockNotebookRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) interfaces.INotebookRepository {
	args := m.Called(ctx, tx)
	return args.Get(0).(interfaces.INotebookRepository)
}

// ============================================================================
// TEST SETUP
// ============================================================================

func setupTestService(t *testing.T) (*notebookService, *mockNotebookRepository) {
	mockRepo := new(mockNotebookRepository)

	// Silent logger untuk test
	logrus.SetOutput(io.Discard)

	service := &notebookService{
		notebookRepository: mockRepo,
		db:                 nil,
	}

	return service, mockRepo
}

// ============================================================================
// TEST: GetAll
// ============================================================================

func TestNotebookService_GetAll(t *testing.T) {
	t.Run("Success - Multiple Notebooks", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()

		now := time.Now()
		id1 := uuid.New()
		id2 := uuid.New()

		mockNotebooks := []*entity.Notebook{
			{
				ID:        id1,
				Name:      "Notebook 1",
				ParentId:  nil,
				CreatedAt: now,
			},
			{
				ID:        id2,
				Name:      "Notebook 2",
				ParentId:  nil,
				CreatedAt: now,
			},
		}

		mockRepo.On("GetAll", ctx).Return(mockNotebooks, nil).Once()

		result, err := service.GetAll(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Notebook 1", result[0].Name)
		assert.Equal(t, "Notebook 2", result[1].Name)
		assert.Equal(t, id1, result[0].ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Empty Result", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()

		mockRepo.On("GetAll", ctx).Return([]*entity.Notebook{}, nil).Once()

		result, err := service.GetAll(ctx)

		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.Len(t, result, 0)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()

		dbError := errors.New("database connection failed")
		mockRepo.On("GetAll", ctx).Return(nil, dbError).Once()

		result, err := service.GetAll(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbError, err)
		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// TEST: CreateNotebook
// ============================================================================

func TestNotebookService_CreateNotebook(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()

		req := &dto.CreateNotebookRequest{
			Name:     "New Notebook",
			ParentID: nil,
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.Notebook")).
			Return(nil).
			Once()

		result, err := service.CreateNotebook(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEqual(t, uuid.Nil, result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()

		req := &dto.CreateNotebookRequest{
			Name: "New Notebook",
		}

		dbError := errors.New("insert failed")
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.Notebook")).
			Return(dbError).
			Once()

		result, err := service.CreateNotebook(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbError, err)
		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// TEST: Show
// ============================================================================

func TestNotebookService_Show(t *testing.T) {
	t.Run("Success - Found", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()
		id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

		now := time.Now()
		mockNotebook := &entity.Notebook{
			ID:        id,
			Name:      "Test Notebook",
			ParentId:  nil,
			CreatedAt: now,
			UpdatedAt: &now,
		}

		mockRepo.On("GetByID", ctx, id).Return(mockNotebook, nil).Once()

		result, err := service.Show(ctx, id)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, id, result.ID)
		assert.Equal(t, "Test Notebook", result.Name)
		assert.Nil(t, result.ParentID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()
		id := uuid.New()

		mockRepo.On("GetByID", ctx, id).Return(nil, serverutils.ErrNotFound).Once()

		result, err := service.Show(ctx, id)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, serverutils.ErrNotFound))
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()
		id := uuid.New()

		dbError := errors.New("query failed")
		mockRepo.On("GetByID", ctx, id).Return(nil, dbError).Once()

		result, err := service.Show(ctx, id)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbError, err)
		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// TEST: UpdateNotebook
// ============================================================================

func TestNotebookService_UpdateNotebook(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()

		id := uuid.New()
		now := time.Now()

		req := &dto.UpdateNotebookRequest{
			ID:   id,
			Name: "Updated Name",
		}

		existingNotebook := &entity.Notebook{
			ID:        id,
			Name:      "Old Name",
			UpdatedAt: &now,
		}

		mockRepo.On("GetByID", ctx, id).Return(existingNotebook, nil).Once()
		mockRepo.On("UpdateByID", ctx, mock.AnythingOfType("*entity.Notebook")).
			Return(nil).
			Once()

		result, err := service.UpdateNotebook(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, id, result.ID)
		assert.Equal(t, "Updated Name", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Notebook Not Found", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()

		req := &dto.UpdateNotebookRequest{
			ID:   uuid.New(),
			Name: "Updated Name",
		}

		mockRepo.On("GetByID", ctx, req.ID).Return(nil, serverutils.ErrNotFound).Once()

		result, err := service.UpdateNotebook(ctx, req)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, serverutils.ErrNotFound))
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update Error", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()

		id := uuid.New()
		now := time.Now()

		req := &dto.UpdateNotebookRequest{
			ID:   id,
			Name: "Updated Name",
		}

		existingNotebook := &entity.Notebook{
			ID:        id,
			Name:      "Old Name",
			UpdatedAt: &now,
		}

		dbError := errors.New("update failed")
		mockRepo.On("GetByID", ctx, id).Return(existingNotebook, nil).Once()
		mockRepo.On("UpdateByID", ctx, mock.AnythingOfType("*entity.Notebook")).
			Return(dbError).
			Once()

		result, err := service.UpdateNotebook(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbError, err)
		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// TEST: MoveNotebook
// ============================================================================

func TestNotebookService_MoveNotebook(t *testing.T) {
	t.Run("Success - Move to Parent", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()

		notebookID := uuid.New()
		parentID := uuid.New()

		req := &dto.MoveNotebookRequest{
			ID:       notebookID,
			ParentID: &parentID,
		}

		// Mock GetByID untuk notebook
		existingNotebook := &entity.Notebook{
			ID:   notebookID,
			Name: "Notebook to Move",
		}
		mockRepo.On("GetByID", ctx, notebookID).Return(existingNotebook, nil).Once()

		// Mock GetByID untuk parent
		parentNotebook := &entity.Notebook{
			ID:   parentID,
			Name: "Parent Notebook",
		}
		mockRepo.On("GetByID", ctx, parentID).Return(parentNotebook, nil).Once()

		// Mock UpdateParentID
		mockRepo.On("UpdateParentID", ctx, notebookID, &parentID).Return(nil).Once()

		result, err := service.MoveNotebook(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, notebookID, result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Move to Root (Nil Parent)", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()

		notebookID := uuid.New()

		req := &dto.MoveNotebookRequest{
			ID:       notebookID,
			ParentID: nil, // Move to root
		}

		existingNotebook := &entity.Notebook{
			ID:   notebookID,
			Name: "Notebook to Move",
		}
		mockRepo.On("GetByID", ctx, notebookID).Return(existingNotebook, nil).Once()

		// Tidak perlu GetByID untuk parent karena ParentID nil

		// Mock UpdateParentID dengan nil
		mockRepo.On("UpdateParentID", ctx, notebookID, (*uuid.UUID)(nil)).Return(nil).Once()

		result, err := service.MoveNotebook(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, notebookID, result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Notebook Not Found", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()

		notebookID := uuid.New()
		parentID := uuid.New()

		req := &dto.MoveNotebookRequest{
			ID:       notebookID,
			ParentID: &parentID,
		}

		mockRepo.On("GetByID", ctx, notebookID).Return(nil, serverutils.ErrNotFound).Once()

		result, err := service.MoveNotebook(ctx, req)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, serverutils.ErrNotFound))
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Parent Not Found", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()

		notebookID := uuid.New()
		parentID := uuid.New()

		req := &dto.MoveNotebookRequest{
			ID:       notebookID,
			ParentID: &parentID,
		}

		existingNotebook := &entity.Notebook{
			ID:   notebookID,
			Name: "Notebook to Move",
		}
		mockRepo.On("GetByID", ctx, notebookID).Return(existingNotebook, nil).Once()
		mockRepo.On("GetByID", ctx, parentID).Return(nil, serverutils.ErrNotFound).Once()

		result, err := service.MoveNotebook(ctx, req)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, serverutils.ErrNotFound))
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("UpdateParentID Error", func(t *testing.T) {
		service, mockRepo := setupTestService(t)
		ctx := context.Background()

		notebookID := uuid.New()
		parentID := uuid.New()

		req := &dto.MoveNotebookRequest{
			ID:       notebookID,
			ParentID: &parentID,
		}

		existingNotebook := &entity.Notebook{
			ID:   notebookID,
			Name: "Notebook to Move",
		}
		parentNotebook := &entity.Notebook{
			ID:   parentID,
			Name: "Parent",
		}
		mockRepo.On("GetByID", ctx, notebookID).Return(existingNotebook, nil).Once()
		mockRepo.On("GetByID", ctx, parentID).Return(parentNotebook, nil).Once()

		updateError := errors.New("update failed")
		mockRepo.On("UpdateParentID", ctx, notebookID, &parentID).Return(updateError).Once()

		result, err := service.MoveNotebook(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, updateError, err)
		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// TABLE DRIVEN TEST EXAMPLE
// ============================================================================

func TestNotebookService_GetAll_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*mockNotebookRepository)
		expectedLen   int
		expectedError bool
		errorContains string
	}{
		{
			name: "success with data",
			mockSetup: func(m *mockNotebookRepository) {
				now := time.Now()
				m.On("GetAll", mock.Anything).Return([]*entity.Notebook{
					{ID: uuid.New(), Name: "Test 1", CreatedAt: now},
					{ID: uuid.New(), Name: "Test 2", CreatedAt: now},
				}, nil).Once()
			},
			expectedLen:   2,
			expectedError: false,
		},
		{
			name: "success empty",
			mockSetup: func(m *mockNotebookRepository) {
				m.On("GetAll", mock.Anything).Return([]*entity.Notebook{}, nil).Once()
			},
			expectedLen:   0,
			expectedError: false,
		},
		{
			name: "database error",
			mockSetup: func(m *mockNotebookRepository) {
				m.On("GetAll", mock.Anything).Return(nil, errors.New("db error")).Once()
			},
			expectedLen:   0,
			expectedError: true,
			errorContains: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockNotebookRepository)
			tt.mockSetup(mockRepo)

			// Buat service dengan mock repo
			service := &notebookService{
				notebookRepository: mockRepo,
			}

			result, err := service.GetAll(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedLen)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
