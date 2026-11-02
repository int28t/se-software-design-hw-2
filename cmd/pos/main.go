package main

import (
	"fmt"
	"os"

	"se-software-design-hw-2/internal/adapters/cli"
	"se-software-design-hw-2/internal/adapters/memory"
	"se-software-design-hw-2/internal/application"
)

func main() {
	repository := memory.NewRepository()
	service := application.NewService(repository)
	console := cli.New(service, os.Stdin, os.Stdout)
	if err := console.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка работы консоли:", err)
		os.Exit(1)
	}
}
