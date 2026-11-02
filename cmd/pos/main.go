package main

import (
	"fmt"
	"os"

	"se-software-design-hw-2/internal/bootstrap"
)

func main() {
	console, err := bootstrap.NewConsole(os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка запуска:", err)
		os.Exit(1)
	}
	if err := console.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка работы консоли:", err)
		os.Exit(1)
	}
}
