// Copyright 2022 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package cmd_test

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/terms-client/v2/cmd"
)

type TermSuite struct{}

var _ = gc.Suite(&TermSuite{})

func (s *TermSuite) TestTermStringRoundTrip(c *gc.C) {
	terms := []string{
		"foobar",
		"foobar/27",
		"owner/foobar/27",
		"owner/foobar",
		"owner/foo-bar",
		"own-er/foobar",
		"ibm/j9-jvm/2",
		"cs:foobar/27",
	}
	for i, term := range terms {
		c.Logf("test %d: %s", i, term)
		id, err := cmd.ParseTerm(term)
		c.Check(err, gc.IsNil)
		c.Check(id.String(), gc.Equals, term)
	}
}

func (s *TermSuite) TestParseTerms(c *gc.C) {
	tests := []struct {
		about       string
		term        string
		expectError string
		expectTerm  cmd.TermsId
	}{{
		about:      "valid term",
		term:       "term/1",
		expectTerm: cmd.TermsId{"", "", "term", 1},
	}, {
		about:      "valid term no revision",
		term:       "term",
		expectTerm: cmd.TermsId{"", "", "term", 0},
	}, {
		about:       "revision not a number",
		term:        "term/a",
		expectError: `wrong term name format "a"`,
	}, {
		about:       "negative revision",
		term:        "term/-1",
		expectError: "negative term revision",
	}, {
		about:       "bad revision",
		term:        "owner/term/12a",
		expectError: `invalid revision number "12a" strconv.Atoi: parsing "12a": invalid syntax`,
	}, {
		about:       "wrong format",
		term:        "foobar/term/abc/1",
		expectError: `unknown term id format "foobar/term/abc/1"`,
	}, {
		about:      "term with owner",
		term:       "term/abc/1",
		expectTerm: cmd.TermsId{"", "term", "abc", 1},
	}, {
		about:      "term with owner no rev",
		term:       "term/abc",
		expectTerm: cmd.TermsId{"", "term", "abc", 0},
	}, {
		about:       "term may not contain spaces",
		term:        "term about a term",
		expectError: `wrong term name format "term about a term"`,
	}, {
		about:       "term name must not start with a number",
		term:        "1Term/1",
		expectError: `wrong term name format "1Term"`,
	}, {
		about:      "full term with tenant",
		term:       "tenant:owner/term/1",
		expectTerm: cmd.TermsId{"tenant", "owner", "term", 1},
	}, {
		about:       "bad tenant",
		term:        "tenant::owner/term/1",
		expectError: `wrong owner format ":owner"`,
	}, {
		about:      "ownerless term with tenant",
		term:       "tenant:term/1",
		expectTerm: cmd.TermsId{"tenant", "", "term", 1},
	}, {
		about:      "ownerless revisionless term with tenant",
		term:       "tenant:term",
		expectTerm: cmd.TermsId{"tenant", "", "term", 0},
	}, {
		about:      "owner/term with tenant",
		term:       "tenant:owner/term",
		expectTerm: cmd.TermsId{"tenant", "owner", "term", 0},
	}, {
		about:      "term with tenant",
		term:       "tenant:term",
		expectTerm: cmd.TermsId{"tenant", "", "term", 0},
	}}
	for i, test := range tests {
		c.Logf("running test %v: %v", i, test.about)
		term, err := cmd.ParseTerm(test.term)
		if test.expectError == "" {
			c.Check(err, jc.ErrorIsNil)
			c.Check(term, gc.DeepEquals, &test.expectTerm)
		} else {
			c.Check(err, gc.ErrorMatches, test.expectError)
			c.Check(term, gc.IsNil)
		}
	}
}
