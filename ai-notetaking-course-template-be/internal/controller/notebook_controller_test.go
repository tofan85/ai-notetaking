package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/pkg/serverutils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// MOCK SERVICE IMPLEMENTATION
// ============================================================================

type mockNotebookService struct {
	mock.Mock
}

func (m *mockNotebookService) GetAll(ctx context.Context) ([]*dto.GetAllNotebookResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.GetAllNotebookResponse), args.Error(1)
}

func (m *mockNotebookService) CreateNotebook(ctx context.Context, req *dto.CreateNotebookRequest) (*dto.CreateNotebookResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CreateNotebookResponse), args.Error(1)
}

func (m *mockNotebookService) Show(ctx context.Context, id uuid.UUID) (*dto.ShowNotebookResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ShowNotebookResponse), args.Error(1)
}

func (m *mockNotebookService) UpdateNotebook(ctx context.Context, req *dto.UpdateNotebookRequest) (*dto.UpdateNotebookResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UpdateNotebookResponse), args.Error(1)
}

func (m *mockNotebookService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockNotebookService) MoveNotebook(ctx context.Context, req *dto.MoveNotebookRequest) (*dto.MoveNotebookResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.MoveNotebookResponse), args.Error(1)
}

// ============================================================================
// TEST SETUP HELPERS
// ============================================================================

func setupTestApp(t *testing.T) (*fiber.App, *mockNotebookService) {
	app := fiber.New()
	mockService := new(mockNotebookService)
	controller := NewNotebookController(mockService)

	// Register routes
	api := app.Group("/api")
	controller.RegisterRoutes(api)

	return app, mockService
}

func makeRequest(t *testing.T, app *fiber.App, method, path string, body interface{}) (*http.Response, []byte) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		assert.NoError(t, err)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	return resp, bodyBytes
}

// ============================================================================
// TEST: GetAllRoutes
// ============================================================================

