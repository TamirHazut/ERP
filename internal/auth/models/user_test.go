package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name            string
		user            *User
		createOperation bool
		wantErr         bool
		expectedErrMsg  string
	}{
		{
			name: "valid user on create",
			user: &User{
				TenantID:     "tenant-123",
				Email:        "test@example.com",
				PasswordHash: "hashed-password",
				Status:       UserStatusActive,
				CreatedBy:    "admin",
				Roles:        []UserRole{{RoleID: "role-1", TenantID: "tenant-123"}},
			},
			createOperation: true,
			wantErr:         false,
		},
		{
			name: "valid user on update",
			user: &User{
				ID:           primitive.NewObjectID(),
				TenantID:     "tenant-123",
				Email:        "test@example.com",
				PasswordHash: "hashed-password",
				Status:       UserStatusActive,
				CreatedBy:    "admin",
				Roles:        []UserRole{{RoleID: "role-1", TenantID: "tenant-123"}},
			},
			createOperation: false,
			wantErr:         false,
		},
		{
			name: "missing ID on update",
			user: &User{
				TenantID:     "tenant-123",
				Email:        "test@example.com",
				PasswordHash: "hashed-password",
				Status:       UserStatusActive,
				CreatedBy:    "admin",
				Roles:        []UserRole{{RoleID: "role-1", TenantID: "tenant-123"}},
			},
			createOperation: false,
			wantErr:         true,
			expectedErrMsg:  "ID",
		},
		{
			name: "missing tenantID",
			user: &User{
				Email:        "test@example.com",
				PasswordHash: "hashed-password",
				Status:       UserStatusActive,
				CreatedBy:    "admin",
				Roles:        []UserRole{{RoleID: "role-1"}},
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "TenantID",
		},
		{
			name: "missing email",
			user: &User{
				TenantID:     "tenant-123",
				PasswordHash: "hashed-password",
				Status:       UserStatusActive,
				CreatedBy:    "admin",
				Roles:        []UserRole{{RoleID: "role-1", TenantID: "tenant-123"}},
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "Email",
		},
		{
			name: "missing password hash",
			user: &User{
				TenantID:  "tenant-123",
				Email:     "test@example.com",
				Status:    UserStatusActive,
				CreatedBy: "admin",
				Roles:     []UserRole{{RoleID: "role-1", TenantID: "tenant-123"}},
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "PasswordHash",
		},
		{
			name: "missing status",
			user: &User{
				TenantID:     "tenant-123",
				Email:        "test@example.com",
				PasswordHash: "hashed-password",
				CreatedBy:    "admin",
				Roles:        []UserRole{{RoleID: "role-1", TenantID: "tenant-123"}},
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "Status",
		},
		{
			name: "missing createdBy",
			user: &User{
				TenantID:     "tenant-123",
				Email:        "test@example.com",
				PasswordHash: "hashed-password",
				Status:       UserStatusActive,
				Roles:        []UserRole{{RoleID: "role-1", TenantID: "tenant-123"}},
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "CreatedBy",
		},
		{
			name: "nil roles",
			user: &User{
				TenantID:     "tenant-123",
				Email:        "test@example.com",
				PasswordHash: "hashed-password",
				Status:       UserStatusActive,
				CreatedBy:    "admin",
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "Roles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate(tt.createOperation)
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
