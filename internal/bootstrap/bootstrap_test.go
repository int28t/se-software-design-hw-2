package bootstrap_test

import (
	"bytes"
	"strings"
	"testing"

	"se-software-design-hw-2/internal/bootstrap"
)

func TestContainerBuildsWorkingApplication(t *testing.T) {
	input := "1\nАнна\n3\nЭкзамен\n10\n1\n5\n1\n1\n8\n6\n1\n9\n1\n0\n"
	var output bytes.Buffer
	console, err := bootstrap.NewConsole(strings.NewReader(input), &output)
	if err != nil {
		t.Fatalf("cannot resolve application graph: %v", err)
	}
	if err := console.Run(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Добавлен студент: ID=1 | Анна",
		"Добавлен критерий: ID=1 | Экзамен | максимум=10 | вес=1",
		"Оценки студента ID=1:\nКритерий ID=1 | балл=8",
		"Окончательный итог студента ID=1: 8.00 из 10 | оценок: 1/1",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("graph did not wire scenarios to the same repository, missing %q:\n%s", want, &output)
		}
	}
}

func TestContainersHaveIndependentRepositories(t *testing.T) {
	var firstOutput bytes.Buffer
	first, err := bootstrap.NewConsole(strings.NewReader("1\nАнна\n0\n"), &firstOutput)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Run(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(firstOutput.String(), "Добавлен студент: ID=1 | Анна") {
		t.Fatalf("first session did not create a student:\n%s", &firstOutput)
	}
	var secondOutput bytes.Buffer
	second, err := bootstrap.NewConsole(strings.NewReader("2\n0\n"), &secondOutput)
	if err != nil {
		t.Fatal(err)
	}
	if err := second.Run(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(secondOutput.String(), "Студентов пока нет.") || strings.Contains(secondOutput.String(), "Анна") {
		t.Fatalf("separate graphs share mutable state:\n%s", &secondOutput)
	}
}
