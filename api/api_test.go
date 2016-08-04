// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	stdtesting "testing"
	"time"

	"github.com/juju/romulus/api/terms"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/terms-client/api"
	"github.com/juju/terms-client/api/wireformat"
)

type apiSuite struct {
	client     api.Client
	httpClient *mockHttpClient
}

func Test(t *stdtesting.T) {
	gc.TestingT(t)
}

var _ = gc.Suite(&apiSuite{})

func (s *apiSuite) SetUpTest(c *gc.C) {
	s.httpClient = &mockHttpClient{}
	var err error
	s.client, err = api.NewClient(api.HTTPClient(s.httpClient))
	c.Assert(err, jc.ErrorIsNil)
}

func (s *apiSuite) TestPublish(c *gc.C) {
	termID := struct {
		TermID string `json:"term-id"`
	}{
		TermID: "test-owner/test-term/17",
	}
	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, termID)

	id, err := s.client.Publish("test-owner", "test-term", 17)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(id, gc.Equals, termID.TermID)
	s.httpClient.CheckCall(c, 0, "DoWithBody", "https://api.jujucharms.com/terms/v1/terms/test-owner/test-term/17/publish")
}

func (s *apiSuite) TestSaveOwnedTerm(c *gc.C) {
	term := wireformat.TermIDResponse{
		TermID: "test-term/1",
	}
	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, term)
	savedTerm, err := s.client.SaveTerm("owner", "test-term", "You hereby agree to run this test.")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(savedTerm, jc.DeepEquals, term.TermID)
	s.httpClient.CheckCall(c, 0, "DoWithBody", "https://api.jujucharms.com/terms/v1/terms/owner/test-term")
}

func (s *apiSuite) TestSaveOwnerlessTerm(c *gc.C) {
	term := wireformat.TermIDResponse{
		TermID: "test-term/1",
	}
	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, term)
	savedTerm, err := s.client.SaveTerm("", "test-term", "You hereby agree to run this test.")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(savedTerm, jc.DeepEquals, term.TermID)
	s.httpClient.CheckCall(c, 0, "DoWithBody", "https://api.jujucharms.com/terms/v1/terms/test-term")
}

func (s *apiSuite) TestSaveTermError(c *gc.C) {
	s.httpClient.status = http.StatusInternalServerError
	s.httpClient.SetBody(c, struct {
		Error string `json:"error"`
	}{"silly internal error"})
	_, err := s.client.SaveTerm("", "test-term", "You hereby agree to run this test.")
	c.Assert(err, gc.ErrorMatches, "silly internal error")
}

func (s *apiSuite) TestGetOwnedTermWithRevision(c *gc.C) {
	t := time.Now().Round(time.Second).UTC()
	term := []wireformat.Term{{
		Owner:     "owner",
		Name:      "test-term",
		Revision:  17,
		CreatedOn: wireformat.TimeRFC3339(t),
		Content:   "You hereby agree to run this test.",
	}}
	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, term)
	savedTerm, err := s.client.GetTerm("owner", "test-term", 17)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(savedTerm, jc.DeepEquals, &term[0])
	s.httpClient.CheckCall(c, 0, "Do", "https://api.jujucharms.com/terms/v1/terms/owner/test-term?revision=17")
}

func (s *apiSuite) TestGetOwnedTermWithoutRevision(c *gc.C) {
	t := time.Now().Round(time.Second).UTC()
	term := []wireformat.Term{{
		Owner:     "owner",
		Name:      "test-term",
		Revision:  17,
		CreatedOn: wireformat.TimeRFC3339(t),
		Content:   "You hereby agree to run this test.",
	}}
	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, term)
	savedTerm, err := s.client.GetTerm("owner", "test-term", 0)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(savedTerm, jc.DeepEquals, &term[0])
	s.httpClient.CheckCall(c, 0, "Do", "https://api.jujucharms.com/terms/v1/terms/owner/test-term")
}

