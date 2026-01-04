package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRole_Validate(t *testing.T) {
	tests := []struct {
		name            string
		role            *Role
		createOperation bool
		wantErr         bool
		expectedErrMsg  string
	}{
		{
			name: "valid role on create",
			role: &Role{
				TenantID:    "tenant-123",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				Permissions: []string{"users:read", "users:write"},
			},
			createOperation: true,
			wantErr:         false,
		},
		{
			name: "valid role on update",
			role: &Role{
				ID:          primitive.NewObjectID(),
				TenantID:    "tenant-123",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				Permissions: []string{"users:read", "users:write"},
			},
			createOperation: false,
			wantErr:         false,
		},
		{
			name: "missing ID on update",
			role: &Role{
				TenantID:    "tenant-123",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				Permissions: []string{"users:read"},
			},
			createOperation: false,
			wantErr:         true,
			expectedErrMsg:  "ID",
		},
		{
			name: "missing tenantID",
			role: &Role{
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				Permissions: []string{"users:read"},
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "TenantID",
		},
		{
			name: "missing name",
			role: &Role{
				TenantID:    "tenant-123",
				Status:      "active",
				CreatedBy:   "admin",
				Permissions: []string{"users:read"},
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "Name",
		},
		{
			name: "missing status",
			role: &Role{
				TenantID:    "tenant-123",
				Name:        "Admin",
				CreatedBy:   "admin",
				Permissions: []string{"users:read"},
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "Status",
		},
		{
			name: "missing createdBy",
			role: &Role{
				TenantID:    "tenant-123",
				Name:        "Admin",
				Status:      "active",
				Permissions: []string{"users:read"},
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "CreatedBy",
		},
		{
			name: "nil permissions",
			role: &Role{
				TenantID:  "tenant-123",
				Name:      "Admin",
				Status:    "active",
				CreatedBy: "admin",
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "Permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.role.Validate(tt.createOperation)
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
