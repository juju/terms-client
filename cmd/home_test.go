// Copyright 2021 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package cmd_test

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/terms-client/v2/cmd"
)

type JujuXDGDataHomeSuite struct {}

var _ = gc.Suite(&JujuXDGDataHomeSuite{})

func (s *JujuXDGDataHomeSuite) TearDownTest(c *gc.C) {
	cmd.SetJujuXDGDataHome("")
}

func (s *JujuXDGDataHomeSuite) TestStandardHome(c *gc.C) {
	testJujuXDGDataHome := c.MkDir()
	cmd.SetJujuXDGDataHome(testJujuXDGDataHome)
	c.Assert(cmd.JujuXDGDataHome(), gc.Equals, testJujuXDGDataHome)
}

func (s *JujuXDGDataHomeSuite) TestHomePath(c *gc.C) {
	testJujuHome := c.MkDir()
	cmd.SetJujuXDGDataHome(testJujuHome)
	envPath := cmd.JujuXDGDataHomeDir()
	c.Assert(envPath, gc.Equals, testJujuHome)
}
