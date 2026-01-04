package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPermission_Validate(t *testing.T) {
	tests := []struct {
		name            string
		permission      *Permission
		createOperation bool
		wantErr         bool
		expectedErrMsg  string
	}{
		{
			name: "valid permission on create",
			permission: &Permission{
				TenantID:         "tenant-123",
				Resource:         "users",
				Action:           "read",
				PermissionString: "users:read",
				DisplayName:      "Read Users",
				CreatedBy:        "admin",
			},
			createOperation: true,
			wantErr:         false,
		},
		{
			name: "valid permission on update",
			permission: &Permission{
				ID:               primitive.NewObjectID(),
				TenantID:         "tenant-123",
				Resource:         "users",
				Action:           "read",
				PermissionString: "users:read",
				DisplayName:      "Read Users",
				CreatedBy:        "admin",
			},
			createOperation: false,
			wantErr:         false,
		},
		{
			name: "missing ID on update",
			permission: &Permission{
				TenantID:         "tenant-123",
				Resource:         "users",
				Action:           "read",
				PermissionString: "users:read",
				DisplayName:      "Read Users",
				CreatedBy:        "admin",
			},
			createOperation: false,
			wantErr:         true,
			expectedErrMsg:  "ID",
		},
		{
			name: "missing tenantID",
			permission: &Permission{
				Resource:         "users",
				Action:           "read",
				PermissionString: "users:read",
				DisplayName:      "Read Users",
				CreatedBy:        "admin",
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "TenantID",
		},
		{
			name: "missing resource",
			permission: &Permission{
				TenantID:         "tenant-123",
				Action:           "read",
				PermissionString: "users:read",
				DisplayName:      "Read Users",
				CreatedBy:        "admin",
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "Resource",
		},
		{
			name: "missing action",
			permission: &Permission{
				TenantID:         "tenant-123",
				Resource:         "users",
				PermissionString: "users:read",
				DisplayName:      "Read Users",
				CreatedBy:        "admin",
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "Action",
		},
		{
			name: "missing permission string",
			permission: &Permission{
				TenantID:    "tenant-123",
				Resource:    "users",
				Action:      "read",
				DisplayName: "Read Users",
				CreatedBy:   "admin",
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "PermissionString",
		},
		{
			name: "missing display name",
			permission: &Permission{
				TenantID:         "tenant-123",
				Resource:         "users",
				Action:           "read",
				PermissionString: "users:read",
				CreatedBy:        "admin",
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "DisplayName",
		},
		{
			name: "missing createdBy",
			permission: &Permission{
				TenantID:         "tenant-123",
				Resource:         "users",
				Action:           "read",
				PermissionString: "users:read",
				DisplayName:      "Read Users",
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "CreatedBy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.permission.Validate(tt.createOperation)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
