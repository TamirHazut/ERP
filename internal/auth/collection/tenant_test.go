package collection

import (
	"errors"
	"testing"
	"time"

	mock_collection "erp.localhost/internal/infra/db/mongo/collection/mock"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	"erp.localhost/internal/infra/model/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	baseTenantLogger = logger.NewBaseLogger(shared.ModuleCore)
)

// tenantMatcher is a custom gomock matcher for Tenant objects
// It skips the CreatedAt and UpdatedAt fields which are set dynamically
type tenantMatcher struct {
	expected *authv1.Tenant
}

func (m tenantMatcher) Matches(x interface{}) bool {
	tenant, ok := x.(*authv1.Tenant)
	if !ok {
		return false
	}
	// Match all fields except CreatedAt and UpdatedAt which are set by the function
	return tenant.Name == m.expected.Name &&
		tenant.Status == m.expected.Status &&
		tenant.CreatedBy == m.expected.CreatedBy
}

func (m tenantMatcher) String() string {
	return "matches tenant fields except CreatedAt and UpdatedAt"
}

func TestNewTenantCollection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mock_collection.NewMockCollectionHandler[authv1.Tenant](ctrl)
	collection := NewTenantCollection(mockHandler, baseTenantLogger)

	require.NotNil(t, collection)
	assert.NotNil(t, collection.collection)
	assert.NotNil(t, collection.logger)
}

func TestTenantCollection_CreateTenant(t *testing.T) {
	testCases := []struct {
		name              string
		tenant            *authv1.Tenant
		returnID          string
		returnError       error
		wantErr           bool
		expectedCallTimes int
	}{
		{
			name: "successful create",
			tenant: &authv1.Tenant{
				Name:      "Test Company",
				Status:    authv1.TenantStatus_TENANT_STATUS_ACTIVE,
				CreatedBy: "admin",
			},
			returnID:          "tenant-id-123",
			returnError:       nil,
			wantErr:           false,
			expectedCallTimes: 1,
		},
		{
			name: "create with validation error - missing name",
			tenant: &authv1.Tenant{
				Status:    authv1.TenantStatus_TENANT_STATUS_ACTIVE,
				CreatedBy: "admin",
			},
			returnID:          "",
			returnError:       nil,
			wantErr:           true,
			expectedCallTimes: 0,
		},
		{
			name: "create with database error",
			tenant: &authv1.Tenant{
				Name:      "Test Company",
				Status:    authv1.TenantStatus_TENANT_STATUS_ACTIVE,
				CreatedBy: "admin",
			},
			returnID:          "",
			returnError:       errors.New("database connection failed"),
			wantErr:           true,
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[authv1.Tenant](ctrl)
			if tc.expectedCallTimes > 0 {
				mockHandler.EXPECT().
					Create(tenantMatcher{expected: tc.tenant}).
					Return(tc.returnID, tc.returnError).
					Times(tc.expectedCallTimes)
			}

			collection := NewTenantCollection(mockHandler, baseTenantLogger)
			id, err := collection.CreateTenant(tc.tenant)
			if tc.wantErr {
				require.Error(t, err)
				assert.Empty(t, id)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.returnID, id)
			}
		})
	}
}

func TestTenantCollection_GetTenantByID(t *testing.T) {
	tenantID := "507f1f77bcf86cd799439011"

	testCases := []struct {
		name              string
		tenantID          string
		expectedFilter    map[string]any
		returnTenant      *authv1.Tenant
		returnError       error
		wantErr           bool
		expectedCallTimes int
	}{
		{
			name:     "successful get by id",
			tenantID: tenantID,
			expectedFilter: map[string]any{
				"_id": tenantID,
			},
			returnTenant: &authv1.Tenant{
				Id:     tenantID,
				Name:   "Test Company",
				Status: authv1.TenantStatus_TENANT_STATUS_ACTIVE,
			},
			returnError:       nil,
			wantErr:           false,
			expectedCallTimes: 1,
		},
		{
			name:     "tenant not found",
			tenantID: tenantID,
			expectedFilter: map[string]any{
				"_id": tenantID,
			},
			returnTenant:      nil,
			returnError:       errors.New("tenant not found"),
			wantErr:           true,
			expectedCallTimes: 1,
		},
		{
			name:              "get with empty tenant ID",
			tenantID:          "",
			expectedFilter:    map[string]any{},
			returnTenant:      nil,
			returnError:       nil,
			wantErr:           true,
			expectedCallTimes: 0,
		},
		{
			name:     "database error",
			tenantID: tenantID,
			expectedFilter: map[string]any{
				"_id": tenantID,
			},
			returnTenant:      nil,
			returnError:       errors.New("database query failed"),
			wantErr:           true,
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[authv1.Tenant](ctrl)
			if tc.expectedCallTimes > 0 {
				mockHandler.EXPECT().
					FindOne(tc.expectedFilter).
					Return(tc.returnTenant, tc.returnError).
					Times(tc.expectedCallTimes)
			}

			collection := NewTenantCollection(mockHandler, baseTenantLogger)
			tenant, err := collection.GetTenantByID(tc.tenantID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.tenantID, tenant.Id)
			}
		})
	}
}

