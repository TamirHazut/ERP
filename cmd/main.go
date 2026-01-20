package main

import (
	server_auth "erp.localhost/internal/auth/cmd"
	server_init "erp.localhost/internal/init/cmd"
)

func main() {
	server_init.Main()
	server_auth.Main()
	// server_core.Main()
}
