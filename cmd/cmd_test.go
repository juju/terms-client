// Copyright 2016 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/errors"
	jujutesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/macaroon-bakery.v1/httpbakery"

	"github.com/juju/terms-client/api"
	"github.com/juju/terms-client/api/wireformat"
	"github.com/juju/terms-client/cmd"
)

func TestPackage(t *testing.T) {
	gc.TestingT(t)
}

var _ = gc.Suite(&commandSuite{})

var testTermsAndConditions = "Test Terms and Conditions"

type commandSuite struct {
	client    *mockClient
	idmClient *mockIDMClient
	cleanup   func()
}

func (s *commandSuite) SetUpTest(c *gc.C) {
	s.client = &mockClient{}
	s.idmClient = &mockIDMClient{
		username: "test-user",
		groups: map[string][]string{
			"test-user": []string{"test-user"},
		},
	}

	c1 := jujutesting.PatchValue(cmd.ClientNew, func(...api.ClientOption) (api.Client, error) {
		return s.client, nil
	})

	c2 := jujutesting.PatchValue(cmd.ReadFile, func(string) ([]byte, error) {
		return []byte(testTermsAndConditions), nil
	})

	c3 := jujutesting.PatchValue(cmd.NewIDMClient, func(_ string, _ *httpbakery.Client) cmd.IDMClient {
		return s.idmClient
	})
	s.cleanup = func() {
		c1()
		c2()
		c3()
	}
}

func (s *commandSuite) TearDownTest(c *gc.C) {
	s.cleanup()
}

func (s *commandSuite) TestPushTerm(c *gc.C) {
	tests := []struct {
		about   string
		args    []string
		err     string
		stdout  string
		apiCall []interface{}
	}{{
		about: "everything works",
		args:  []string{"test.txt", "test-term", "--format", "json"},
		stdout: `"test-term/1"
`,
		apiCall: []interface{}{"", "test-term", testTermsAndConditions},
	}, {
		about: "everything works - with owner",
		args:  []string{"test.txt", "test-owner/test-term", "--format", "json"},
		stdout: `"test-owner/test-term/1"
`,
		apiCall: []interface{}{"test-owner", "test-term", testTermsAndConditions},
	}, {
		about: "invalid termid",
		args:  []string{"test.txt", "!!!!!!", "--format", "json"},
		err:   `invalid term id argument: wrong term name format "!!!!!!"`,
	}, {
		about: "fail if specifying term with revision",
		args:  []string{"test.txt", "test-term/1", "--format", "json"},
		err:   "can't specify a revision with a new term",
	}, {
		about: "fail if specifying term with tenant",
		args:  []string{"test.txt", "cs:test-term", "--format", "json"},
		err:   "can't specify a tenant with a new term",
	}, {
		about: "unknown args",
		args:  []string{"test.txt", "test-term", "unknown", "args", "--format", "json"},
		err:   "unknown arguments: unknown,args",
	}, {
		about: "missing args",
		args:  []string{"test-term", "--format", "json"},
		err:   "missing arguments",
	},
	}
	for i, test := range tests {
		s.client.ResetCalls()
		c.Logf("running test %d: %s", i, test.about)
		ctx, err := cmdtesting.RunCommand(c, cmd.NewPushTermCommand(), test.args...)
		if test.err != "" {
			c.Assert(err, gc.ErrorMatches, test.err)
		} else {
			c.Assert(err, jc.ErrorIsNil)
		}
		if ctx != nil {
			c.Assert(cmdtesting.Stdout(ctx), gc.Equals, test.stdout)
		}
		if len(test.apiCall) > 0 {
			s.client.CheckCall(c, 0, "SaveTerm", test.apiCall...)
		}
	}
}