func (s *apiSuite) TestGetOwnerlessTermWithRevision(c *gc.C) {
	t := time.Now().Round(time.Second).UTC()
	term := []wireformat.Term{{
		Name:      "test-term",
		Revision:  17,
		CreatedOn: wireformat.TimeRFC3339(t),
		Content:   "You hereby agree to run this test.",
	}}
	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, term)
	savedTerm, err := s.client.GetTerm("", "test-term", 17)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(savedTerm, jc.DeepEquals, &term[0])
	s.httpClient.CheckCall(c, 0, "Do", "https://api.jujucharms.com/terms/v1/terms/test-term?revision=17")
}

func (s *apiSuite) TestGetOwnerlessTermWithoutRevision(c *gc.C) {
	t := time.Now().Round(time.Second).UTC()
	term := []wireformat.Term{{
		Name:      "test-term",
		Revision:  17,
		CreatedOn: wireformat.TimeRFC3339(t),
		Content:   "You hereby agree to run this test.",
	}}
	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, term)
	savedTerm, err := s.client.GetTerm("", "test-term", 0)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(savedTerm, jc.DeepEquals, &term[0])
	s.httpClient.CheckCall(c, 0, "Do", "https://api.jujucharms.com/terms/v1/terms/test-term")
}

func (s *apiSuite) TestGetTermError(c *gc.C) {
	s.httpClient.status = http.StatusInternalServerError
	s.httpClient.SetBody(c, struct {
		Error string `json:"error"`
	}{"silly internal error"})
	_, err := s.client.GetTerm("", "test-term", 17)
	c.Assert(err, gc.ErrorMatches, "silly internal error")
}

func (s *apiSuite) TestSignedAgreements(c *gc.C) {
	t := time.Now().Round(time.Second).UTC()
	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, []wireformat.Agreement{{
		User:      "test-user",
		Term:      "hello-world-terms",
		Revision:  1,
		CreatedOn: wireformat.TimeRFC3339(t),
	}, {
		User:      "test-user",
		Term:      "hello-universe-terms",
		Revision:  42,
		CreatedOn: wireformat.TimeRFC3339(t),
	},
	})
	signedAgreements, err := s.client.GetUsersAgreements()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(signedAgreements, gc.HasLen, 2)
	c.Assert(signedAgreements[0].User, gc.Equals, "test-user")
	c.Assert(signedAgreements[0].Term, gc.Equals, "hello-world-terms")
	c.Assert(signedAgreements[0].Revision, gc.Equals, 1)
	c.Assert(signedAgreements[0].CreatedOn, gc.DeepEquals, t)
	c.Assert(signedAgreements[1].User, gc.Equals, "test-user")
	c.Assert(signedAgreements[1].Term, gc.Equals, "hello-universe-terms")
	c.Assert(signedAgreements[1].Revision, gc.Equals, 42)
	c.Assert(signedAgreements[1].CreatedOn, gc.DeepEquals, t)
}

func (s *apiSuite) TestUnsignedTerms(c *gc.C) {
	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, []wireformat.Term{{
		Name:     "hello-world-terms",
		Revision: 1,
		Content:  "terms doc content",
	}, {
		Name:     "hello-universe-terms",
		Revision: 1,
		Content:  "universal terms doc content",
	},
	})
	missingAgreements, err := s.client.GetUnsignedTerms(&terms.CheckAgreementsRequest{
		Terms: []string{
			"hello-world-terms/1",
			"hello-universe-terms/1",
		},
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(missingAgreements, gc.HasLen, 2)
	c.Assert(missingAgreements[0].Name, gc.Equals, "hello-world-terms")
	c.Assert(missingAgreements[0].Revision, gc.Equals, 1)
	c.Assert(missingAgreements[0].Content, gc.Equals, "terms doc content")
	c.Assert(missingAgreements[1].Name, gc.Equals, "hello-universe-terms")
	c.Assert(missingAgreements[1].Revision, gc.Equals, 1)
	c.Assert(missingAgreements[1].Content, gc.Equals, "universal terms doc content")
	s.httpClient.SetBody(c, terms.SaveAgreementResponses{
		Agreements: []terms.AgreementResponse{{
			User:     "test-user",
			Term:     "hello-world-terms",
			Revision: 1,
		}}},
	)

	p1 := terms.SaveAgreements{
		Agreements: []terms.SaveAgreement{{
			TermName:     "hello-world-terms",
			TermRevision: 1,
		}},
	}
	response, err := s.client.SaveAgreement(&p1)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response.Agreements, gc.HasLen, 1)
	c.Assert(response.Agreements[0].User, gc.Equals, "test-user")
	c.Assert(response.Agreements[0].Term, gc.Equals, "hello-world-terms")
	c.Assert(response.Agreements[0].Revision, gc.Equals, 1)
}

