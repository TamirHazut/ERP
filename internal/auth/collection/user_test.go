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
	baseUserLogger = logger.NewBaseLogger(shared.ModuleCore)
)

// userMatcher is a custom gomock matcher for User objects
// It skips the CreatedAt and UpdatedAt fields which are set dynamically
type userMatcher struct {
	expected *authv1.User
}

func (m userMatcher) Matches(x interface{}) bool {
	user, ok := x.(*authv1.User)
	if !ok {
		return false
	}
	// Match all fields except CreatedAt and UpdatedAt which are set by the function
	return user.TenantId == m.expected.TenantId &&
		user.Email == m.expected.Email &&
		user.Username == m.expected.Username &&
		user.PasswordHash == m.expected.PasswordHash &&
		user.Status == m.expected.Status &&
		user.CreatedBy == m.expected.CreatedBy &&
		len(user.Roles) == len(m.expected.Roles)
}

func (m userMatcher) String() string {
	return "matches user fields except CreatedAt and UpdatedAt"
}

func TestNewUserCollection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mock_collection.NewMockCollectionHandler[authv1.User](ctrl)
	collection := NewUserCollection(mockHandler, baseUserLogger)

	require.NotNil(t, collection)
	require.NotNil(t, collection.collection)
	require.NotNil(t, collection.logger)
}

func TestUserCollection_CreateUser(t *testing.T) {
	testCases := []struct {
		name              string
		user              *authv1.User
		returnID          string
		returnError       error
		wantErr           bool
		expectedCallTimes int
	}{
		{
			name: "successful create",
			user: &authv1.User{
				TenantId:     "tenant1",
				Email:        "test@example.com",
				Username:     "testuser",
				PasswordHash: "hashed_password",
				Status:       authv1.UserStatus_USER_STATUS_ACTIVE,
				CreatedBy:    "admin",
				Roles:        []*authv1.UserRole{},
			},
			returnID:          "user-id-123",
			returnError:       nil,
			wantErr:           false,
			expectedCallTimes: 1,
		},
		{
			name: "create with validation error - missing tenant ID",
			user: &authv1.User{
				Email:        "test@example.com",
				Username:     "testuser",
				PasswordHash: "hashed_password",
				Status:       authv1.UserStatus_USER_STATUS_ACTIVE,
				CreatedBy:    "admin",
				Roles:        []*authv1.UserRole{},
			},
			returnID:          "",
			returnError:       nil,
			wantErr:           true,
			expectedCallTimes: 0,
		},
		{
			name: "create with database error",
			user: &authv1.User{
				TenantId:     "tenant1",
				Email:        "test@example.com",
				Username:     "testuser",
				PasswordHash: "hashed_password",
				Status:       authv1.UserStatus_USER_STATUS_ACTIVE,
				CreatedBy:    "admin",
				Roles:        []*authv1.UserRole{},
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

			mockHandler := mock_collection.NewMockCollectionHandler[authv1.User](ctrl)
			if tc.expectedCallTimes > 0 {
				mockHandler.EXPECT().
					Create(userMatcher{expected: tc.user}).
					Return(tc.returnID, tc.returnError).
					Times(tc.expectedCallTimes)
			}

			collection := NewUserCollection(mockHandler, baseUserLogger)
			id, err := collection.CreateUser(tc.user)
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

func TestUserCollection_GetUserByID(t *testing.T) {
	userID := "507f1f77bcf86cd799439011"

	testCases := []struct {
		name           string
		tenantID       string
		userID         string
		expectedFilter map[string]any
		returnUser     *authv1.User
		returnError    error
		wantErr        bool
	}{
		{
			name:     "successful get by id",
			tenantID: "tenant1",
			userID:   "507f1f77bcf86cd799439011",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "507f1f77bcf86cd799439011",
			},
			returnUser: &authv1.User{
				Id:       userID,
				TenantId: "tenant1",
				Email:    "test@example.com",
				Username: "testuser",
			},
			returnError: nil,
			wantErr:     false,
		},
		{
			name:     "user not found",
			tenantID: "tenant1",
			userID:   "507f1f77bcf86cd799439011",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "507f1f77bcf86cd799439011",
			},
			returnUser:  nil,
			returnError: errors.New("user not found"),
			wantErr:     true,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			userID:   "507f1f77bcf86cd799439011",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "507f1f77bcf86cd799439011",
			},
			returnUser:  nil,
			returnError: errors.New("database query failed"),
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[authv1.User](ctrl)
			mockHandler.EXPECT().
				FindOne(tc.expectedFilter).
				Return(tc.returnUser, tc.returnError)

			collection := NewUserCollection(mockHandler, baseUserLogger)
			user, err := collection.GetUserByID(tc.tenantID, tc.userID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.tenantID, user.TenantId)
			}
		})
	}
}

