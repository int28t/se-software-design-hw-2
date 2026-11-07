package cli

import "fmt"

func (c *Console) listResults() error {
	rows, err := c.service.Results()
	if err != nil {
		return c.printError(err)
	}
	if _, err := fmt.Fprintln(c.output, "Итоги студентов:"); err != nil {
		return err
	}
	for _, row := range rows {
		value := "неполная ведомость"
		if row.Total != nil {
			value = fmt.Sprintf("%.2f из 10", row.Total.Value())
		}
		if _, err := fmt.Fprintf(c.output, "ID=%d | %s | итог=%s | оценок: %d/%d\n",
			row.StudentID, row.StudentName, value, row.GradedCount, row.CriterionCount); err != nil {
			return err
		}
	}
	return nil
}

func (c *Console) exportResults() error {
	destination, err := c.promptLine("Путь нового CSV-файла: ")
	if err != nil {
		return err
	}
	if err := c.service.ExportResults(destination); err != nil {
		return c.printError(err)
	}
	_, err = fmt.Fprintf(c.output, "Итоги экспортированы в CSV: %s\n", destination)
	return err
}
