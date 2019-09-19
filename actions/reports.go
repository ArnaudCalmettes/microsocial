package actions

import (
	"github.com/ArnaudCalmettes/microsocial/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/pkg/errors"
)

// ReportsCreate reports given user
func ReportsCreate(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("No transaction found"))
	}

	report := &models.Report{}
	if err := c.Bind(report); err != nil {
		return c.Error(400, err)
	}

	auth := getCredentials(c)
	report.ByID = auth.ID

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
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	auth := getCredentials(c)
	if !auth.Admin {
		return c.Error(403, errors.New("Forbidden"))
	}

	reports := &models.Reports{}
	q := tx.PaginateFromParams(c.Params())
	if err := q.All(reports); err != nil {
		return errors.WithStack(err)
	}

	c.Set("pagination", q.Paginator)
	return c.Render(200, r.JSON(reports))
}