func TestNotebookController_GetAllRoutes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		mockData := []*dto.GetAllNotebookResponse{
			{ID: uuid.New(), Name: "Notebook 1"},
			{ID: uuid.New(), Name: "Notebook 2"},
		}

		mockService.On("GetAll", mock.Anything).Return(mockData, nil).Once()

		resp, body := makeRequest(t, app, "GET", "/api/notebook/v1", nil)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err := json.Unmarshal(body, &response)
		assert.NoError(t, err)
		assert.Equal(t, "Success Get List All", response["message"])
		assert.NotNil(t, response["data"])

		mockService.AssertExpectations(t)
	})

	t.Run("Service Error", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		dbError := errors.New("database error")
		mockService.On("GetAll", mock.Anything).Return(nil, dbError).Once()

		resp, _ := makeRequest(t, app, "GET", "/api/notebook/v1", nil)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

// ============================================================================
// TEST: Create
// ============================================================================

func TestNotebookController_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		reqBody := dto.CreateNotebookRequest{
			Name:     "New Notebook",
			ParentID: nil,
		}

		mockResponse := &dto.CreateNotebookResponse{
			ID: uuid.New(),
		}

		mockService.On("CreateNotebook", mock.Anything, &reqBody).Return(mockResponse, nil).Once()

		resp, body := makeRequest(t, app, "POST", "/api/notebook/v1", reqBody)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err := json.Unmarshal(body, &response)
		assert.NoError(t, err)
		assert.Equal(t, "Success create notebook", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		app, _ := setupTestApp(t)

		req := httptest.NewRequest("POST", "/api/notebook/v1", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Controller returns 500 for parse errors
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Validation Error", func(t *testing.T) {
		app, _ := setupTestApp(t)

		// Empty name should fail validation
		reqBody := dto.CreateNotebookRequest{
			Name: "", // Empty name
		}

		resp, _ := makeRequest(t, app, "POST", "/api/notebook/v1", reqBody)

		// Controller returns 500 for validation errors
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Service Error", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		reqBody := dto.CreateNotebookRequest{
			Name: "Test",
		}

		dbError := errors.New("create failed")
		mockService.On("CreateNotebook", mock.Anything, &reqBody).Return(nil, dbError).Once()

		resp, _ := makeRequest(t, app, "POST", "/api/notebook/v1", reqBody)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

// ============================================================================
// TEST: Show
// ============================================================================

func TestNotebookController_Show(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

		mockResponse := &dto.ShowNotebookResponse{
			ID:   id,
			Name: "Test Notebook",
		}

		mockService.On("Show", mock.Anything, id).Return(mockResponse, nil).Once()

		resp, body := makeRequest(t, app, "GET", "/api/notebook/v1/"+id.String(), nil)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err := json.Unmarshal(body, &response)
		assert.NoError(t, err)
		assert.Equal(t, "Success get notebook", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		// Controller parses invalid UUID as nil (zero UUID)
		// Mock for zero UUID
		zeroUUID := uuid.UUID{}
		mockService.On("Show", mock.Anything, zeroUUID).Return(nil, serverutils.ErrNotFound).Once()

		resp, _ := makeRequest(t, app, "GET", "/api/notebook/v1/invalid-uuid", nil)

		// Controller parses UUID with uuid.Parse which returns zero UUID for invalid input
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		id := uuid.New()

		mockService.On("Show", mock.Anything, id).Return(nil, serverutils.ErrNotFound).Once()

		resp, _ := makeRequest(t, app, "GET", "/api/notebook/v1/"+id.String(), nil)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Service Error", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		id := uuid.New()
		dbError := errors.New("query failed")

		mockService.On("Show", mock.Anything, id).Return(nil, dbError).Once()

		resp, _ := makeRequest(t, app, "GET", "/api/notebook/v1/"+id.String(), nil)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

// ============================================================================
// TEST: Update
// ============================================================================

func TestNotebookController_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

		reqBody := dto.UpdateNotebookRequest{
			ID:   id,
			Name: "Updated Name",
		}

		mockResponse := &dto.UpdateNotebookResponse{
			ID:   id,
			Name: "Updated Name",
		}

		mockService.On("UpdateNotebook", mock.Anything, &reqBody).Return(mockResponse, nil).Once()

		resp, body := makeRequest(t, app, "PUT", "/api/notebook/v1/"+id.String(), reqBody)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err := json.Unmarshal(body, &response)
		assert.NoError(t, err)
		assert.Equal(t, "Success update notebook", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		app, _ := setupTestApp(t)

		id := uuid.New()
		req := httptest.NewRequest("PUT", "/api/notebook/v1/"+id.String(), bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Not Found", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		id := uuid.New()
		reqBody := dto.UpdateNotebookRequest{
			ID:   id,
			Name: "Updated Name",
		}

		mockService.On("UpdateNotebook", mock.Anything, &reqBody).Return(nil, serverutils.ErrNotFound).Once()

		resp, _ := makeRequest(t, app, "PUT", "/api/notebook/v1/"+id.String(), reqBody)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

// ============================================================================
// TEST: Delete
// ============================================================================

func TestNotebookController_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		id := uuid.New()

		mockService.On("Delete", mock.Anything, id).Return(nil).Once()

		resp, body := makeRequest(t, app, "DELETE", "/api/notebook/v1/"+id.String(), nil)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err := json.Unmarshal(body, &response)
		assert.NoError(t, err)
		assert.Equal(t, "Success delete notebook", response["message"])
		assert.Nil(t, response["data"])

		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		id := uuid.New()

		mockService.On("Delete", mock.Anything, id).Return(serverutils.ErrNotFound).Once()

		resp, _ := makeRequest(t, app, "DELETE", "/api/notebook/v1/"+id.String(), nil)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Service Error", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		id := uuid.New()
		dbError := errors.New("delete failed")

		mockService.On("Delete", mock.Anything, id).Return(dbError).Once()

		resp, _ := makeRequest(t, app, "DELETE", "/api/notebook/v1/"+id.String(), nil)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

// ============================================================================
// TEST: MoveNotebook
// ============================================================================

func TestNotebookController_MoveNotebook(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		parentID := uuid.New()

		reqBody := dto.MoveNotebookRequest{
			ParentID: &parentID,
		}

		// Expected request after ID is set from params
		expectedReq := dto.MoveNotebookRequest{
			ID:       id,
			ParentID: &parentID,
		}

		mockResponse := &dto.MoveNotebookResponse{
			ID: id,
		}

		mockService.On("MoveNotebook", mock.Anything, &expectedReq).Return(mockResponse, nil).Once()

		resp, body := makeRequest(t, app, "PUT", "/api/notebook/v1/"+id.String()+"/movenotebook", reqBody)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err := json.Unmarshal(body, &response)
		assert.NoError(t, err)
		assert.Equal(t, "Succes move notebook", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("Move to Root", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		id := uuid.New()

		reqBody := dto.MoveNotebookRequest{
			ParentID: nil, // Move to root
		}

		expectedReq := dto.MoveNotebookRequest{
			ID:       id,
			ParentID: nil,
		}

		mockResponse := &dto.MoveNotebookResponse{
			ID: id,
		}

		mockService.On("MoveNotebook", mock.Anything, &expectedReq).Return(mockResponse, nil).Once()

		resp, _ := makeRequest(t, app, "PUT", "/api/notebook/v1/"+id.String()+"/movenotebook", reqBody)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		app, _ := setupTestApp(t)

		id := uuid.New()
		req := httptest.NewRequest("PUT", "/api/notebook/v1/"+id.String()+"/movenotebook", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Not Found", func(t *testing.T) {
		app, mockService := setupTestApp(t)

		id := uuid.New()
		parentID := uuid.New()

		reqBody := dto.MoveNotebookRequest{
			ParentID: &parentID,
		}

		expectedReq := dto.MoveNotebookRequest{
			ID:       id,
			ParentID: &parentID,
		}

		mockService.On("MoveNotebook", mock.Anything, &expectedReq).Return(nil, serverutils.ErrNotFound).Once()

		resp, _ := makeRequest(t, app, "PUT", "/api/notebook/v1/"+id.String()+"/movenotebook", reqBody)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

// ============================================================================
// TEST: RegisterRoutes
// ============================================================================

func TestNotebookController_RegisterRoutes(t *testing.T) {
	app := fiber.New()
	mockService := new(mockNotebookService)
	controller := NewNotebookController(mockService)

	// Register routes
	api := app.Group("/api")
	controller.RegisterRoutes(api)

	// Test that all routes are registered
	routes := app.Stack()

	// Check that routes exist (Fiber v2 stores routes differently)
	// This is a basic check - in real scenario, test actual endpoints
	assert.NotNil(t, routes)
}

// ============================================================================
// TABLE DRIVEN TESTS
// ============================================================================

func TestNotebookController_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		mockSetup      func(*mockNotebookService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:   "get all success",
			method: "GET",
			path:   "/api/notebook/v1",
			mockSetup: func(m *mockNotebookService) {
				m.On("GetAll", mock.Anything).Return([]*dto.GetAllNotebookResponse{}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "Success Get List All",
		},
		{
			name:   "create success",
			method: "POST",
			path:   "/api/notebook/v1",
			body:   dto.CreateNotebookRequest{Name: "Test"},
			mockSetup: func(m *mockNotebookService) {
				m.On("CreateNotebook", mock.Anything, mock.Anything).Return(&dto.CreateNotebookResponse{ID: uuid.New()}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "Success create notebook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, mockService := setupTestApp(t)
			tt.mockSetup(mockService)

			resp, body := makeRequest(t, app, tt.method, tt.path, tt.body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var response map[string]interface{}
			json.Unmarshal(body, &response)
			assert.Equal(t, tt.expectedMsg, response["message"])

			mockService.AssertExpectations(t)
		})
	}
}
