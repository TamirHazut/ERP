package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			result := VerifyHash(tc.password, tc.hash)
			assert.Equal(t, tc.want, result)
		})
	}
}
