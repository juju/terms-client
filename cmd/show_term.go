// Copyright 2016 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"strings"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/persistent-cookiejar"
	"gopkg.in/juju/charm.v6-unstable"
	"gopkg.in/macaroon-bakery.v1/httpbakery"
	"launchpad.net/gnuflag"

	"github.com/juju/terms-client/api"
)

const showTermDoc = `
show-term is used to show a specific Terms and Conditions document.
Examples
show-term enterprise-plan/1
   shows revision 1 of the enterprise-plan Terms and Conditions.
show-term enterprise-plan
   shows the latest revision of the enterprise plan Terms and Conditions.   
`

// NewShowTermCommand returns a new command that can be used
// to shows Terms and Conditions document.
func NewShowTermCommand() *showTermCommand {
	return &showTermCommand{}
}

type showTermCommand struct {
	cmd.CommandBase
	out cmd.Output

	TermID               string
	TermsServiceLocation string
}

// SetFlags implements Command.SetFlags.
func (c *showTermCommand) SetFlags(f *gnuflag.FlagSet) {
	// TODO (mattyw) Use JUJU_TERMS
	f.StringVar(&c.TermsServiceLocation, "url", defaultTermServiceLocation, "url of the terms service")
	c.out.AddFlags(f, "yaml", cmd.DefaultFormatters)
}

// Info implements Command.Info.
func (c *showTermCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "show-term",
		Args:    "<term id>",
		Purpose: "shows the specified term",
		Doc:     showTermDoc,
	}
}

// Init reads and verifies the arguments.
func (c *showTermCommand) Init(args []string) error {
	if len(args) < 1 {
		return errors.New("missing arguments")
	}

	c.TermID = args[0]
	if err := cmd.CheckEmpty(args[1:]); err != nil {
		return errors.Errorf("unknown arguments: %v", strings.Join(args[1:], ","))
	}

	return nil
}

// Run implements Command.Run.
func (c *showTermCommand) Run(ctx *cmd.Context) error {
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
	termsId, err := charm.ParseTerm(c.TermID)
	if err != nil {
		return errors.Annotate(err, "invalid term format")
	}

	response, err := termsClient.GetTerm(termsId.Owner, termsId.Name, termsId.Revision)
	if err != nil {
		return errors.Trace(err)
	}

	err = c.out.Write(ctx, response)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
