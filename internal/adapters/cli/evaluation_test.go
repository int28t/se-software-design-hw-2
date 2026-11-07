package cli_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"se-software-design-hw-2/internal/adapters/cli"
	"se-software-design-hw-2/internal/adapters/memory"
	"se-software-design-hw-2/internal/application"
)

func evaluationService(t *testing.T) *application.Service {
	t.Helper()
	service := application.NewService(memory.NewRepository(), nil)
	for _, name := range []string{"Анна", "Борис"} {
		if _, err := service.AddStudent(name); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := service.AddCriterion("Экзамен", 10, 1); err != nil {
		t.Fatal(err)
	}
	return service
}

func TestEvaluationConsoleLifecycle(t *testing.T) {
	service := application.NewService(memory.NewRepository(), nil)
	input := "1\nАнна\n1\nБорис\n3\nЭкзамен\n10\n1\n6\n1\n5\n1\n1\n0\n5\n2\n1\n4.5\n6\n1\n7\n1\n1\n10\n6\n1\n8\n1\n1\n6\n1\n6\n2\n0\n"
	var output bytes.Buffer
	if err := cli.New(service, strings.NewReader(input), &output).Run(); err != nil {
		t.Fatal(err)
	}
	remaining := output.String()
	for _, want := range []string{
		"У студента пока нет оценок.",
		"Оценка выставлена: студент ID=1 | критерий ID=1 | балл=0",
		"Оценка выставлена: студент ID=2 | критерий ID=1 | балл=4.5",
		"Оценки студента ID=1:\nКритерий ID=1 | балл=0",
		"Оценка исправлена: студент ID=1 | критерий ID=1 | балл=10",
		"Оценки студента ID=1:\nКритерий ID=1 | балл=10",
		"Оценка отменена: студент ID=1 | критерий ID=1",
		"У студента пока нет оценок.",
		"Оценки студента ID=2:\nКритерий ID=1 | балл=4.5",
	} {
		_, after, found := strings.Cut(remaining, want)
		if !found {
			t.Fatalf("missing expected output in order: %q\n%s", want, &output)
		}
		remaining = after
	}
	got, err := service.Evaluations(1)
	if err != nil || len(got) != 0 {
		t.Fatalf("cancellation did not persist: %#v, %v", got, err)
	}
}

func TestInvalidEvaluationInputAllowsRetry(t *testing.T) {
	service := evaluationService(t)
	if _, err := service.AddEvaluation(1, 1, 8); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{
		"5\n1\n1\n5\n",   // duplicate
		"5\n99\n1\n5\n",  // unknown student
		"5\n2\n99\n5\n",  // unknown criterion
		"5\nabc\n1\n5\n", // malformed student ID
		"5\n2\n0\n5\n",   // nonpositive criterion ID
		"5\n2\n1\nabc\n", // malformed score
		"5\n2\n1\nNaN\n", // nonfinite score
		"7\n1\n1\n11\n",  // correction outside range
		"7\n2\n1\n5\n",   // correction must not create
		"8\n2\n1\n",      // cancellation must find an evaluation
		"6\n99\n",        // list must find the student
		"6\n-1\n",        // malformed list target
		"8\n1\nabc\n",    // malformed cancellation target
	} {
		t.Run(input, func(t *testing.T) {
			var output bytes.Buffer
			if err := cli.New(service, strings.NewReader(input+"6\n1\n0\n"), &output).Run(); err != nil {
				t.Fatal(err)
			}
			if strings.Count(output.String(), "Ошибка:") != 1 || !strings.Contains(output.String(), "Оценки студента ID=1:\nКритерий ID=1 | балл=8") {
				t.Fatalf("error changed state or ended session:\n%s", &output)
			}
		})
	}
}

func TestEvaluationEOFDoesNotSavePartialForms(t *testing.T) {
	for _, input := range []string{
		"5\n", "5\n1\n", "5\n1\n1\n",
		"6\n", "7\n", "7\n1\n", "7\n1\n1\n",
		"8\n", "8\n1\n",
	} {
		t.Run(input, func(t *testing.T) {
			service := evaluationService(t)
			if _, err := service.AddEvaluation(1, 1, 8); err != nil {
				t.Fatal(err)
			}
			if err := cli.New(service, strings.NewReader(input), io.Discard).Run(); err != nil {
				t.Fatal(err)
			}
			got, err := service.Evaluations(1)
			if err != nil || len(got) != 1 || got[0].Score().Value() != 8 {
				t.Fatalf("partial form changed the evaluation: %#v, %v", got, err)
			}
		})
	}
}
