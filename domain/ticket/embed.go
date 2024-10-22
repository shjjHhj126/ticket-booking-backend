package ticket

import (
	_ "embed"
	"fmt"
)

//go:embed reserve_seats_return_broadcast_info.lua
var luaBookingScript string

func readLuaBookingScript() (string, error) {
	if luaBookingScript == "" {
		return "", fmt.Errorf("lua booking script is empty")
	}
	return luaBookingScript, nil
}
