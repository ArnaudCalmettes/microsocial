package actions

import (
	"github.com/ArnaudCalmettes/microsocial/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/pkg/errors"
)

// ReportsCreate reports given user
// @Summary Report a user to the moderators
// @Description Report a user to the moderators
// @security Bearer
// @Accept  json
// @Produce  json
// @Param user_id path string true "The user's ID"
// @Param userinfo body models.LightReport true "Mandatory report information"
// @Success 201 {object} models.Report
// @Failure 400 {object} FormattedError
// @Failure 401 {object} FormattedError
// @Failure 404 {object} FormattedError
// @Failure 409 {object} FormattedError "You can't report yourself"
// @Router /users/{user_id}/report [post]
func ReportsCreate(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("No transaction found"))
	}

	light_report := &models.LightReport{}
	if err := c.Bind(light_report); err != nil {
		return c.Error(400, err)
	}
	report := models.ReportFromLight(light_report)

	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(404, errors.New("This user doesn't exist"))
	}

	auth := getCredentials(c)

	if user.ID == auth.ID {
		return c.Error(409, errors.New("Can't report yourself"))
	}

	report.ByID = auth.ID
	report.AboutID = user.ID

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
// @Summary List available reports (requires admin credentials)
// @Description List available reports (requires admin credentials)
// @security Bearer
// @Produce  json
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} models.Reports
// @Header 200  {object} X-Pagination "pagination information"
// @Failure 401 {object} FormattedError
// @Failure 403 {object} FormattedError
// @Router /reports/ [get]
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
	if err := q.Eager().All(reports); err != nil {
		return errors.WithStack(err)
	}

	c.Set("pagination", q.Paginator)
	return c.Render(200, r.JSON(reports))
}