func (s *commandSuite) TestShowTerm(c *gc.C) {
	s.client.setTerms([]wireformat.Term{{
		Id:       "test-term/1",
		Name:     "test-term",
		Revision: 1,
		Content:  testTermsAndConditions,
	}})
	tests := []struct {
		about   string
		args    []string
		err     string
		stdout  string
		apiCall []interface{}
	}{{
		about: "everything works",
		args:  []string{"test-term/1", "--format", "json"},
		stdout: `{"id":"test-term/1","name":"test-term","revision":1,"created-on":"0001-01-01T00:00:00Z","published":false,"content":"Test Terms and Conditions"}
`,
		apiCall: []interface{}{"", "test-term", 1},
	}, {
		about:   "everything works for content",
		args:    []string{"test-term/1", "--content"},
		stdout:  `Test Terms and Conditions`,
		apiCall: []interface{}{"", "test-term", 1},
	}, {
		about: "everything works in yaml",
		args:  []string{"test-term/1", "--format", "yaml"},
		stdout: `id: test-term/1
name: test-term
revision: 1
createdon: 0001-01-01T00:00:00Z
published: false
content: Test Terms and Conditions
`,
		apiCall: []interface{}{"", "test-term", 1},
	}, {
		about: "cannot parse revision number",
		args:  []string{"owner/test-term/abc", "--format", "json"},
		err:   `invalid term format: invalid revision number "abc" strconv.Atoi: parsing "abc": invalid syntax`,
	}, {
		about: "get latest version",
		args:  []string{"test-term", "--format", "json"},
		stdout: `{"id":"test-term/1","name":"test-term","revision":1,"created-on":"0001-01-01T00:00:00Z","published":false,"content":"Test Terms and Conditions"}
`,
		apiCall: []interface{}{"", "test-term", 0},
	}, {
		about: "unknown arguments",
		args:  []string{"test-term/1", "unknown", "arguments", "--format", "json"},
		err:   "unknown arguments: unknown,arguments",
	}, {
		about: "unknown arguments",
		args:  []string{},
		err:   "missing arguments",
	},
	}
	for i, test := range tests {
		s.client.ResetCalls()
		c.Logf("running test %d: %s", i, test.about)
		ctx, err := cmdtesting.RunCommand(c, cmd.NewShowTermCommand(), test.args...)
		if test.err != "" {
			c.Assert(err, gc.ErrorMatches, test.err)
		} else {
			c.Assert(err, jc.ErrorIsNil)
		}
		if ctx != nil {
			c.Assert(cmdtesting.Stdout(ctx), gc.Equals, test.stdout)
		}
		if len(test.apiCall) > 0 {
			s.client.CheckCall(c, 0, "GetTerm", test.apiCall...)
		}
	}
}

func (s *commandSuite) TestShowTermsWithOwners(c *gc.C) {
	s.client.setTerms([]wireformat.Term{{
		Id:       "owner/test-term/1",
		Owner:    "owner",
		Name:     "test-term",
		Revision: 1,
		Content:  testTermsAndConditions,
	}})
	tests := []struct {
		about   string
		args    []string
		err     string
		stdout  string
		apiCall []interface{}
	}{{
		about: "everything works - with owner",
		args:  []string{"test-owner/test-term/1", "--format", "json"},
		stdout: `{"id":"owner/test-term/1","owner":"owner","name":"test-term","revision":1,"created-on":"0001-01-01T00:00:00Z","published":false,"content":"Test Terms and Conditions"}
`,
		apiCall: []interface{}{"test-owner", "test-term", 1},
	}, {
		about: "parse owner/term-name",
		args:  []string{"test-owner/abc", "--format", "json"},
		stdout: `{"id":"owner/test-term/1","owner":"owner","name":"test-term","revision":1,"created-on":"0001-01-01T00:00:00Z","published":false,"content":"Test Terms and Conditions"}
`,
		apiCall: []interface{}{"test-owner", "abc", 0},
	}, {
		about: "everything works with owners in yaml",
		args:  []string{"owner/test-term/1", "--format", "yaml"},
		stdout: `id: owner/test-term/1
owner: owner
name: test-term
revision: 1
createdon: 0001-01-01T00:00:00Z
published: false
content: Test Terms and Conditions
`,
		apiCall: []interface{}{"owner", "test-term", 1},
	}}
	for i, test := range tests {
		s.client.ResetCalls()
		c.Logf("running test %d: %s", i, test.about)
		ctx, err := cmdtesting.RunCommand(c, cmd.NewShowTermCommand(), test.args...)
		if test.err != "" {
			c.Assert(err, gc.ErrorMatches, test.err)
		} else {
			c.Assert(err, jc.ErrorIsNil)
		}
		if ctx != nil {
			c.Assert(cmdtesting.Stdout(ctx), gc.Equals, test.stdout)
		}
		if len(test.apiCall) > 0 {
			s.client.CheckCall(c, 0, "GetTerm", test.apiCall...)
		}
	}
}