func TestUserCollection_GetUserByUsername(t *testing.T) {
	testCases := []struct {
		name           string
		tenantID       string
		username       string
		returnUser     *authv1.User
		returnError    error
		expectedFilter map[string]any
		wantErr        bool
	}{
		{
			name:     "successful get by username",
			tenantID: "tenant1",
			username: "testuser",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"username":  "testuser",
			},
			returnUser: &authv1.User{
				TenantId: "tenant1",
				Username: "testuser",
				Email:    "test@example.com",
			},
			returnError: nil,
			wantErr:     false,
		},
		{
			name:     "user not found",
			tenantID: "tenant1",
			username: "nonexistent",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"username":  "nonexistent",
			},
			returnUser:  nil,
			returnError: errors.New("user not found"),
			wantErr:     true,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			username: "testuser",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"username":  "testuser",
			},
			returnUser:  nil,
			returnError: errors.New("database query failed"),
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[authv1.User](ctrl)
			mockHandler.EXPECT().
				FindOne(tc.expectedFilter).
				Return(tc.returnUser, tc.returnError)

			collection := NewUserCollection(mockHandler, baseUserLogger)
			user, err := collection.GetUserByUsername(tc.tenantID, tc.username)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.username, user.Username)
			}
		})
	}
}

func TestUserCollection_GetUsersByTenantID(t *testing.T) {
	testCases := []struct {
		name           string
		tenantID       string
		returnUsers    []*authv1.User
		returnError    error
		expectedFilter map[string]any
		wantCount      int
		wantErr        bool
	}{
		{
			name:     "successful get users by tenant",
			tenantID: "tenant1",
			returnUsers: []*authv1.User{
				&authv1.User{TenantId: "tenant1", Username: "user1"},
				&authv1.User{TenantId: "tenant1", Username: "user2"},
			},
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
			},
			returnError: nil,
			wantCount:   2,
			wantErr:     false,
		},
		{
			name:        "no users found",
			tenantID:    "tenant1",
			returnUsers: []*authv1.User{},
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
			},
			returnError: nil,
			wantCount:   0,
			wantErr:     false,
		},
		{
			name:        "database error",
			tenantID:    "tenant1",
			returnUsers: []*authv1.User{},
			returnError: errors.New("database query failed"),
			wantCount:   0,
			wantErr:     true,
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[authv1.User](ctrl)
			mockHandler.EXPECT().
				FindAll(tc.expectedFilter).
				Return(tc.returnUsers, tc.returnError)

			collection := NewUserCollection(mockHandler, baseUserLogger)
			users, err := collection.GetUsersByTenantID(tc.tenantID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				if tc.wantCount == 0 {
					assert.Empty(t, users)
				} else {
					require.NoError(t, err)
					assert.Len(t, users, tc.wantCount)
				}
			}
		})
	}
}

func TestUserCollection_GetUsersByRoleID(t *testing.T) {
	testCases := []struct {
		name           string
		tenantID       string
		roleID         string
		returnUsers    []*authv1.User
		returnError    error
		expectedFilter map[string]any
		wantCount      int
		wantErr        bool
	}{
		{
			name:     "successful get users by role",
			tenantID: "tenant1",
			roleID:   "role1",
			returnUsers: []*authv1.User{
				&authv1.User{TenantId: "tenant1", Username: "user1"},
			},
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"role_id":   "role1",
			},
			returnError: nil,
			wantCount:   1,
			wantErr:     false,
		},
		{
			name:        "database error",
			tenantID:    "tenant1",
			roleID:      "role1",
			returnUsers: []*authv1.User{},
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"role_id":   "role1",
			},
			returnError: errors.New("database query failed"),
			wantCount:   0,
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[authv1.User](ctrl)
			mockHandler.EXPECT().
				FindAll(tc.expectedFilter).
				Return(tc.returnUsers, tc.returnError)

			collection := NewUserCollection(mockHandler, baseUserLogger)
			users, err := collection.GetUsersByRoleID(tc.tenantID, tc.roleID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, users, tc.wantCount)
			}
		})
	}
}

