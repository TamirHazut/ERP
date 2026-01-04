package collection

import (
	"errors"
	"testing"
	"time"

	mongo_mocks "erp.localhost/internal/infra/db/mongo/mocks"
	erp_errors "erp.localhost/internal/infra/errors"
	auth_models "erp.localhost/internal/infra/models/auth"
	events_models "erp.localhost/internal/infra/models/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// auditLogMatcher is a custom gomock matcher for AuditLog objects
// It skips the Timestamp field which is set dynamically
type auditLogMatcher struct {
	expected events_models.AuditLog
}

func (m auditLogMatcher) Matches(x interface{}) bool {
	log, ok := x.(events_models.AuditLog)
	if !ok {
		return false
	}
	// Match all fields except Timestamp which is set by the function
	return log.TenantID == m.expected.TenantID &&
		log.Action == m.expected.Action &&
		log.Category == m.expected.Category &&
		log.Severity == m.expected.Severity &&
		log.Result == m.expected.Result &&
		log.ActorType == m.expected.ActorType &&
		log.ActorID == m.expected.ActorID &&
		log.ActorName == m.expected.ActorName &&
		log.TargetType == m.expected.TargetType &&
		log.TargetID == m.expected.TargetID &&
		log.TargetName == m.expected.TargetName
}

func (m auditLogMatcher) String() string {
	return "matches audit log fields except Timestamp"
}

func TestNewAuditLogsCollection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mongo_mocks.NewMockCollectionHandler[events_models.AuditLog](ctrl)
	collection := NewAuditLogsCollection(mockHandler)

	require.NotNil(t, collection)
	require.NotNil(t, collection.collection)
	require.NotNil(t, collection.logger)
}

