package cli

import (
	"fmt"

	"se-software-design-hw-2/internal/domain"
)

func (c *Console) showTotal(calculate func(domain.StudentID) (domain.Total, error)) error {
	text, err := c.readLine("ID студента: ")
	if err != nil {
		return err
	}
	id, err := parseID(text, "студента")
	if err != nil {
		return c.report(err)
	}
	total, err := calculate(domain.StudentID(id))
	if err != nil {
		return c.report(err)
	}
	label := "Окончательный итог"
	if total.Preliminary() {
		label = "Предварительный итог"
	}
	_, err = fmt.Fprintf(c.output, "%s студента ID=%d: %.2f из 10 | оценок: %d/%d\n",
		label, total.StudentID(), total.Value(), total.GradedCount(), total.CriterionCount())
	return err
}
