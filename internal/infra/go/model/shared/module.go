package shared

import "strings"

type Module string

const (
	ModuleAuth    Module = "Auth"
	ModuleConfig  Module = "Config"
	ModuleCore    Module = "Core"
	ModuleDB      Module = "DB"
	ModuleEvent   Module = "Event"
	ModuleGateway Module = "Gateway"
	ModuleInit    Module = "Init"
	ModuleSidecar Module = "Sidecar"
	ModuleWebUI   Module = "WebUI"
)

func IsValidModule(module string) bool {
	validModules := map[string]bool{
		strings.ToLower(string(ModuleAuth)):    true,
		strings.ToLower(string(ModuleConfig)):  true,
		strings.ToLower(string(ModuleCore)):    true,
		strings.ToLower(string(ModuleDB)):      true,
		strings.ToLower(string(ModuleEvent)):   true,
		strings.ToLower(string(ModuleGateway)): true,
		strings.ToLower(string(ModuleInit)):    true,
		strings.ToLower(string(ModuleSidecar)): true,
		strings.ToLower(string(ModuleWebUI)):   true,
	}
	return validModules[strings.ToLower(module)]
}
