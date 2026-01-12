package main

import "secretKeeper/internal/client/app/commands"

// Функция точки входа в клиент
func main() {
	if err := commands.Execute(); err != nil {
		panic(err)
	}
}
