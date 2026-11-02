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

type Console struct {
	service *application.Service
	input   *bufio.Scanner
	output  io.Writer
}

func New(service *application.Service, input io.Reader, output io.Writer) *Console {
	return &Console{service: service, input: bufio.NewScanner(input), output: output}
}

func (c *Console) Run() error {
	if _, err := fmt.Fprintln(c.output, "ПОС. Данные хранятся в памяти до выхода из программы."); err != nil {
		return err
	}
	for {
		if _, err := fmt.Fprint(c.output, "\n1 — Добавить студента\n2 — Список студентов\n3 — Добавить критерий\n4 — Список критериев\n5 — Выставить оценку\n6 — Оценки студента\n7 — Исправить оценку\n8 — Отменить оценку\n0 — Выход\n"); err != nil {
			return err
		}
		choice, err := c.readLine("Выберите действие: ")
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if choice == "0" {
			return nil
		}
		if err := c.action(choice); err != nil {
			if errors.Is(err, io.EOF) {
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
	default:
		return c.report(errors.New("неизвестное действие; выберите число от 0 до 8"))
	}
}

func (c *Console) addStudent() error {
	name, err := c.readLine("Имя студента: ")
	if err != nil {
		return err
	}
	student, err := c.service.AddStudent(name)
	if err != nil {
		return c.report(err)
	}
	_, err = fmt.Fprintf(c.output, "Добавлен студент: ID=%d | %s\n", student.ID(), student.Name())
	return err
}

func (c *Console) listStudents() error {
	students, err := c.service.Students()
	if err != nil {
		return c.report(err)
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
	name, err := c.readLine("Название критерия: ")
	if err != nil {
		return err
	}
	maximum, err := c.readLine("Максимальный балл (например, 10 или 20.5): ")
	if err != nil {
		return err
	}
	weightText, err := c.readLine("Вес (например, 1 или 0.25): ")
	if err != nil {
		return err
	}
	maxScore, err := strconv.ParseFloat(maximum, 64)
	if err != nil {
		return c.report(errors.New("максимальный балл нужно ввести числом с точкой в качестве десятичного разделителя"))
	}
	weight, err := strconv.ParseFloat(weightText, 64)
	if err != nil {
		return c.report(errors.New("вес нужно ввести числом с точкой в качестве десятичного разделителя"))
	}
	criterion, err := c.service.AddCriterion(name, maxScore, weight)
	if err != nil {
		return c.report(err)
	}
	_, err = fmt.Fprintf(c.output, "Добавлен критерий: ID=%d | %s | максимум=%g | вес=%g\n",
		criterion.ID(), criterion.Name(), criterion.MaxScore(), criterion.Weight().Value())
	return err
}

func (c *Console) listCriteria() error {
	criteria, err := c.service.Criteria()
	if err != nil {
		return c.report(err)
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

func (c *Console) readLine(prompt string) (string, error) {
	if _, err := fmt.Fprint(c.output, prompt); err != nil {
		return "", err
	}
	if c.input.Scan() {
		return strings.TrimSpace(c.input.Text()), nil
	}
	if err := c.input.Err(); err != nil {
		return "", fmt.Errorf("чтение ввода: %w", err)
	}
	return "", io.EOF
}

// report handles an operation error without ending the console session.
func (c *Console) report(cause error) error {
	_, err := fmt.Fprintf(c.output, "Ошибка: %v\n", cause)
	return err
}
