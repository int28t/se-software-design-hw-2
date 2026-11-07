// Package cli provides the interactive console adapter for grading use cases.
package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"se-software-design-hw-2/internal/application"
)

const menu = `
1 — Добавить студента
2 — Список студентов
3 — Добавить критерий
4 — Список критериев
5 — Выставить оценку
6 — Оценки студента
7 — Исправить оценку
8 — Отменить оценку
9 — Окончательный итог
10 — Предварительный итог
11 — Ведомость итогов
12 — Экспорт итогов в CSV
0 — Выход
`

// Console reads commands and presents the results of application use cases.
type Console struct {
	service *application.Service
	input   *bufio.Scanner
	output  io.Writer
}

// endOfInput marks normal scanner exhaustion rather than an arbitrary stream error.
type endOfInput struct{}

// Error describes normal exhaustion of the input stream.
func (endOfInput) Error() string { return "конец ввода" }

// New creates a console using the supplied service and input/output streams.
func New(service *application.Service, input io.Reader, output io.Writer) *Console {
	return &Console{service: service, input: bufio.NewScanner(input), output: output}
}

// Run processes commands until exit or EOF and returns stream failures.
func (c *Console) Run() error {
	if _, err := fmt.Fprintln(c.output, "ПОС. Данные хранятся в памяти до выхода из программы."); err != nil {
		return err
	}
	for {
		if _, err := fmt.Fprint(c.output, menu); err != nil {
			return err
		}
		choice, err := c.promptLine("Выберите действие: ")
		if _, ended := errors.AsType[endOfInput](err); ended {
			return nil
		}
		if err != nil {
			return err
		}
		if choice == "0" {
			return nil
		}
		if err := c.action(choice); err != nil {
			if _, ended := errors.AsType[endOfInput](err); ended {
				return nil
			}
			return err
		}
	}
}

func (c *Console) action(choice string) error {
	switch choice {
	case "1":
		return c.addStudent()
	case "2":
		return c.listStudents()
	case "3":
		return c.addCriterion()
	case "4":
		return c.listCriteria()
	case "5":
		return c.recordEvaluation(c.service.AddEvaluation, "Оценка выставлена")
	case "6":
		return c.listEvaluations()
	case "7":
		return c.recordEvaluation(c.service.CorrectEvaluation, "Оценка исправлена")
	case "8":
		return c.cancelEvaluation()
	case "9":
		return c.showTotal(c.service.Total)
	case "10":
		return c.showTotal(c.service.PreliminaryTotal)
	case "11":
		return c.listResults()
	case "12":
		return c.exportResults()
	default:
		return c.printError(errors.New("неизвестное действие; выберите число от 0 до 12"))
	}
}

func (c *Console) addStudent() error {
	name, err := c.promptLine("Имя студента: ")
	if err != nil {
		return err
	}
	student, err := c.service.AddStudent(name)
	if err != nil {
		return c.printError(err)
	}
	_, err = fmt.Fprintf(c.output, "Добавлен студент: ID=%d | %s\n", student.ID(), student.Name())
	return err
}

func (c *Console) listStudents() error {
	students, err := c.service.Students()
	if err != nil {
		return c.printError(err)
	}
	if len(students) == 0 {
		_, err := fmt.Fprintln(c.output, "Студентов пока нет.")
		return err
	}
	if _, err := fmt.Fprintln(c.output, "Студенты:"); err != nil {
		return err
	}
	for _, student := range students {
		if _, err := fmt.Fprintf(c.output, "ID=%d | %s\n", student.ID(), student.Name()); err != nil {
			return err
		}
	}
	return nil
}

func (c *Console) addCriterion() error {
	name, err := c.promptLine("Название критерия: ")
	if err != nil {
		return err
	}
	maximum, err := c.promptLine("Максимальный балл (например, 10 или 20.5): ")
	if err != nil {
		return err
	}
	weightText, err := c.promptLine("Вес (например, 1 или 0.25): ")
	if err != nil {
		return err
	}
	maxScore, err := strconv.ParseFloat(maximum, 64)
	if err != nil {
		return c.printError(errors.New("максимальный балл нужно ввести числом с точкой в качестве десятичного разделителя"))
	}
	weight, err := strconv.ParseFloat(weightText, 64)
	if err != nil {
		return c.printError(errors.New("вес нужно ввести числом с точкой в качестве десятичного разделителя"))
	}
	criterion, err := c.service.AddCriterion(name, maxScore, weight)
	if err != nil {
		return c.printError(err)
	}
	_, err = fmt.Fprintf(c.output, "Добавлен критерий: ID=%d | %s | максимум=%g | вес=%g\n",
		criterion.ID(), criterion.Name(), criterion.MaxScore(), criterion.Weight().Value())
	return err
}

func (c *Console) listCriteria() error {
	criteria, err := c.service.Criteria()
	if err != nil {
		return c.printError(err)
	}
	if len(criteria) == 0 {
		_, err := fmt.Fprintln(c.output, "Критериев пока нет.")
		return err
	}
	if _, err := fmt.Fprintln(c.output, "Критерии:"); err != nil {
		return err
	}
	for _, criterion := range criteria {
		if _, err := fmt.Fprintf(c.output, "ID=%d | %s | максимум=%g | вес=%g\n",
			criterion.ID(), criterion.Name(), criterion.MaxScore(), criterion.Weight().Value()); err != nil {
			return err
		}
	}
	return nil
}

// promptLine writes the prompt and reads one trimmed line from the input.
func (c *Console) promptLine(prompt string) (string, error) {
	if _, err := fmt.Fprint(c.output, prompt); err != nil {
		return "", err
	}
	if c.input.Scan() {
		return strings.TrimSpace(c.input.Text()), nil
	}
	if err := c.input.Err(); err != nil {
		return "", fmt.Errorf("чтение ввода: %w", err)
	}
	return "", endOfInput{}
}

// printError displays an operation failure and returns only an output error.
func (c *Console) printError(cause error) error {
	_, err := fmt.Fprintf(c.output, "Ошибка: %v\n", cause)
	return err
}