func (s *commandSuite) TestPublishOwnerlessTerm(c *gc.C) {
	ctx, err := cmdtesting.RunCommand(c, cmd.NewReleaseTermCommand(), "test-term/1")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(cmdtesting.Stdout(ctx), gc.Equals, `only terms with owners require releasing
`)
	s.client.CheckNoCalls(c)
}

func (s *commandSuite) TestPublishTerm(c *gc.C) {
	tests := []struct {
		about   string
		args    []string
		err     string
		stdout  string
		apiCall []interface{}
	}{{
		about: "everything works",
		args:  []string{"owner/test-term/1", "--format", "json"},
		stdout: `"owner/name/1"
`,
		apiCall: []interface{}{"owner", "test-term", 1},
	}, {
		about: "unknown args",
		args:  []string{"test-term", "unknown", "args"},
		err:   "unknown arguments: unknown,args",
	}, {
		about: "missing args",
		args:  []string{},
		err:   "missing arguments",
	}, {
		about: "missing revision",
		args:  []string{"owner/test-term/0"},
		err:   `must specify a term revision`,
	},
	}
	for i, test := range tests {
		s.client.ResetCalls()
		c.Logf("running test %d: %s", i, test.about)
		ctx, err := cmdtesting.RunCommand(c, cmd.NewReleaseTermCommand(), test.args...)
		if test.err != "" {
			c.Assert(err, gc.ErrorMatches, test.err)
		} else {
			c.Assert(err, jc.ErrorIsNil)
		}
		if ctx != nil {
			c.Assert(cmdtesting.Stdout(ctx), gc.Equals, test.stdout)
		}
		if len(test.apiCall) > 0 {
			s.client.CheckCall(c, 0, "Publish", test.apiCall...)
		}
	}
}

func (s *commandSuite) TestListTerms(c *gc.C) {
	s.client.setTerms([]wireformat.Term{{
		Id:       "test-user/test-term/1",
		Owner:    "test-user",
		Name:     "test-term",
		Revision: 1,
		Content:  testTermsAndConditions,
	}})
	tests := []struct {
		about            string
		declaredUsername string
		args             []string
		err              string
		stdout           string
		apiCalls         []jujutesting.StubCall
	}{{
		about:            "everything works",
		declaredUsername: "test-user",
		args:             []string{},
		stdout: `- test-user/test-term/1
`,
		apiCalls: []jujutesting.StubCall{
			{FuncName: "GetTermsByOwner", Args: []interface{}{"test-user"}},
		},
	}, {
		about:            "unknown user",
		declaredUsername: "test-unknown-user",
		args:             []string{},
		err:              "user not found",
	}, {
		about:            "test private groups",
		declaredUsername: "test-user",
		args:             []string{"--groups", "private-group1, private-group2"},
		stdout: `- test-user/test-term/1
- test-user/test-term/1
- test-user/test-term/1
`,
		apiCalls: []jujutesting.StubCall{
			{FuncName: "GetTermsByOwner", Args: []interface{}{"private-group1"}},
			{FuncName: "GetTermsByOwner", Args: []interface{}{"private-group2"}},
			{FuncName: "GetTermsByOwner", Args: []interface{}{"test-user"}},
		},
	}}
	for i, test := range tests {
		s.client.ResetCalls()
		s.idmClient.username = test.declaredUsername
		c.Logf("running test %d: %s", i, test.about)
		ctx, err := cmdtesting.RunCommand(c, cmd.NewListTermsCommand(), test.args...)
		if test.err != "" {
			c.Assert(err, gc.ErrorMatches, test.err)
		} else {
			c.Assert(err, jc.ErrorIsNil)
		}
		if ctx != nil {
			c.Assert(cmdtesting.Stdout(ctx), gc.Equals, test.stdout)
		}
		if len(test.apiCalls) > 0 {
			s.client.CheckCalls(c, test.apiCalls)
		}
	}
}

