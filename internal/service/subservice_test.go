package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/iokiris/efm-subscription-api/internal/logger"

	"github.com/iokiris/efm-subscription-api/internal/model"
	"github.com/iokiris/efm-subscription-api/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// -------------------- Mocks --------------------

func init() {
	logger.L = zap.NewNop()
}

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Create(ctx context.Context, sub *model.Subscription) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *MockRepo) Update(ctx context.Context, sub *model.Subscription) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *MockRepo) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockRepo) GetByID(ctx context.Context, id int64) (*model.Subscription, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockRepo) List(ctx context.Context, userID string) ([]model.Subscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.Subscription), args.Error(1)
}

func (m *MockRepo) GetSummary(ctx context.Context, userID, serviceName string, from, to time.Time) (int, error) {
	args := m.Called(ctx, userID, serviceName, from, to)
	return args.Int(0), args.Error(1)
}

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(exchange, routingKey string, body []byte) error {
	return m.Called(exchange, routingKey, body).Error(0)
}

func (m *MockPublisher) Close() {
	m.Called()
}

// -------------------- Tests --------------------

func TestSubscriptionService_Create(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockPub := new(MockPublisher)

	svc := service.NewSubscriptionService(mockRepo, nil, mockPub, time.Minute)
	sub := &model.Subscription{ID: 1, UserID: "user1", Service: "test_service"}

	mockRepo.On("Create", ctx, sub).Return(nil)
	mockPub.On("Publish", "subscriptions", "created", mock.Anything).Return(nil)

	err := svc.Create(ctx, sub)
	assert.NoError(t, err)

	mockRepo.AssertCalled(t, "Create", ctx, sub)
	mockPub.AssertCalled(t, "Publish", "subscriptions", "created", mock.Anything)
}

func TestSubscriptionService_Update(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockPub := new(MockPublisher)

	svc := service.NewSubscriptionService(mockRepo, nil, mockPub, time.Minute)
	sub := &model.Subscription{ID: 1, UserID: "user1", Service: "service"}

	mockRepo.On("Update", ctx, sub).Return(nil)
	mockPub.On("Publish", "subscriptions", "updated", mock.Anything).Return(nil)

	err := svc.Update(ctx, sub)
	assert.NoError(t, err)

	mockRepo.AssertCalled(t, "Update", ctx, sub)
	mockPub.AssertCalled(t, "Publish", "subscriptions", "updated", mock.Anything)
}

func TestSubscriptionService_Delete(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockPub := new(MockPublisher)

	svc := service.NewSubscriptionService(mockRepo, nil, mockPub, time.Minute)

	sub := &model.Subscription{ID: 1, UserID: "user1"}
	mockRepo.On("GetByID", ctx, int64(1)).Return(sub, nil)
	mockRepo.On("Delete", ctx, int64(1)).Return(nil)
	mockPub.On("Publish", "subscriptions", "deleted", mock.Anything).Return(nil)

	err := svc.Delete(ctx, 1, "")
	assert.NoError(t, err)

	mockRepo.AssertCalled(t, "Delete", ctx, int64(1))
	mockPub.AssertCalled(t, "Publish", "subscriptions", "deleted", mock.Anything)
}

func TestSubscriptionService_Get(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	svc := service.NewSubscriptionService(mockRepo, nil, nil, time.Minute)

	sub := &model.Subscription{ID: 1, UserID: "user1", Service: "s"}
	mockRepo.On("GetByID", ctx, int64(1)).Return(sub, nil)

	result, err := svc.Get(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, sub, result)
}

func TestSubscriptionService_List(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	svc := service.NewSubscriptionService(mockRepo, nil, nil, time.Minute)

	list := []model.Subscription{{ID: 1, UserID: "user1", Service: "s"}}
	mockRepo.On("List", ctx, "user1").Return(list, nil)

	result, err := svc.List(ctx, "user1")
	assert.NoError(t, err)
	assert.Equal(t, list, result)
}

func TestSubscriptionService_GetSummary_DBCall(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	svc := service.NewSubscriptionService(mockRepo, nil, nil, time.Minute)

	from := "01-2025"
	to := "12-2025"

	fromTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	toTime := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)

	mockRepo.On("GetSummary", ctx, "user1", "test", fromTime, toTime).Return(100, nil)

	total, err := svc.GetSummary(ctx, "user1", "test", from, to)
	assert.NoError(t, err)
	assert.Equal(t, 100, total)

	mockRepo.AssertCalled(t, "GetSummary", ctx, "user1", "test", fromTime, toTime)
}
