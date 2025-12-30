package cmd

import auth_server "erp.localhost/internal/auth/server"

// TODO: when breaking to microservices, this will be the entry point for the auth service
func Main() {
	auth_server.StartGRPCServer()
}
