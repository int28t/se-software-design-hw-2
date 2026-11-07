package cli_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"se-software-design-hw-2/internal/adapters/cli"
	"se-software-design-hw-2/internal/adapters/memory"
	"se-software-design-hw-2/internal/application"
)

type resultExporterStub struct {
	destination string
	rows        []application.ResultRow
	calls       int
	err         error
}

func (e *resultExporterStub) Export(destination string, rows []application.ResultRow) error {
	e.destination, e.rows = destination, rows
	e.calls++
	return e.err
}

func resultService(t *testing.T, exporter application.ResultExporter) *application.Service {
	t.Helper()
	service := application.NewService(memory.NewRepository(), exporter)
	for _, name := range []string{"Анна", "Борис"} {
		if _, err := service.AddStudent(name); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := service.AddCriterion("Экзамен", 10, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddEvaluation(1, 1, 8); err != nil {
		t.Fatal(err)
	}
	return service
}

func TestResultConsoleListsAndExports(t *testing.T) {
	exporter := &resultExporterStub{}
	var output bytes.Buffer
	if err := cli.New(resultService(t, exporter), strings.NewReader("11\n12\nresults.csv\n0\n"), &output).Run(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"ID=1 | Анна | итог=8.00 из 10 | оценок: 1/1",
		"ID=2 | Борис | итог=неполная ведомость | оценок: 0/1",
		"Итоги экспортированы в CSV: results.csv",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("missing %q:\n%s", want, &output)
		}
	}
	if exporter.calls != 1 || exporter.destination != "results.csv" || len(exporter.rows) != 2 || exporter.rows[1].Total != nil {
		t.Fatalf("unexpected exporter call: %#v", exporter)
	}
}

func TestExportFailureDoesNotEndSession(t *testing.T) {
	exporter := &resultExporterStub{err: errors.New("write failure")}
	var output bytes.Buffer
	if err := cli.New(resultService(t, exporter), strings.NewReader("12\nresults.csv\n11\n0\n"), &output).Run(); err != nil {
		t.Fatal(err)
	}
	if strings.Count(output.String(), "Ошибка:") != 1 || strings.Contains(output.String(), "Итоги экспортированы") || !strings.Contains(output.String(), "ID=1 | Анна | итог=8.00") {
		t.Fatalf("export failure was hidden or ended session:\n%s", &output)
	}
}

func TestExportEOFDoesNotCallExporter(t *testing.T) {
	exporter := &resultExporterStub{}
	var output bytes.Buffer
	if err := cli.New(resultService(t, exporter), strings.NewReader("12\n"), &output).Run(); err != nil {
		t.Fatal(err)
	}
	if exporter.calls != 0 || strings.Contains(output.String(), "Ошибка:") || strings.Contains(output.String(), "Итоги экспортированы") {
		t.Fatalf("EOF triggered export:\n%s", &output)
	}
}
