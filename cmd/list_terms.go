// Copyright 2016 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/gnuflag"
	"github.com/juju/persistent-cookiejar"
	"gopkg.in/macaroon-bakery.v1/httpbakery"

	"github.com/juju/terms-client/api"
)

const (
	listTermsDoc = `
list-terms show a list of "Terms and Conditions" documents owned by you
or one of the public groups you belong to.
Examples
list-terms --groups test-group1,test-group2
   lists the terms owned by you or one of the public groups you belong to or
   to one of the test-group1 and test-group2, which you must belong to, but may
   not be public.
`
	listTermsPurpose = "list owned terms"

	defaultIDMURL = "https://api.jujucharms.com/identity/v1"
)

// NewListTermsCommand returns a new command that can be used
// to list owned Terms and Conditions documents.
func NewListTermsCommand() cmd.Command {
	return &listTermsCommand{}
}

type listTermsCommand struct {
	cmd.CommandBase
	out cmd.Output

	GroupList            string
	TermsServiceLocation string
}

// SetFlags implements Command.SetFlags.
func (c *listTermsCommand) SetFlags(f *gnuflag.FlagSet) {
	c.out.AddFlags(f, "yaml", cmd.DefaultFormatters)
	f.StringVar(&c.GroupList, "groups", "", "a comma separated list of additional groups")
}

// Info implements Command.Info.
func (c *listTermsCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "list-terms",
		Purpose: listTermsPurpose,
		Doc:     listTermsDoc,
	}
}

// Init reads and verifies the arguments.
func (c *listTermsCommand) Init(args []string) error {
	c.TermsServiceLocation = api.BaseURL()

	if err := cmd.CheckEmpty(args); err != nil {
		return errors.Errorf("unknown arguments: %v", strings.Join(args[1:], ","))
	}

	return nil
}

// Description returns a one-line description of the command.
func (c *listTermsCommand) Description() string {
	return listTermsPurpose
}

// Run implements Command.Run.
func (c *listTermsCommand) Run(ctx *cmd.Context) error {
	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: cookieFile(),
	})
	if err != nil {
		return errors.Trace(err)
	}
	defer jar.Save()
	bakeryClient := httpbakery.NewClient()
	bakeryClient.Jar = jar
	bakeryClient.VisitWebPage = httpbakery.OpenWebBrowser

	idmClient := newIDMClient(idmBaseURL(), bakeryClient)

	// first we perform a whoami request
	username, err := idmClient.WhoAmI()
	if err != nil {
		return errors.Trace(err)
	}

	// the we get the public groups the user belongs to
	userGroups, err := idmClient.Groups(username)
	if err != nil {
		return errors.Trace(err)
	}

	groups := make(map[string]bool, len(userGroups))
	for _, groupName := range userGroups {
		groups[groupName] = true
	}
	if c.GroupList != "" {
		privateGroups := strings.Split(c.GroupList, ",")
		for _, groupName := range privateGroups {
			g := strings.TrimSpace(groupName)
			groups[g] = true
		}
	}
	groupSlice := make([]string, len(groups))
	i := 0
	for g, _ := range groups {
		groupSlice[i] = g
		i++
	}
	sort.Strings(groupSlice)

	termsClient, err := clientNew(
		api.ServiceURL(c.TermsServiceLocation),
		api.HTTPClient(bakeryClient),
	)
	if err != nil {
		return errors.Trace(err)
	}

	terms := []string{}
	for _, groupName := range groupSlice {
		groupTerms, err := termsClient.GetTermsByOwner(groupName)
		if err != nil {
			return errors.Trace(err)
		}
		for _, t := range groupTerms {
			terms = append(terms, t.Id)
		}
	}

	err = c.out.Write(ctx, terms)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

var newIDMClient = func(idmURL string, client *httpbakery.Client) IDMClient {
	return &idmClient{
		idmURL: idmURL,
		client: client,
	}
}

// IDMClient defines the interface of an identity client used by the list-terms
// command.
type IDMClient interface {
	// WhoAmI returns the user's username as declared by the identity managed.
	WhoAmI() (string, error)
	// Groups returns public grous the specified user belongs to.
	Groups(username string) ([]string, error)
}

type idmClient struct {
	idmURL string
	client *httpbakery.Client
}

// WhoAmI returns the user's username as declared by the identity managed.
func (c *idmClient) WhoAmI() (string, error) {
	fail := func(err error) (string, error) {
		return "", err
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/whoami", c.idmURL), nil)
	if err != nil {
		return fail(errors.Trace(err))
	}
	response, err := c.client.Do(req)
	if err != nil {
		return fail(errors.Trace(err))
	}
	var username struct {
		User string `json:"user"`
	}
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&username)
	if err != nil {
		return fail(errors.Trace(err))
	}
	return username.User, nil
}

// Groups returns public grous the specified user belongs to.
func (c *idmClient) Groups(username string) ([]string, error) {
	fail := func(err error) ([]string, error) {
		return nil, err
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/u/%s/groups", c.idmURL, username), nil)
	if err != nil {
		return fail(errors.Trace(err))
	}
	response, err := c.client.Do(req)
	if err != nil {
		return fail(errors.Trace(err))
	}
	var groups []string
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&groups)
	if err != nil {
		return fail(errors.Trace(err))
	}
	return groups, nil
}

func idmBaseURL() string {
	baseURL := defaultIDMURL
	if idmURL := os.Getenv("IDM_URL"); idmURL != "" {
		baseURL = idmURL
	}
	return baseURL
}
