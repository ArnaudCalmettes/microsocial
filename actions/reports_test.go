package actions

import (
	"encoding/json"
	"fmt"

	"github.com/ArnaudCalmettes/microsocial/models"
)

func (as *ActionSuite) Test_Reports_Create() {
	user, user_token := as.createUserAndToken(false)
	subject, subject_token := as.createUserAndToken(false)
	payload := map[string]string{"info": "This user is a jerk!"}

	url := fmt.Sprintf("/users/%s/report", subject.ID)

	// Unauthorized
	resp := as.JSON(url).Post(payload)
	as.Equal(401, resp.Code)

	// "Subject" tries to file a report on himself
	resp = as.createAuthRequest(url, subject_token).Post(payload)
	as.Equal(409, resp.Code)

	// User files a report on Subject
	resp = as.createAuthRequest(url, user_token).Post(payload)
	as.Equalf(201, resp.Code, resp.Body.String())

	report := &models.Report{}
	err := json.Unmarshal(resp.Body.Bytes(), report)
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
	_, admin_token := as.createUserAndToken(true)
	user, user_token := as.createUserAndToken(false)
	other := as.createRandomUser()

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
	resp = as.createAuthRequest(url, user_token).Get()
	as.Equal(403, resp.Code)

	// Anybody with admin credentials
	resp = as.createAuthRequest(url, admin_token).Get()
	as.Equal(200, resp.Code)

	reports := models.Reports{}
	err = json.Unmarshal(resp.Body.Bytes(), &reports)
	as.NoError(err)

	as.Equal(len(infos), len(reports))
}
