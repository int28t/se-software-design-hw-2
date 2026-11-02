// Package bootstrap is the composition root: it binds ports to adapters and
// builds the application without exposing the DI container to business code.
package bootstrap

import (
	"fmt"
	"io"

	"go.uber.org/dig"

	"se-software-design-hw-2/internal/adapters/cli"
	csvadapter "se-software-design-hw-2/internal/adapters/csv"
	"se-software-design-hw-2/internal/adapters/memory"
	"se-software-design-hw-2/internal/application"
)

func NewConsole(input io.Reader, output io.Writer) (*cli.Console, error) {
	container := dig.New()
	for _, constructor := range []any{
		func() (io.Reader, io.Writer) { return input, output },
		func() application.GradebookRepository { return memory.NewRepository() },
		func() application.ResultExporter { return csvadapter.NewExporter() },
		application.NewService,
		cli.New,
	} {
		if err := container.Provide(constructor); err != nil {
			return nil, fmt.Errorf("регистрация зависимостей приложения: %w", err)
		}
	}
	var console *cli.Console
	if err := container.Invoke(func(resolved *cli.Console) { console = resolved }); err != nil {
		return nil, fmt.Errorf("сборка приложения: %w", err)
	}
	return console, nil
}
