package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"se-software-design-hw-2/internal/adapters/cli"
	"se-software-design-hw-2/internal/adapters/memory"
	"se-software-design-hw-2/internal/application"
)

func TestTotalConsoleReflectsGradeAndRubricChanges(t *testing.T) {
	service := application.NewService(memory.NewRepository())
	input := "1\nАнна\n3\nПроект\n10\n1\n3\nЭкзамен\n20\n3\n5\n1\n1\n8\n5\n1\n2\n10\n9\n1\n7\n1\n1\n10\n9\n1\n8\n1\n2\n9\n1\n10\n1\n3\nАктивность\n5\n1\n10\n1\n5\n1\n2\n20\n5\n1\n3\n5\n9\n1\n0\n"
	var output bytes.Buffer
	if err := cli.New(service, strings.NewReader(input), &output).Run(); err != nil {
		t.Fatal(err)
	}
	remaining := output.String()
	for _, want := range []string{
		"Окончательный итог студента ID=1: 5.75 из 10 | оценок: 2/2",
		"Окончательный итог студента ID=1: 6.25 из 10 | оценок: 2/2",
		"Ошибка: для окончательного итога нужны оценки по всем критериям: выставлено 1 из 2",
		"Предварительный итог студента ID=1: 2.50 из 10 | оценок: 1/2",
		"Предварительный итог студента ID=1: 2.00 из 10 | оценок: 1/3",
		"Окончательный итог студента ID=1: 10.00 из 10 | оценок: 3/3",
	} {
		_, after, found := strings.Cut(remaining, want)
		if !found {
			t.Fatalf("missing expected output in order: %q\n%s", want, &output)
		}
		remaining = after
	}
}

func TestTotalConsoleHandlesInvalidStudentAndEOF(t *testing.T) {
	for _, command := range []string{"9", "10"} {
		for _, id := range []string{"abc", "0", "-1", "99"} {
			t.Run(command+"/"+id, func(t *testing.T) {
				service := evaluationService(t)
				if _, err := service.AddEvaluation(1, 1, 8); err != nil {
					t.Fatal(err)
				}
				var output bytes.Buffer
				input := command + "\n" + id + "\n9\n1\n0\n"
				if err := cli.New(service, strings.NewReader(input), &output).Run(); err != nil {
					t.Fatal(err)
				}
				if strings.Count(output.String(), "Ошибка:") != 1 || !strings.Contains(output.String(), "Окончательный итог студента ID=1: 8.00 из 10 | оценок: 1/1") {
					t.Fatalf("invalid ID ended session or changed total:\n%s", &output)
				}
			})
		}
		t.Run(command+"/EOF", func(t *testing.T) {
			var output bytes.Buffer
			if err := cli.New(evaluationService(t), strings.NewReader(command+"\n"), &output).Run(); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(output.String(), "Ошибка:") || strings.Contains(output.String(), "итог студента ID=") {
				t.Fatalf("EOF was treated as a completed request:\n%s", &output)
			}
		})
	}
}

func TestTotalConsoleRoundsOnlyTheOutput(t *testing.T) {
	service := application.NewService(memory.NewRepository())
	if _, err := service.AddStudent("Анна"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddCriterion("Проект", 3, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddEvaluation(1, 1, 1); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := cli.New(service, strings.NewReader("9\n1\n0\n"), &output).Run(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "3.33 из 10") {
		t.Fatalf("unexpected rounding:\n%s", &output)
	}
	total, err := service.Total(1)
	if err != nil {
		t.Fatal(err)
	}
	if total.Value() == 3.33 {
		t.Fatal("presentation rounding leaked into the domain")
	}
}