func (s *apiSuite) TestNotFoundError(c *gc.C) {
	s.httpClient.status = http.StatusNotFound
	s.httpClient.body = []byte("something failed")
	_, err := s.client.GetUnsignedTerms(&terms.CheckAgreementsRequest{
		Terms: []string{
			"hello-world-terms/1",
			"hello-universe-terms/1",
		},
	})
	c.Assert(err, gc.ErrorMatches, "failed to get unsigned terms: Not Found: something failed")
}

func (s *apiSuite) TestSignedAgreementsEnvTermsURL(c *gc.C) {
	cleanup := testing.PatchEnvironment("JUJU_TERMS", "http://example.com")
	defer cleanup()

	t := time.Now().UTC()
	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, []terms.AgreementResponse{
		{
			User:      "test-user",
			Term:      "hello-world-terms",
			Revision:  1,
			CreatedOn: t,
		},
		{
			User:      "test-user",
			Term:      "hello-universe-terms",
			Revision:  42,
			CreatedOn: t,
		},
	})
	_, err := s.client.GetUsersAgreements()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s.httpClient.Calls(), gc.HasLen, 1)
	s.httpClient.CheckCall(c, 0, "Do", "http://example.com/agreements")
}

func (s *apiSuite) TestUnsignedTermsEnvTermsURL(c *gc.C) {
	cleanup := testing.PatchEnvironment("JUJU_TERMS", "http://example.com")
	defer cleanup()

	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, []terms.GetTermsResponse{
		{
			Name:     "hello-world-terms",
			Revision: 1,
			Content:  "terms doc content",
		},
		{
			Name:     "hello-universe-terms",
			Revision: 1,
			Content:  "universal terms doc content",
		},
	})
	_, err := s.client.GetUnsignedTerms(&terms.CheckAgreementsRequest{
		Terms: []string{
			"hello-world-terms/1",
			"hello-universe-terms/1",
		},
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s.httpClient.Calls(), gc.HasLen, 1)
	s.httpClient.CheckCall(c, 0, "Do", "http://example.com/agreement?Terms=hello-world-terms%2F1&Terms=hello-universe-terms%2F1")
	s.httpClient.ResetCalls()
}

func (s *apiSuite) TestSaveAgreementEnvTermsURL(c *gc.C) {
	cleanup := testing.PatchEnvironment("JUJU_TERMS", "http://example.com")
	defer cleanup()

	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, terms.SaveAgreementResponses{})
	p1 := terms.SaveAgreements{
		Agreements: []terms.SaveAgreement{{
			TermName:     "hello-world-terms",
			TermRevision: 1,
		}},
	}
	_, err := s.client.SaveAgreement(&p1)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s.httpClient.Calls(), gc.HasLen, 1)
	s.httpClient.CheckCall(c, 0, "DoWithBody", "http://example.com/agreement")
}

type mockHttpClient struct {
	testing.Stub
	status int
	body   []byte
}

func (m *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	m.AddCall("Do", req.URL.String())
	return &http.Response{
		Status:     http.StatusText(m.status),
		StatusCode: m.status,
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Body:       ioutil.NopCloser(bytes.NewReader(m.body)),
	}, nil
}

func (m *mockHttpClient) DoWithBody(req *http.Request, body io.ReadSeeker) (*http.Response, error) {
	m.AddCall("DoWithBody", req.URL.String())
	return &http.Response{
		Status:     http.StatusText(m.status),
		StatusCode: m.status,
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Body:       ioutil.NopCloser(bytes.NewReader(m.body)),
	}, nil
}

func (m *mockHttpClient) SetBody(c *gc.C, v interface{}) {
	b, err := json.Marshal(&v)
	c.Assert(err, jc.ErrorIsNil)
	m.body = b
}
