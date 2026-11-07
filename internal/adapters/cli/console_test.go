package cli_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"se-software-design-hw-2/internal/adapters/cli"
	"se-software-design-hw-2/internal/adapters/memory"
	"se-software-design-hw-2/internal/application"
)

func TestStudentAndCriterionConsoleScenario(t *testing.T) {
	service := application.NewService(memory.NewRepository(), nil)
	var output bytes.Buffer
	input := "2\n4\n1\nАнна Иванова\n1\nБорис Петров\n3\nПроект\n20\n3\n3\nЭкзамен\n10.5\n0.25\n2\n4\n0\n"
	if err := cli.New(service, strings.NewReader(input), &output).Run(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Студентов пока нет.", "Критериев пока нет.",
		"\nID=1 | Анна Иванова\n", "\nID=2 | Борис Петров\n",
		"\nID=1 | Проект | максимум=20 | вес=3\n",
		"\nID=2 | Экзамен | максимум=10.5 | вес=0.25\n",
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("missing %q in output:\n%s", want, &output)
		}
	}
	students, err := service.Students()
	if err != nil || len(students) != 2 {
		t.Fatalf("students: %v, error: %v", students, err)
	}
	criteria, err := service.Criteria()
	if err != nil || len(criteria) != 2 {
		t.Fatalf("criteria: %v, error: %v", criteria, err)
	}
}

func TestBadInputDoesNotEndSession(t *testing.T) {
	service := application.NewService(memory.NewRepository(), nil)
	var output bytes.Buffer
	input := "wrong\n1\n   \n3\nПроект\nabc\n1\n3\nПроект\n10\nabc\n3\nПроект\nNaN\n1\n3\nПроект\n10\n-1\n1\nАнна\n3\nПроект\n10\n1\n0\n"
	if err := cli.New(service, strings.NewReader(input), &output).Run(); err != nil {
		t.Fatal(err)
	}
	if strings.Count(output.String(), "Ошибка:") != 6 {
		t.Fatalf("expected six rejected operations:\n%s", &output)
	}
	students, err := service.Students()
	if err != nil || len(students) != 1 || students[0].ID() != 1 {
		t.Fatalf("invalid operations changed students: %v, %v", students, err)
	}
	criteria, err := service.Criteria()
	if err != nil || len(criteria) != 1 || criteria[0].ID() != 1 {
		t.Fatalf("invalid operations changed criteria: %v, %v", criteria, err)
	}
}

func TestEOFDoesNotCommitIncompleteInput(t *testing.T) {
	for _, input := range []string{"", "1\n", "3\n", "3\nПроект\n", "3\nПроект\n10\n"} {
		t.Run(input, func(t *testing.T) {
			service := application.NewService(memory.NewRepository(), nil)
			if err := cli.New(service, strings.NewReader(input), io.Discard).Run(); err != nil {
				t.Fatal(err)
			}
			students, err := service.Students()
			if err != nil || len(students) != 0 {
				t.Fatalf("incomplete student was committed: %v, %v", students, err)
			}
			criteria, err := service.Criteria()
			if err != nil || len(criteria) != 0 {
				t.Fatalf("incomplete criterion was committed: %v, %v", criteria, err)
			}
		})
	}
}

type failingIO struct{ err error }

func (f failingIO) Read([]byte) (int, error) { return 0, f.err }

func (f failingIO) Write([]byte) (int, error) { return 0, f.err }

func TestIOErrors(t *testing.T) {
	wantErr := errors.New("I/O failure")
	service := application.NewService(memory.NewRepository(), nil)
	if err := cli.New(service, failingIO{wantErr}, io.Discard).Run(); !errors.Is(err, wantErr) {
		t.Fatalf("lost reader error: %v", err)
	}
	if err := cli.New(service, strings.NewReader("0\n"), failingIO{wantErr}).Run(); !errors.Is(err, wantErr) {
		t.Fatalf("lost writer error: %v", err)
	}
}

type failingMessageWriter struct {
	message string
	err     error
}

func (w failingMessageWriter) Write(data []byte) (int, error) {
	if strings.Contains(string(data), w.message) {
		return 0, w.err
	}
	return len(data), nil
}

func TestOutputEOFIsNotTreatedAsEndOfInput(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   string
		message string
	}{
		{name: "menu prompt", input: "0\n", message: "Выберите действие:"},
		{name: "form prompt", input: "1\nАнна\n0\n", message: "Имя студента:"},
		{name: "confirmation", input: "1\nАнна\n0\n", message: "Добавлен студент:"},
		{name: "operation error", input: "99\n0\n", message: "Ошибка:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, wantErr := range []error{io.EOF, fmt.Errorf("output failure: %w", io.EOF)} {
				service := application.NewService(memory.NewRepository(), nil)
				output := failingMessageWriter{message: tc.message, err: wantErr}
				err := cli.New(service, strings.NewReader(tc.input), output).Run()
				if !errors.Is(err, wantErr) {
					t.Fatalf("output failure was swallowed: got %v, want %v", err, wantErr)
				}
			}
		})
	}
}

func TestWrappedInputEOFIsReported(t *testing.T) {
	wantErr := fmt.Errorf("input failure: %w", io.EOF)
	service := application.NewService(memory.NewRepository(), nil)
	err := cli.New(service, failingIO{err: wantErr}, io.Discard).Run()
	if !errors.Is(err, wantErr) {
		t.Fatalf("input failure was swallowed: got %v, want %v", err, wantErr)
	}
}
