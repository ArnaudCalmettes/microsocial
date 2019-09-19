package actions

import (
	"encoding/json"
	"time"

	"github.com/ArnaudCalmettes/microsocial/models"
)

func (as *ActionSuite) Test_Reports_Create() {
	user := as.createRandomUser()
	subject := as.createRandomUser()
	url := "/reports"
	payload := map[string]string{
		"about_id": subject.ID.String(),
		"info":     "This user is a jerk!",
	}

	// Unauthorized
	resp := as.JSON(url).Post(payload)
	as.Equal(401, resp.Code)

	user_token, err := newToken(user, time.Minute)
	as.NoError(err)

	subject_token, err := newToken(subject, time.Minute)
	as.NoError(err)

	// "Subject" tries to file a report on himself
	resp = as.createAuthRequest(url, subject_token).Post(payload)
	as.Equal(400, resp.Code)

	// User files a report on Subject
	resp = as.createAuthRequest(url, user_token).Post(payload)
	as.Equalf(201, resp.Code, resp.Body.String())

	report := &models.Report{}
	err = json.Unmarshal(resp.Body.Bytes(), report)
	as.NoError(err)
	as.False(report.CreatedAt.IsZero())
	as.Equal(user.ID, report.ByID)
	as.Equal(subject.ID, report.AboutID)
	as.Equal("This user is a jerk!", report.Info)

	count, err := as.DB.Count("reports")
	as.NoError(err)
	as.Equal(1, count)
}

func (as *ActionSuite) Test_Reports_List() {
	user := as.createRandomUser()
	other := as.createRandomUser()
	admin := as.createRandomUser()
	admin.Admin = true
	admin.Update(as.DB)

	infos := []string{
		"This user is a jerk",
		"Really",
		"Please do something",
		"I can't take it anymore",
		"Do you guys even read this?",
	}

	// Create a bunch of reports
	for _, info := range infos {
		report := &models.Report{
			ByID:    user.ID,
			AboutID: other.ID,
			Info:    info,
		}
		verrs, err := report.Create(as.DB)
		as.NoError(err)
		as.Falsef(verrs.HasAny(), verrs.String())
	}

	count, err := as.DB.Count("reports")
	as.NoError(err)
	as.Equal(len(infos), count)

	url := "/reports"

	// Unauthorized
	resp := as.JSON(url).Get()
	as.Equal(401, resp.Code)

	// Non-admin
	token, err := newToken(user, time.Minute)
	resp = as.createAuthRequest(url, token).Get()
	as.Equal(403, resp.Code)

	// Anybody with admin credentials
	token, err = newToken(admin, time.Minute)
	resp = as.createAuthRequest(url, token).Get()
	as.Equal(200, resp.Code)

	reports := models.Reports{}
	err = json.Unmarshal(resp.Body.Bytes(), &reports)
	as.NoError(err)

	as.Equal(len(infos), len(reports))
}
