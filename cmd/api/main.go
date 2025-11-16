package main

import (
	"avito/internal/config"
	"fmt"
)

func main() {
	fmt.Println("Hello World")

	_ = config.GetConfig()
}
