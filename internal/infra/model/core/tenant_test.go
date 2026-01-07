package core

import (
	"testing"

	model_auth "erp.localhost/internal/infra/model/auth"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestTenant_Validate(t *testing.T) {
	tests := []struct {
		name            string
		tenant          *Tenant
		createOperation bool
		wantErr         bool
		expectedErrMsg  string
	}{
		{
			name: "valid tenant on create",
			tenant: &Tenant{
				Name:      "Test Tenant",
				Status:    model_auth.TenantStatusActive,
				CreatedBy: "admin",
			},
			createOperation: true,
			wantErr:         false,
		},
		{
			name: "valid tenant on update",
			tenant: &Tenant{
				ID:        primitive.NewObjectID(),
				Name:      "Test Tenant",
				Status:    model_auth.TenantStatusActive,
				CreatedBy: "admin",
			},
			createOperation: false,
			wantErr:         false,
		},
		{
			name: "missing ID on update",
			tenant: &Tenant{
				Name:      "Test Tenant",
				Status:    model_auth.TenantStatusActive,
				CreatedBy: "admin",
			},
			createOperation: false,
			wantErr:         true,
			expectedErrMsg:  "ID",
		},
		{
			name: "missing name",
			tenant: &Tenant{
				Status:    model_auth.TenantStatusActive,
				CreatedBy: "admin",
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "Name",
		},
		{
			name: "missing status",
			tenant: &Tenant{
				Name:      "Test Tenant",
				CreatedBy: "admin",
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "Status",
		},
		{
			name: "missing createdBy",
			tenant: &Tenant{
				Name:   "Test Tenant",
				Status: model_auth.TenantStatusActive,
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "CreatedBy",
		},
		{
			name: "multiple missing fields",
			tenant: &Tenant{
				Name: "Test Tenant",
			},
			createOperation: true,
			wantErr:         true,
			expectedErrMsg:  "Status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tenant.Validate(tt.createOperation)
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
