package hooks

import (
	"github.com/go-viper/mapstructure/v2"
)

var registeredHooks = []mapstructure.DecodeHookFunc{}

func RegisterHook(hook mapstructure.DecodeHookFunc) {
	if hook == nil {
		return
	}
	registeredHooks = append(registeredHooks, hook)
}

func GetExtraHooks() []mapstructure.DecodeHookFunc {
	return registeredHooks
}
