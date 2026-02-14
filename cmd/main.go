package main

import (
	server_auth "erp.localhost/auth/cmd/utils"
	server_init "erp.localhost/init/cmd/utils"
)

func main() {
	server_init.Main()
	server_auth.Main()
	// server_core.Main()
}
