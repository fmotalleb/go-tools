package main

import (
	"encoding/json"
	"fmt"

	"github.com/FMotalleb/go-tools/builders/nested"
)

func main() {
	b := nested.New()
	b.Set("core.test", 15).
		Set("core.code", 55).
		Set("core.killer.test", 55)
	j, _ := json.MarshalIndent(b.Data, "", "  ")
	fmt.Println(string(j))
}
