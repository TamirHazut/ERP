package main

import (
	server_auth "erp.localhost/auth/cmd/lib"
	server_init "erp.localhost/init/cmd/lib"
)

func main() {
	server_init.Main()
	server_auth.Main()
	// server_core.Main()
}
