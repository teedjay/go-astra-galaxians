package main

import (
	"go-astra-galaxians/internal/game"
	"os"
)

func main() {
	demo := len(os.Args) > 1 && os.Args[1] == "--demo"
	if err := game.Run(demo); err != nil {
		panic(err)
	}
}
