package models

func (ms *ModelSuite) Test_Report_Create() {
	count, err := ms.DB.Count("reports")
	ms.NoError(err)
	ms.Equal(0, count)

	user := ms.createRandomUser()
	other := ms.createRandomUser()

	report := &Report{
		UserID:    user.ID,
		SubjectID: other.ID,
		Message:   "This guy is mean to me!",
	}

	verrs, err := report.Create(ms.DB)
	ms.NoError(err)
	ms.Falsef(verrs.HasAny(), verrs.String())
	ms.NotZero(report.ID)

	count, err = ms.DB.Count("reports")
	ms.NoError(err)
	ms.Equal(1, count)
}

func (ms *ModelSuite) Test_Report_Validate() {
	user := ms.createRandomUser()
	other := ms.createRandomUser()

	report := &Report{
		UserID:  user.ID,
		Message: "Some message",
	}

	verrs, err := report.Validate(ms.DB)
	ms.NoError(err)
	ms.Truef(verrs.HasAny(), "Expected validation error (missing subject_id)")

	report = &Report{
		SubjectID: user.ID,
		Message:   "Some message",
	}

	verrs, err = report.Validate(ms.DB)
	ms.NoError(err)
	ms.Truef(verrs.HasAny(), "Expected validation error (missing user_id)")

	report = &Report{
		UserID:    user.ID,
		SubjectID: other.ID,
	}
	ms.NoError(err)
	ms.Truef(verrs.HasAny(), "Expected validation error (missing message)")

	report = &Report{
		UserID:    user.ID,
		SubjectID: user.ID,
		Message:   "Some message",
	}

	verrs, err = report.Validate(ms.DB)
	ms.NoError(err)
	ms.Truef(verrs.HasAny(), "Expected validation error (subject_id == user_id)")
}
