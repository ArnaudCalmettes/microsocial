package grifts

import (
	"github.com/ArnaudCalmettes/microsocial/actions"
	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
