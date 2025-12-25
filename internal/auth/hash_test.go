package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "valid strong password", password: "1aAm!&25@*zgTY$pwL", wantErr: false},
		{name: "valid weak password", password: "password", wantErr: true},
		{name: "invalidempty password", password: "", wantErr: true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := HashPassword(tc.password)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.True(t, VerifyPassword(tc.password, hash))
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{name: "valid password", password: "password", hash: "$2a$10$YxNnIaPMWRFglNffZjPEv.mJoa63BZWObp2yjHC7P6/aG61C.mJyC", want: true},
		{name: "empty password", password: "", hash: "$2a$10$YxNnIaPMWRFglNffZjPEv.mJoa63BZWObp2yjHC7P6/aG61C.mJyC", want: false},
		{name: "invalid password", password: "invalid", hash: "$2a$10$YxNnIaPMWRFglNffZjPEv.mJoa63BZWObp2yjHC7P6/aG61C.mJyC", want: false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := VerifyPassword(tc.password, tc.hash)
			assert.Equal(t, tc.want, result)
		})
	}
}