func TestAuditLogsCollection_CreateAuditLog(t *testing.T) {
	testCases := []struct {
		name              string
		tenantID          string
		auditLog          events_models.AuditLog
		returnID          string
		returnError       error
		expectedError     error
		expectedCallTimes int
	}{
		{
			name:     "successful create",
			tenantID: "tenant-1",
			auditLog: events_models.AuditLog{
				TenantID:   "tenant-1",
				Action:     auth_models.ActionLogin,
				Category:   auth_models.CategoryAuth,
				Severity:   auth_models.SeverityInfo,
				Result:     auth_models.ResultSuccess,
				ActorType:  auth_models.ActorTypeUser,
				ActorID:    "user-1",
				ActorName:  "John Doe",
				TargetType: auth_models.TargetTypeUser,
				TargetID:   "user-1",
				TargetName: "John Doe",
			},
			returnID:          "audit-log-1",
			returnError:       nil,
			expectedError:     nil,
			expectedCallTimes: 1,
		},
		{
			name:              "missing tenantID",
			tenantID:          "",
			auditLog:          events_models.AuditLog{},
			returnID:          "",
			returnError:       nil,
			expectedError:     erp_errors.Validation(erp_errors.ValidationRequiredFields, "tenantID"),
			expectedCallTimes: 0,
		},
		{
			name:     "invalid audit log - missing action",
			tenantID: "tenant-1",
			auditLog: events_models.AuditLog{
				Category: "authentication",
				Severity: "info",
			},
			returnID:          "",
			returnError:       nil,
			expectedError:     erp_errors.Validation(erp_errors.ValidationRequiredFields, "action"),
			expectedCallTimes: 0,
		},
		{
			name:     "database error during create",
			tenantID: "tenant-1",
			auditLog: events_models.AuditLog{
				TenantID:   "tenant-1",
				Action:     auth_models.ActionLogin,
				Category:   auth_models.CategoryAuth,
				Severity:   auth_models.SeverityInfo,
				Result:     auth_models.ResultSuccess,
				ActorType:  auth_models.ActorTypeUser,
				ActorID:    "user-1",
				ActorName:  "John Doe",
				TargetType: auth_models.TargetTypeUser,
				TargetID:   "user-1",
				TargetName: "John Doe",
			},
			returnID:          "",
			returnError:       errors.New("database connection failed"),
			expectedError:     errors.New("database connection failed"),
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mongo_mocks.NewMockCollectionHandler[events_models.AuditLog](ctrl)

			// Only expect Create call if we expect it to be called
			if tc.expectedCallTimes > 0 {
				mockHandler.EXPECT().
					Create(auditLogMatcher{expected: tc.auditLog}).
					Return(tc.returnID, tc.returnError).
					Times(tc.expectedCallTimes)
			}

			collection := NewAuditLogsCollection(mockHandler)
			err := collection.CreateAuditLog(tc.tenantID, tc.auditLog)

			if tc.expectedError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuditLogsCollection_GetAuditLogsByFilter(t *testing.T) {
	testAuditLog1 := events_models.AuditLog{
		TenantID:  "tenant-1",
		Action:    "user.login",
		Category:  "authentication",
		Severity:  "info",
		Result:    "success",
		Timestamp: time.Now(),
		ActorType: auth_models.ActorTypeUser,
		ActorID:   "user-1",
		ActorName: "John Doe",
	}

	testAuditLog2 := events_models.AuditLog{
		TenantID:  "tenant-1",
		Action:    "user.logout",
		Category:  "authentication",
		Severity:  "info",
		Result:    "success",
		Timestamp: time.Now(),
		ActorType: auth_models.ActorTypeUser,
		ActorID:   "user-1",
		ActorName: "John Doe",
	}

	testCases := []struct {
		name              string
		tenantID          string
		filter            map[string]any
		expectedFilter    map[string]any
		returnLogs        []events_models.AuditLog
		returnError       error
		expectedLogs      []events_models.AuditLog
		expectedError     error
		expectedCallTimes int
	}{
		{
			name:     "successful get with action filter",
			tenantID: "tenant-1",
			filter:   map[string]any{"action": "user.login"},
			expectedFilter: map[string]any{
				"action":    "user.login",
				"tenant_id": "tenant-1",
			},
			returnLogs: []events_models.AuditLog{
				testAuditLog1,
			},
			returnError:       nil,
			expectedLogs:      []events_models.AuditLog{testAuditLog1},
			expectedError:     nil,
			expectedCallTimes: 1,
		},
		{
			name:     "successful get with category filter",
			tenantID: "tenant-1",
			filter:   map[string]any{"category": "authentication"},
			expectedFilter: map[string]any{
				"category":  "authentication",
				"tenant_id": "tenant-1",
			},
			returnLogs: []events_models.AuditLog{
				testAuditLog1,
				testAuditLog2,
			},
			returnError: nil,
			expectedLogs: []events_models.AuditLog{
				testAuditLog1,
				testAuditLog2,
			},
			expectedError:     nil,
			expectedCallTimes: 1,
		},
		{
			name:     "successful get with no filter",
			tenantID: "tenant-1",
			filter:   nil,
			expectedFilter: map[string]any{
				"tenant_id": "tenant-1",
			},
			returnLogs: []events_models.AuditLog{
				testAuditLog1,
				testAuditLog2,
			},
			returnError: nil,
			expectedLogs: []events_models.AuditLog{
				testAuditLog1,
				testAuditLog2,
			},
			expectedError:     nil,
			expectedCallTimes: 1,
		},
		{
			name:     "successful get with empty results",
			tenantID: "tenant-1",
			filter:   map[string]any{"action": "nonexistent.action"},
			expectedFilter: map[string]any{
				"action":    "nonexistent.action",
				"tenant_id": "tenant-1",
			},
			returnLogs:        []events_models.AuditLog{},
			returnError:       nil,
			expectedLogs:      []events_models.AuditLog{},
			expectedError:     nil,
			expectedCallTimes: 1,
		},
		{
			name:              "missing tenantID",
			tenantID:          "",
			filter:            map[string]any{"action": "user.login"},
			expectedFilter:    nil,
			returnLogs:        nil,
			returnError:       nil,
			expectedLogs:      nil,
			expectedError:     erp_errors.Validation(erp_errors.ValidationRequiredFields, "tenantID"),
			expectedCallTimes: 0,
		},
		{
			name:     "database error during find",
			tenantID: "tenant-1",
			filter:   map[string]any{"action": "user.login"},
			expectedFilter: map[string]any{
				"action":    "user.login",
				"tenant_id": "tenant-1",
			},
			returnLogs:        nil,
			returnError:       errors.New("database query failed"),
			expectedLogs:      nil,
			expectedError:     errors.New("database query failed"),
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mongo_mocks.NewMockCollectionHandler[events_models.AuditLog](ctrl)
			// Only expect FindAll call if we expect it to be called
			if tc.expectedCallTimes > 0 {
				mockHandler.EXPECT().
					FindAll(tc.expectedFilter).
					Return(tc.returnLogs, tc.returnError).
					Times(tc.expectedCallTimes)
			}

			collection := NewAuditLogsCollection(mockHandler)
			logs, err := collection.GetAuditLogsByFilter(tc.tenantID, tc.filter)

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError.Error())
				assert.Nil(t, logs)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedLogs, logs)
			}
		})
	}
}
