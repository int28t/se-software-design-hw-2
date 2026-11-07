package application_test

import (
	"errors"
	"testing"

	"se-software-design-hw-2/internal/application"
	"se-software-design-hw-2/internal/domain"
)

type recordingExporter struct {
	rows        []application.ResultRow
	destination string
	calls       int
	err         error
}

func (e *recordingExporter) Export(destination string, rows []application.ResultRow) error {
	e.calls++
	e.destination = destination
	e.rows = rows
	return e.err
}

func TestResultsUseCurrentSnapshotAndDistinguishMissingTotals(t *testing.T) {
	book := evaluationScenarioBook(t)
	if _, err := book.AddStudent("Борис"); err != nil {
		t.Fatal(err)
	}
	if _, err := book.AddEvaluation(1, 2, 6); err != nil {
		t.Fatal(err)
	}
	repo := &evaluationRepository{book: book}
	exporter := &recordingExporter{}
	service := application.NewService(repo, exporter)
	rows, err := service.Results()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].StudentID != 1 || rows[0].StudentName != "Анна" || rows[0].Total == nil || rows[0].Total.Value() != 7 || rows[0].GradedCount != 2 || rows[0].CriterionCount != 2 {
		t.Fatalf("unexpected completed row: %#v", rows)
	}
	if rows[1].StudentID != 2 || rows[1].StudentName != "Борис" || rows[1].Total != nil || rows[1].GradedCount != 0 || rows[1].CriterionCount != 2 {
		t.Fatalf("missing grades were presented as a zero total: %#v", rows[1])
	}
	if _, err := service.CorrectEvaluation(1, 1, 10); err != nil {
		t.Fatal(err)
	}
	savesBefore := repo.saves
	if err := service.ExportResults("results.csv"); err != nil {
		t.Fatal(err)
	}
	if exporter.calls != 1 || exporter.destination != "results.csv" || exporter.rows[0].Total.Value() != 8 || repo.saves != savesBefore {
		t.Fatal("export used stale results or saved state")
	}
	if err := service.CancelEvaluation(1, 2); err != nil {
		t.Fatal(err)
	}
	rows, err = service.Results()
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].Total != nil || rows[0].GradedCount != 1 {
		t.Fatalf("cancellation was not reflected: %#v", rows[0])
	}
}

func TestExportDoesNotRunWhenReportCannotBeBuilt(t *testing.T) {
	loadErr := errors.New("load failure")
	var noRubric domain.Gradebook
	if _, err := noRubric.AddStudent("Анна"); err != nil {
		t.Fatal(err)
	}
	for _, repo := range []*evaluationRepository{
		{loadErr: loadErr}, {}, {book: noRubric},
	} {
		exporter := &recordingExporter{}
		err := application.NewService(repo, exporter).ExportResults("results.csv")
		if err == nil || exporter.calls != 0 || repo.saves != 0 {
			t.Fatalf("exported an unavailable report: err=%v calls=%d saves=%d", err, exporter.calls, repo.saves)
		}
		if repo.loadErr != nil && !errors.Is(err, loadErr) {
			t.Fatalf("lost load error: %v", err)
		}
	}
}

func TestExportPropagatesExporterFailure(t *testing.T) {
	wantErr := errors.New("export failure")
	repo := &evaluationRepository{book: evaluationScenarioBook(t)}
	exporter := &recordingExporter{err: wantErr}
	err := application.NewService(repo, exporter).ExportResults("results.csv")
	if !errors.Is(err, wantErr) || exporter.calls != 1 || repo.saves != 0 {
		t.Fatalf("export failure lost or changed state: %v", err)
	}
	if err := application.NewService(repo, nil).ExportResults("results.csv"); err == nil {
		t.Fatal("accepted missing exporter")
	}
}
