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
	action := suite.NewAction(App())

	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}
