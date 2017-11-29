// Copyright 2016 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/gnuflag"
	"github.com/juju/persistent-cookiejar"
	"github.com/juju/utils"
	"gopkg.in/juju/charm.v6-unstable"
	"gopkg.in/macaroon-bakery.v1/httpbakery"

	"github.com/juju/terms-client/api"
)

var readFile = ioutil.ReadFile

const pushTermDoc = `
push-term is used to create a new Terms and Conditions document.
Examples
push-term text.txt user/enterprise-plan
   creates a new Terms and Conditions with the content from 
   file text.txt and the name enterprise-plan and 
   returns the revision of the created document.
`
const pushTermPurpose = "create new Terms and Conditions document (revision)"

// NewPushTermCommand returns a new command that can be
// used to create new (revisions) of Terms and
// Conditions documents.
func NewPushTermCommand() cmd.Command {
	return &pushTermCommand{}
}

// pushTermCommand creates a new Terms and Conditions document.
type pushTermCommand struct {
	cmd.CommandBase
	out cmd.Output

	TermID               string
	TermFilename         string
	TermsServiceLocation string
}

// SetFlags implements Command.SetFlags.
func (c *pushTermCommand) SetFlags(f *gnuflag.FlagSet) {
	c.out.AddFlags(f, "yaml", cmd.DefaultFormatters)
}

// Info implements Command.Info.
func (c *pushTermCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "push-term",
		Args:    "<filename> <term id>",
		Purpose: pushTermPurpose,
		Doc:     pushTermDoc,
	}
}

// Init read and verifies the arguments.
func (c *pushTermCommand) Init(args []string) error {
	c.TermsServiceLocation = api.BaseURL()
	if len(args) < 2 {
		return errors.New("missing arguments")
	}

	fn, id, args := args[0], args[1], args[2:]

	if err := cmd.CheckEmpty(args); err != nil {
		return errors.Errorf("unknown arguments: %v", strings.Join(args, ","))
	}

	c.TermID = id
	c.TermFilename = fn
	return nil
}

// Description returns a one-line description of the command.
func (c *pushTermCommand) Description() string {
	return pushTermPurpose
}

// Run implements Command.Run.
func (c *pushTermCommand) Run(ctx *cmd.Context) error {
	termid, err := charm.ParseTerm(c.TermID)
	if err != nil {
		return errors.Annotatef(err, "invalid term id argument")
	}
	if termid.Revision > 0 {
		return errors.Errorf("can't specify a revision with a new term")
	}
	if termid.Tenant != "" {
		return errors.Errorf("can't specify a tenant with a new term")
	}
	data, err := readFile(c.TermFilename)
	if err != nil {
		return errors.Annotatef(err, "could not read contents of %q", c.TermFilename)
	}

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

	termsClient, err := clientNew(
		api.ServiceURL(c.TermsServiceLocation),
		api.HTTPClient(bakeryClient),
	)
	if err != nil {
		return errors.Trace(err)
	}

	response, err := termsClient.SaveTerm(
		termid.Owner,
		termid.Name,
		string(data),
	)
	if err != nil {
		return errors.Trace(err)
	}

	err = c.out.Write(ctx, response)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// cookieFile returns the path to the cookie used to store authorization
// macaroons. The returned value can be overridden by setting the
// JUJU_COOKIEFILE environment variable.
func cookieFile() string {
	if file := os.Getenv("JUJU_COOKIEFILE"); file != "" {
		return file
	}
	return path.Join(utils.Home(), ".go-cookies")
}
