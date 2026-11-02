package bootstrap_test

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"se-software-design-hw-2/internal/bootstrap"
)

func TestFullConsoleScenarioExportsCurrentTotals(t *testing.T) {
	directory := t.TempDir()
	completePath := filepath.Join(directory, "complete.csv")
	incompletePath := filepath.Join(directory, "incomplete.csv")
	input := fmt.Sprintf("1\nАнна\n1\nБорис\n3\nПроект\n10\n1\n3\nЭкзамен\n20\n3\n5\n1\n1\n8\n5\n1\n2\n10\n5\n2\n1\n10\n5\n2\n2\n20\n9\n1\n11\n12\n%s\n7\n1\n1\n12\n7\n1\n1\n10\n9\n1\n8\n1\n2\n9\n1\n11\n12\n%s\n0\n", completePath, incompletePath)
	var output bytes.Buffer
	console, err := bootstrap.NewConsole(strings.NewReader(input), &output)
	if err != nil {
		t.Fatal(err)
	}
	if err := console.Run(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Окончательный итог студента ID=1: 5.75 из 10",
		"ID=2 | Борис | итог=10.00 из 10",
		"Окончательный итог студента ID=1: 6.25 из 10",
		"ID=1 | Анна | итог=неполная ведомость | оценок: 1/2",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("missing %q:\n%s", want, &output)
		}
	}
	if strings.Count(output.String(), "Ошибка:") != 2 || strings.Count(output.String(), "Итоги экспортированы в CSV:") != 2 {
		t.Fatalf("unexpected results for invalid correction or incomplete total:\n%s", &output)
	}
	header := []string{"student_id", "student_name", "total", "graded_count", "criterion_count", "status"}
	for path, want := range map[string][][]string{
		completePath:   {header, {"1", "Анна", "5.75", "2", "2", "final"}, {"2", "Борис", "10.00", "2", "2", "final"}},
		incompletePath: {header, {"1", "Анна", "", "1", "2", "incomplete"}, {"2", "Борис", "10.00", "2", "2", "final"}},
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		got, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("console and export disagree: got %#v, want %#v", got, want)
		}
	}
}
