package main

import (
	server_auth "erp.localhost/auth/cmd"
	server_init "erp.localhost/init/cmd"
)

func main() {
	server_init.Main()
	server_auth.Main()
	// server_core.Main()
}
