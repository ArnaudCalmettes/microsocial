package actions

import (
	"github.com/ArnaudCalmettes/microsocial/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// ReportsCreate reports given user
func ReportsCreate(c buffalo.Context) error {
	report := &models.Report{}
	if err := c.Bind(report); err != nil {
		return c.Error(400, err)
	}

	user_id, _ := getCredentials(c)
	report.ByID = uuid.FromStringOrNil(user_id)

	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("No transaction found"))
	}

	verrs, err := report.Create(tx)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		return c.Error(400, verrs)
	}

	return c.Render(201, r.JSON(report))
}

// ReportsList lists available reports
func ReportsList(c buffalo.Context) error {
	_, is_admin := getCredentials(c)
	if !is_admin {
		return c.Error(403, errors.New("Forbidden"))
	}

	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	reports := &models.Reports{}
	q := tx.PaginateFromParams(c.Params())
	if err := q.All(reports); err != nil {
		return errors.WithStack(err)
	}

	c.Set("pagination", q.Paginator)
	return c.Render(200, r.JSON(reports))
}