func TestUserCollection_UpdateUser(t *testing.T) {
	userID := "507f1f77bcf86cd799439011"
	createdAt := timestamppb.New(time.Now().Add(-24 * time.Hour))

	testCases := []struct {
		name                 string
		user                 *authv1.User
		returnFindUser       *authv1.User
		returnUpdateError    error
		wantErr              bool
		expectedUpdateCalls  int
		expectedUpdateFilter map[string]any
	}{
		{
			name: "successful update",
			user: &authv1.User{
				Id:           userID,
				TenantId:     "tenant1",
				Email:        "updated@example.com",
				Username:     "testuser",
				PasswordHash: "hashed_password",
				Status:       authv1.UserStatus_USER_STATUS_ACTIVE,
				CreatedBy:    "admin",
				CreatedAt:    createdAt,
			},
			expectedUpdateFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       userID,
			},
			returnUpdateError:   nil,
			wantErr:             false,
			expectedUpdateCalls: 1,
		},
		{
			name: "update with validation error",
			user: &authv1.User{
				TenantId: "tenant1",
			},
			expectedUpdateFilter: map[string]any{},
			returnFindUser:       nil,
			returnUpdateError:    errors.New("validation error"),
			wantErr:              true,
			expectedUpdateCalls:  0,
		},
		{
			name: "update with database error",
			user: &authv1.User{
				Id:           userID,
				TenantId:     "tenant1",
				Email:        "test@example.com",
				Username:     "testuser",
				PasswordHash: "hashed_password",
				Status:       authv1.UserStatus_USER_STATUS_ACTIVE,
				CreatedBy:    "admin",
				CreatedAt:    createdAt,
				Roles:        []*authv1.UserRole{},
			},
			expectedUpdateFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       userID,
			},
			returnFindUser: &authv1.User{
				Id:        userID,
				TenantId:  "tenant1",
				Username:  "testuser",
				CreatedAt: createdAt,
			},
			returnUpdateError:   errors.New("update failed"),
			wantErr:             true,
			expectedUpdateCalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[authv1.User](ctrl)
			if tc.expectedUpdateCalls > 0 {
				mockHandler.EXPECT().
					Update(tc.expectedUpdateFilter, userMatcher{expected: tc.user}).
					Return(tc.returnUpdateError).
					Times(tc.expectedUpdateCalls)
			}

			collection := NewUserCollection(mockHandler, baseUserLogger)
			err := collection.UpdateUser(tc.user)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUserCollection_DeleteUser(t *testing.T) {
	testCases := []struct {
		name              string
		tenantID          string
		userID            string
		expectedFilter    map[string]any
		returnError       error
		wantErr           bool
		expectedCallTimes int
	}{
		{
			name:     "successful delete",
			tenantID: "tenant1",
			userID:   "user-id-123",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "user-id-123",
			},
			returnError:       nil,
			wantErr:           false,
			expectedCallTimes: 1,
		},
		{
			name:              "delete with empty tenant ID",
			tenantID:          "",
			userID:            "user-id-123",
			expectedFilter:    map[string]any{},
			returnError:       nil,
			wantErr:           true,
			expectedCallTimes: 0,
		},
		{
			name:              "delete with empty user ID",
			tenantID:          "tenant1",
			userID:            "",
			expectedFilter:    nil,
			returnError:       nil,
			wantErr:           true,
			expectedCallTimes: 0,
		},
		{
			name:     "delete with database error",
			tenantID: "tenant1",
			userID:   "user-id-123",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "user-id-123",
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

			mockHandler := mock_collection.NewMockCollectionHandler[authv1.User](ctrl)
			if tc.expectedCallTimes > 0 {
				mockHandler.EXPECT().
					Delete(tc.expectedFilter).
					Return(tc.returnError).
					Times(tc.expectedCallTimes)
			}

			collection := NewUserCollection(mockHandler, baseUserLogger)
			err := collection.DeleteUser(tc.tenantID, tc.userID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
