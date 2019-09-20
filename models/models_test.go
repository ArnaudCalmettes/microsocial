package models

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/gobuffalo/suite"
)

type ModelSuite struct {
	*suite.Model
}

func Test_ModelSuite(t *testing.T) {
	gofakeit.Seed(time.Now().UnixNano())
	as := &ModelSuite{suite.NewModel()}
	suite.Run(t, as)
}
