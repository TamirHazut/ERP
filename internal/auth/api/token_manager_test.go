package api

import (
	"errors"
	"testing"
	"time"

	mock_token "erp.localhost/internal/auth/handler/mock"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	authv1_cache "erp.localhost/internal/infra/model/auth/v1/cache"
	"erp.localhost/internal/infra/model/shared"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTokenManager_StoreTokens(t *testing.T) {
	testCases := []struct {
		name                      string
		tenantID                  string
		userID                    string
		accessTokenMetadata       *authv1_cache.TokenMetadata
		refreshToken              *authv1_cache.RefreshToken
		accessStoreError          error
		refreshStoreError         error
		deleteError               error
		wantErr                   bool
		expectedAccessStoreCalls  int
		expectedRefreshStoreCalls int
		expectedDeleteCalls       int
	}{
		{
			name:     "successful store",
			tenantID: "tenant-1",
			userID:   "user-1",
			accessTokenMetadata: &authv1_cache.TokenMetadata{
				UserId:   "user-1",
				TenantId: "tenant-1",
			},
			refreshToken: &authv1_cache.RefreshToken{
				UserId:    "user-1",
				TenantId:  "tenant-1",
				ExpiresAt: timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
				CreatedAt: timestamppb.Now(),
			},
			accessStoreError:          nil,
			refreshStoreError:         nil,
			wantErr:                   false,
			expectedAccessStoreCalls:  1,
			expectedRefreshStoreCalls: 1,
		},
		{
			name:     "access token store fails",
			tenantID: "tenant-1",
			userID:   "user-1",
			accessTokenMetadata: &authv1_cache.TokenMetadata{
				UserId:   "user-1",
				TenantId: "tenant-1",
			},
			refreshToken: &authv1_cache.RefreshToken{
				UserId:    "user-1",
				TenantId:  "tenant-1",
				ExpiresAt: timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
				CreatedAt: timestamppb.Now(),
			},
			accessStoreError:          infra_error.Internal(infra_error.InternalDatabaseError, errors.New("store failed")),
			refreshStoreError:         nil,
			wantErr:                   true,
			expectedAccessStoreCalls:  1,
			expectedRefreshStoreCalls: 0,
		},
		{
			name:     "refresh token store fails - access token cleaned up",
			tenantID: "tenant-1",
			userID:   "user-1",
			accessTokenMetadata: &authv1_cache.TokenMetadata{
				UserId:   "user-1",
				TenantId: "tenant-1",
			},
			refreshToken: &authv1_cache.RefreshToken{
				UserId:    "user-1",
				TenantId:  "tenant-1",
				ExpiresAt: timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
				CreatedAt: timestamppb.Now(),
			},
			deleteError:               infra_error.Internal(infra_error.InternalDatabaseError, errors.New("delete failed")),
			accessStoreError:          nil,
			refreshStoreError:         infra_error.Internal(infra_error.InternalDatabaseError, errors.New("store failed")),
			wantErr:                   true,
			expectedAccessStoreCalls:  1,
			expectedRefreshStoreCalls: 1,
			expectedDeleteCalls:       1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			accessMock := mock_token.NewMockTokenHandler[authv1_cache.TokenMetadata](ctrl)
			refreshMock := mock_token.NewMockTokenHandler[authv1_cache.RefreshToken](ctrl)

			if tc.expectedAccessStoreCalls > 0 {
				accessMock.EXPECT().
					Store(tc.tenantID, tc.userID, tc.accessTokenMetadata).
					Return(tc.accessStoreError).
					Times(tc.expectedAccessStoreCalls)
			}

			if tc.expectedRefreshStoreCalls > 0 {
				refreshMock.EXPECT().
					Store(tc.tenantID, tc.userID, tc.refreshToken).
					Return(tc.refreshStoreError).
					Times(tc.expectedRefreshStoreCalls)
			}
			if tc.expectedDeleteCalls > 0 {
				accessMock.EXPECT().
					Delete(tc.tenantID, tc.userID).
					Return(tc.deleteError).
					Times(tc.expectedDeleteCalls)
			}

			tm := &TokenAPI{
				accessTokenHandler:  accessMock,
				refreshTokenHandler: refreshMock,
				logger:              logger.NewBaseLogger(shared.ModuleAuth),
			}

			err := tm.StoreTokens(
				tc.tenantID, tc.userID,
				tc.accessTokenMetadata, tc.refreshToken,
			)

			if tc.wantErr {
				require.NotNil(t, err)
				require.Equal(t, err.Category, infra_error.CategoryInternal)
			} else {
				require.Nil(t, err)
			}
		})
	}
}
