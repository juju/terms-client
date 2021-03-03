// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package wireformat_test

import (
	"encoding/json"
	stdtesting "testing"
	"time"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"

	"github.com/juju/terms-client/v2/api/wireformat"
)

func Test(t *stdtesting.T) {
	gc.TestingT(t)
}

type wireformatSuite struct{}

var _ = gc.Suite(&wireformatSuite{})

func (s *wireformatSuite) TestYAML(c *gc.C) {
	t := time.Date(2016, 1, 2, 4, 8, 16, 32, time.UTC)
	agreement := wireformat.Agreement{
		User:      "test-user",
		Owner:     "test-owner",
		Term:      "test-term",
		Revision:  17,
		CreatedOn: wireformat.TimeRFC3339(t),
	}
	data, err := yaml.Marshal(agreement)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(string(data), jc.DeepEquals, `user: test-user
owner: test-owner
term: test-term
revision: 17
createdon: "2016-01-02T04:08:16Z"
`)
	var agreementOut wireformat.Agreement
	err = yaml.Unmarshal(data, &agreementOut)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(agreement.User, gc.Equals, agreementOut.User)
	c.Assert(agreement.Owner, gc.Equals, agreementOut.Owner)
	c.Assert(agreement.Term, gc.Equals, agreementOut.Term)
	c.Assert(agreement.Revision, gc.Equals, agreementOut.Revision)
	c.Assert(time.Time(agreement.CreatedOn).Truncate(time.Second), gc.Equals, time.Time(agreementOut.CreatedOn))
}

func (s *wireformatSuite) TestJSON(c *gc.C) {
	t := time.Date(2016, 1, 2, 4, 8, 16, 32, time.UTC)
	agreement := wireformat.Agreement{
		User:      "test-user",
		Owner:     "test-owner",
		Term:      "test-term",
		Revision:  17,
		CreatedOn: wireformat.TimeRFC3339(t),
	}
	data, err := json.Marshal(agreement)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(string(data), jc.DeepEquals, `{"user":"test-user","owner":"test-owner","term":"test-term","revision":17,"created-on":"2016-01-02T04:08:16Z"}`)
	var agreementOut wireformat.Agreement
	err = json.Unmarshal(data, &agreementOut)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(agreement.User, gc.Equals, agreementOut.User)
	c.Assert(agreement.Owner, gc.Equals, agreementOut.Owner)
	c.Assert(agreement.Term, gc.Equals, agreementOut.Term)
	c.Assert(agreement.Revision, gc.Equals, agreementOut.Revision)
	c.Assert(time.Time(agreement.CreatedOn).Truncate(time.Second), gc.Equals, time.Time(agreementOut.CreatedOn))
}
