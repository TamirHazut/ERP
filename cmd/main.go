package main

import (
	auth_server "erp.localhost/internal/auth/cmd"
	core_server "erp.localhost/internal/core/cmd"
)

func main() {
	core_server.Main()
	auth_server.Main()
}