type mockClient struct {
	api.Client
	jujutesting.Stub

	lock          sync.Mutex
	user          string
	terms         []wireformat.Term
	unsignedTerms []wireformat.Term
}

func (c *mockClient) setTerms(t []wireformat.Term) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.terms = t
}

func (c *mockClient) setUnsignedTerms(t []wireformat.Term) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.unsignedTerms = t
}

func (c *mockClient) SaveTerm(owner, name, content string) (string, error) {
	c.AddCall("SaveTerm", owner, name, content)
	if owner == "" {
		return fmt.Sprintf("%s/1", name), nil
	}
	return fmt.Sprintf("%s/%s/1", owner, name), nil
}

// GetTerms returns matching Terms and Conditions documents.
func (c *mockClient) GetTerm(owner, name string, revision int) (*wireformat.Term, error) {
	c.AddCall("GetTerm", owner, name, revision)
	c.lock.Lock()
	defer c.lock.Unlock()

	if len(c.terms) == 0 {
		return nil, errors.NotFoundf("term")
	}
	t := c.terms[0]
	return &t, nil
}

// SaveAgreement saves user's agreement to the specified
// revision of the Terms and Conditions document.s
func (c *mockClient) SaveAgreement(agreements *wireformat.SaveAgreements) (*wireformat.SaveAgreementResponses, error) {
	c.AddCall("SaveAgreement", agreements)
	responses := make([]wireformat.AgreementResponse, len(agreements.Agreements))
	for i, agreement := range agreements.Agreements {
		responses[i] = wireformat.AgreementResponse{
			User:     c.user,
			Owner:    agreement.TermOwner,
			Term:     agreement.TermName,
			Revision: agreement.TermRevision,
		}
	}
	return &wireformat.SaveAgreementResponses{Agreements: responses}, nil
}

func (c *mockClient) GetUnsignedTerms(terms *wireformat.CheckAgreementsRequest) ([]wireformat.GetTermsResponse, error) {
	c.MethodCall(c, "GetUnunsignedTerms", terms)
	r := make([]wireformat.GetTermsResponse, len(c.unsignedTerms))
	for i, term := range c.terms {
		r[i].Owner = term.Owner
		r[i].Name = term.Name
		r[i].Revision = term.Revision
	}
	return r, nil
}

func (c *mockClient) Publish(owner, name string, revision int) (string, error) {
	c.MethodCall(c, "Publish", owner, name, revision)
	return "owner/name/1", nil
}

func (c *mockClient) GetTermsByOwner(owner string) ([]wireformat.Term, error) {
	c.MethodCall(c, "GetTermsByOwner", owner)
	return c.terms, c.NextErr()
}

type mockIDMClient struct {
	jujutesting.Stub
	username string
	groups   map[string][]string
}

func (c *mockIDMClient) WhoAmI() (string, error) {
	c.MethodCall(c, "WhoAmI")
	return c.username, c.NextErr()
}

func (c *mockIDMClient) Groups(username string) ([]string, error) {
	c.MethodCall(c, "Groups", username)
	groups, ok := c.groups[username]
	if !ok {
		return nil, errors.NotFoundf("user")
	}
	return groups, c.NextErr()
}
