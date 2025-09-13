package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/iokiris/efm-subscription-api/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSubscriptionService мок для SubscriptionService
type MockSubscriptionService struct {
	mock.Mock
}

func (m *MockSubscriptionService) Create(ctx context.Context, sub *model.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSubscriptionService) Update(ctx context.Context, sub *model.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSubscriptionService) Delete(ctx context.Context, id int64, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockSubscriptionService) Get(ctx context.Context, id int64) (*model.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) List(ctx context.Context, userID string) ([]model.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) GetSummary(ctx context.Context, userID, serviceName, from, to string) (int, error) {
	args := m.Called(ctx, userID, serviceName, from, to)
	return args.Int(0), args.Error(1)
}

func setupTestRouter(mockSvc *MockSubscriptionService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := &SubscriptionHandler{svc: mockSvc}
	h.RegisterRoutes(r, false)

	return r
}

func TestSubscriptionHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		mockSetup      func(*MockSubscriptionService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful creation",
			requestBody: map[string]interface{}{
				"service_name": "Yandex Plus",
				"price":        400,
				"user_id":      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
				"start_date":   "07-2025",
			},
			mockSetup: func(m *MockSubscriptionService) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(sub *model.Subscription) bool {
					return sub.Service == "Yandex Plus" &&
						sub.Price == 400 &&
						sub.UserID == "60601fee-2bf1-4721-ae6f-7636e79a0cba" &&
						time.Time(sub.StartDate).Equal(time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "invalid JSON",
			requestBody: map[string]interface{}{
				"service_name": 123, // некорректный тип
			},
			mockSetup: func(_ *MockSubscriptionService) {
				// Не должен вызываться
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "service error",
			requestBody: map[string]interface{}{
				"service_name": "Yandex Plus",
				"price":        400,
				"user_id":      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
				"start_date":   "07-2025",
			},
			mockSetup: func(m *MockSubscriptionService) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*model.Subscription")).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockSubscriptionService)
			tt.mockSetup(mockSvc)

			router := setupTestRouter(mockSvc)
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/subscriptions", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody["service_name"], response["service_name"])
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestSubscriptionHandler_Get(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(*MockSubscriptionService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful get",
			id:   "1",
			mockSetup: func(m *MockSubscriptionService) {
				sub := &model.Subscription{
					ID:        1,
					Service:   "Yandex Plus",
					Price:     400,
					UserID:    "60601fee-2bf1-4721-ae6f-7636e79a0cba",
					StartDate: model.MonthYear(time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)),
				}
				m.On("Get", mock.Anything, int64(1)).Return(sub, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "invalid id",
			id:   "invalid",
			mockSetup: func(_ *MockSubscriptionService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "not found",
			id:   "999",
			mockSetup: func(m *MockSubscriptionService) {
				m.On("Get", mock.Anything, int64(999)).Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockSubscriptionService)
			tt.mockSetup(mockSvc)

			router := setupTestRouter(mockSvc)

			req := httptest.NewRequest("GET", "/subscriptions/"+tt.id, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestSubscriptionHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		requestBody    map[string]interface{}
		mockSetup      func(*MockSubscriptionService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful update",
			id:   "1",
			requestBody: map[string]interface{}{
				"service_name": "Yandex Plus Updated",
				"price":        500,
				"user_id":      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
				"start_date":   "08-2025",
			},
			mockSetup: func(m *MockSubscriptionService) {
				m.On("Update", mock.Anything, mock.MatchedBy(func(sub *model.Subscription) bool {
					return sub.Service == "Yandex Plus Updated" &&
						sub.Price == 500 &&
						sub.UserID == "60601fee-2bf1-4721-ae6f-7636e79a0cba" &&
						sub.StartDate == model.MonthYear(time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC))
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "invalid id",
			id:   "invalid",
			requestBody: map[string]interface{}{
				"service_name": "Yandex Plus",
			},
			mockSetup: func(_ *MockSubscriptionService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockSubscriptionService)
			tt.mockSetup(mockSvc)

			router := setupTestRouter(mockSvc)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/subscriptions/"+tt.id, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestSubscriptionHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(*MockSubscriptionService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful delete",
			id:   "1",
			mockSetup: func(m *MockSubscriptionService) {
				m.On("Delete", mock.Anything, int64(1), "").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedError:  false,
		},
		{
			name: "invalid id",
			id:   "invalid",
			mockSetup: func(_ *MockSubscriptionService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "not found",
			id:   "999",
			mockSetup: func(m *MockSubscriptionService) {
				m.On("Delete", mock.Anything, int64(999), "").Return(errors.New("not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockSubscriptionService)
			tt.mockSetup(mockSvc)

			router := setupTestRouter(mockSvc)

			req := httptest.NewRequest("DELETE", "/subscriptions/"+tt.id, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestSubscriptionHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockSubscriptionService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "successful list",
			userID: "60601fee-2bf1-4721-ae6f-7636e79a0cba",
			mockSetup: func(m *MockSubscriptionService) {
				subs := []model.Subscription{
					{
						ID:        1,
						Service:   "Yandex Plus",
						Price:     400,
						UserID:    "60601fee-2bf1-4721-ae6f-7636e79a0cba",
						StartDate: model.MonthYear(time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)),
					},
				}
				m.On("List", mock.Anything, "60601fee-2bf1-4721-ae6f-7636e79a0cba").Return(subs, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "missing user_id",
			userID: "",
			mockSetup: func(_ *MockSubscriptionService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "service error",
			userID: "60601fee-2bf1-4721-ae6f-7636e79a0cba",
			mockSetup: func(m *MockSubscriptionService) {
				m.On("List", mock.Anything, "60601fee-2bf1-4721-ae6f-7636e79a0cba").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockSubscriptionService)
			tt.mockSetup(mockSvc)

			router := setupTestRouter(mockSvc)

			u := "/subscriptions"
			if tt.userID != "" {
				u += "?user_id=" + tt.userID
			}

			req := httptest.NewRequest("GET", u, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			} else if tt.userID != "" {
				var response []model.Subscription
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 1)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestSubscriptionHandler_Summary(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockSubscriptionService)
		expectedStatus int
		expectedError  bool
		expectedTotal  int
	}{
		{
			name:        "successful summary",
			queryParams: "user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba",
			mockSetup: func(m *MockSubscriptionService) {
				m.On("GetSummary", mock.Anything, "60601fee-2bf1-4721-ae6f-7636e79a0cba", "", "", "").Return(1200, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			expectedTotal:  1200,
		},
		{
			name:        "summary with filters",
			queryParams: "user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex Plus&from=07-2025&to=08-2025",
			mockSetup: func(m *MockSubscriptionService) {
				m.On("GetSummary", mock.Anything, "60601fee-2bf1-4721-ae6f-7636e79a0cba", "Yandex Plus", "07-2025", "08-2025").Return(400, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			expectedTotal:  400,
		},
		{
			name:        "missing user_id",
			queryParams: "",
			mockSetup: func(_ *MockSubscriptionService) {
				// Не должен вызываться
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "invalid date format",
			queryParams: "user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&from=invalid-date",
			mockSetup: func(m *MockSubscriptionService) {
				m.On("GetSummary", mock.Anything, "60601fee-2bf1-4721-ae6f-7636e79a0cba", "", "invalid-date", "").Return(0, errors.New("invalid date format"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockSubscriptionService)
			tt.mockSetup(mockSvc)

			router := setupTestRouter(mockSvc)
			baseURL := "/subscriptions/summary"
			fullURL := baseURL
			if tt.queryParams != "" {
				parsed, err := url.ParseQuery(tt.queryParams)
				assert.NoError(t, err)
				fullURL = baseURL + "?" + parsed.Encode()
			}

			req := httptest.NewRequest("GET", fullURL, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(tt.expectedTotal), response["total"])
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
