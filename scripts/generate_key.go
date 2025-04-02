package main

import (
	"fmt"
	"postui_api/pkg/auth"
)

func main() {
	fmt.Println(auth.GenerateRandomKey())
}
