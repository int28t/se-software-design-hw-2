package cli

import (
	"errors"
	"fmt"
	"strconv"

	"se-software-design-hw-2/internal/domain"
)

// recordEvaluation shares the input form; the selected application scenario
// determines whether the evaluation must be new or already exist.
func (c *Console) recordEvaluation(
	record func(domain.StudentID, domain.CriterionID, float64) (domain.Evaluation, error),
	message string,
) error {
	studentText, err := c.readLine("ID студента: ")
	if err != nil {
		return err
	}
	criterionText, err := c.readLine("ID критерия: ")
	if err != nil {
		return err
	}
	scoreText, err := c.readLine("Балл (десятичный разделитель — точка): ")
	if err != nil {
		return err
	}
	studentID, criterionID, err := parseEvaluationTarget(studentText, criterionText)
	if err != nil {
		return c.report(err)
	}
	score, err := strconv.ParseFloat(scoreText, 64)
	if err != nil {
		return c.report(errors.New("балл нужно ввести числом с точкой в качестве десятичного разделителя"))
	}
	evaluation, err := record(studentID, criterionID, score)
	if err != nil {
		return c.report(err)
	}
	_, err = fmt.Fprintf(c.output, "%s: студент ID=%d | критерий ID=%d | балл=%g\n",
		message, evaluation.StudentID(), evaluation.CriterionID(), evaluation.Score().Value())
	return err
}

func (c *Console) listEvaluations() error {
	text, err := c.readLine("ID студента: ")
	if err != nil {
		return err
	}
	id, err := parseID(text, "студента")
	if err != nil {
		return c.report(err)
	}
	evaluations, err := c.service.Evaluations(domain.StudentID(id))
	if err != nil {
		return c.report(err)
	}
	if len(evaluations) == 0 {
		_, err := fmt.Fprintln(c.output, "У студента пока нет оценок.")
		return err
	}
	if _, err := fmt.Fprintf(c.output, "Оценки студента ID=%d:\n", id); err != nil {
		return err
	}
	for _, evaluation := range evaluations {
		if _, err := fmt.Fprintf(c.output, "Критерий ID=%d | балл=%g\n",
			evaluation.CriterionID(), evaluation.Score().Value()); err != nil {
			return err
		}
	}
	return nil
}

func (c *Console) cancelEvaluation() error {
	studentText, err := c.readLine("ID студента: ")
	if err != nil {
		return err
	}
	criterionText, err := c.readLine("ID критерия: ")
	if err != nil {
		return err
	}
	studentID, criterionID, err := parseEvaluationTarget(studentText, criterionText)
	if err != nil {
		return c.report(err)
	}
	if err := c.service.CancelEvaluation(studentID, criterionID); err != nil {
		return c.report(err)
	}
	_, err = fmt.Fprintf(c.output, "Оценка отменена: студент ID=%d | критерий ID=%d\n", studentID, criterionID)
	return err
}

func parseEvaluationTarget(studentText, criterionText string) (domain.StudentID, domain.CriterionID, error) {
	studentID, err := parseID(studentText, "студента")
	if err != nil {
		return 0, 0, err
	}
	criterionID, err := parseID(criterionText, "критерия")
	if err != nil {
		return 0, 0, err
	}
	return domain.StudentID(studentID), domain.CriterionID(criterionID), nil
}

func parseID(text, entity string) (int, error) {
	id, err := strconv.Atoi(text)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("ID %s должен быть положительным целым числом", entity)
	}
	return id, nil
}
