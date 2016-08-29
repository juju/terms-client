package api

import (
	"os"

	jujutesting "github.com/juju/testing"
	gc "gopkg.in/check.v1"
)

var _ = gc.Suite(&envSuite{})

type envSuite struct {
	jujutesting.OsEnvSuite
}

func (s *envSuite) TestDefault(c *gc.C) {
	c.Assert(BaseURL(), gc.Equals, defaultURL)
}

func (s *envSuite) TestEnvOverrides(c *gc.C) {
	os.Setenv("JUJU_TERMS", "something-else")
	c.Assert(BaseURL(), gc.Equals, "something-else")
}