func TestTenantCollection_UpdateTenant(t *testing.T) {
	tenantID := "507f1f77bcf86cd799439011"
	createdAt := timestamppb.New(time.Now().Add(-24 * time.Hour))

	testCases := []struct {
		name                 string
		tenant               *authv1.Tenant
		expectedFindFilter   map[string]any
		expectedUpdateFilter map[string]any
		returnFindTenant     *authv1.Tenant
		returnFindError      error
		returnUpdateError    error
		wantErr              bool
		expectedFindCalls    int
		expectedUpdateCalls  int
	}{
		{
			name: "successful update",
			tenant: &authv1.Tenant{
				Id:        tenantID,
				Name:      "Test Company",
				Status:    authv1.TenantStatus_TENANT_STATUS_ACTIVE,
				CreatedBy: "admin",
				CreatedAt: createdAt,
				Domain:    "domain",
			},
			expectedFindFilter: map[string]any{
				"_id": tenantID,
			},
			expectedUpdateFilter: map[string]any{
				"_id": tenantID,
			},
			returnFindTenant: &authv1.Tenant{
				Id:        tenantID,
				Name:      "Test Company",
				Status:    authv1.TenantStatus_TENANT_STATUS_ACTIVE,
				CreatedBy: "admin",
				CreatedAt: createdAt,
			},
			returnFindError:     nil,
			returnUpdateError:   nil,
			wantErr:             false,
			expectedFindCalls:   1,
			expectedUpdateCalls: 1,
		},
		{
			name: "update with validation error",
			tenant: &authv1.Tenant{
				Id: tenantID,
			},
			expectedFindFilter:   map[string]any{},
			expectedUpdateFilter: map[string]any{},
			returnFindTenant:     nil,
			returnFindError:      errors.New("validation error"),
			returnUpdateError:    nil,
			wantErr:              true,
			expectedFindCalls:    0,
			expectedUpdateCalls:  0,
		},
		{
			name: "update with tenant not found",
			tenant: &authv1.Tenant{
				Id:        tenantID,
				Name:      "Test Company",
				Status:    authv1.TenantStatus_TENANT_STATUS_ACTIVE,
				CreatedBy: "admin",
				CreatedAt: createdAt,
			},
			expectedFindFilter: map[string]any{
				"_id": tenantID,
			},
			expectedUpdateFilter: map[string]any{},
			returnFindTenant:     nil,
			returnFindError:      errors.New("tenant not found"),
			returnUpdateError:    errors.New("validation error"),
			wantErr:              true,
			expectedFindCalls:    1,
			expectedUpdateCalls:  0,
		},
		{
			name: "update with restricted field change - CreatedAt",
			tenant: &authv1.Tenant{
				Id:        tenantID,
				Name:      "Test Company",
				Status:    authv1.TenantStatus_TENANT_STATUS_ACTIVE,
				CreatedBy: "admin",
				CreatedAt: timestamppb.Now(),
			},
			expectedFindFilter: map[string]any{
				"_id": tenantID,
			},
			expectedUpdateFilter: map[string]any{},
			returnFindTenant: &authv1.Tenant{
				Id:        tenantID,
				CreatedAt: createdAt,
			},
			returnFindError:     nil,
			returnUpdateError:   errors.New("unauthorized to change created at"),
			wantErr:             true,
			expectedFindCalls:   1,
			expectedUpdateCalls: 0,
		},
		{
			name: "update with database error",
			tenant: &authv1.Tenant{
				Id:        tenantID,
				Name:      "Test Company",
				Status:    authv1.TenantStatus_TENANT_STATUS_ACTIVE,
				CreatedBy: "admin",
				CreatedAt: createdAt,
			},
			expectedFindFilter: map[string]any{
				"_id": tenantID,
			},
			expectedUpdateFilter: map[string]any{
				"_id": tenantID,
			},
			returnFindTenant: &authv1.Tenant{
				Id:        tenantID,
				CreatedAt: createdAt,
			},
			returnFindError:     nil,
			returnUpdateError:   errors.New("update failed"),
			wantErr:             true,
			expectedFindCalls:   1,
			expectedUpdateCalls: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[authv1.Tenant](ctrl)
			if tc.expectedFindCalls > 0 {
				mockHandler.EXPECT().
					FindOne(tc.expectedFindFilter).
					Return(tc.returnFindTenant, tc.returnFindError).
					Times(tc.expectedFindCalls)
			}
			if tc.expectedUpdateCalls > 0 {
				mockHandler.EXPECT().
					Update(tc.expectedUpdateFilter, tenantMatcher{expected: tc.tenant}).
					Return(tc.returnUpdateError).
					Times(tc.expectedUpdateCalls)
			}

			collection := NewTenantCollection(mockHandler, baseTenantLogger)
			err := collection.UpdateTenant(tc.tenant)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTenantCollection_DeleteTenant(t *testing.T) {
	testCases := []struct {
		name              string
		tenantID          string
		expectedFilter    map[string]any
		returnError       error
		wantErr           bool
		expectedCallTimes int
	}{
		{
			name:     "successful delete",
			tenantID: "507f1f77bcf86cd799439011",
			expectedFilter: map[string]any{
				"_id": "507f1f77bcf86cd799439011",
			},
			returnError:       nil,
			wantErr:           false,
			expectedCallTimes: 1,
		},
		{
			name:              "delete with empty tenant ID",
			tenantID:          "",
			expectedFilter:    map[string]any{},
			returnError:       nil,
			wantErr:           true,
			expectedCallTimes: 0,
		},
		{
			name:     "delete with database error",
			tenantID: "507f1f77bcf86cd799439011",
			expectedFilter: map[string]any{
				"_id": "507f1f77bcf86cd799439011",
			},
			returnError:       errors.New("delete failed"),
			wantErr:           true,
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[authv1.Tenant](ctrl)
			if tc.expectedCallTimes > 0 {
				mockHandler.EXPECT().
					Delete(tc.expectedFilter).
					Return(tc.returnError).
					Times(tc.expectedCallTimes)
			}

			collection := NewTenantCollection(mockHandler, baseTenantLogger)
			err := collection.DeleteTenant(tc.tenantID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
