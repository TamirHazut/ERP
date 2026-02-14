package server

import (
	"testing"

	"erp.localhost/infra/logging/logger/mock"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
)

func TestRegisterHealthCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := mock.NewMockLogger(ctrl)

	// Expect Info log call
	log.EXPECT().Info("gRPC health check enabled").Times(1)

	// Create a fresh gRPC server
	srv := grpc.NewServer()

	// Should not panic
	RegisterHealthCheck(srv, log)

	// Verify health service is registered
	serviceInfo := srv.GetServiceInfo()
	if _, exists := serviceInfo["grpc.health.v1.Health"]; !exists {
		t.Error("expected health service to be registered")
	}
}

func TestRegisterHealthCheck_ServiceInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := mock.NewMockLogger(ctrl)

	log.EXPECT().Info(gomock.Any()).Times(1)

	srv := grpc.NewServer()
	RegisterHealthCheck(srv, log)

	serviceInfo := srv.GetServiceInfo()

	// Verify the health service is present
	healthInfo, exists := serviceInfo["grpc.health.v1.Health"]
	if !exists {
		t.Fatal("health service not found in service info")
	}

	// Verify it has the Check method (standard health check RPC)
	hasCheckMethod := false
	for _, method := range healthInfo.Methods {
		if method.Name == "Check" {
			hasCheckMethod = true
			break
		}
	}

	if !hasCheckMethod {
		t.Error("expected health service to have Check method")
	}
}
