package actions

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/gobuffalo/suite"
)

type ActionSuite struct {
	*suite.Action
}

func Test_ActionSuite(t *testing.T) {
	gofakeit.Seed(time.Now().UnixNano())
	action, err := suite.NewAction(App())
	if err != nil {
		t.Fatal(err)
	}

	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}
