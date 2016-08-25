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

const publishTermDoc = `
publish-term is used to publish a Terms and Conditions document.
Examples
publish-term me/my-terms
`
const publishTermPurpose = "publishes the given terms document"

// NewPublishTermCommand returns a new command that can be
// used to publish existing owner terms
// Conditions documents.
func NewPublishTermCommand() cmd.Command {
	return WrapPlugin(&publishTermCommand{})
}

type publishTermCommand struct {
	cmd.CommandBase
	out cmd.Output

	TermID               string
	TermsServiceLocation string
}

// SetFlags implements Command.SetFlags.
func (c *publishTermCommand) SetFlags(f *gnuflag.FlagSet) {
	// TODO (mattyw) JUJU_TERMS
	f.StringVar(&c.TermsServiceLocation, "url", defaultTermServiceLocation, "url of the terms service")
	c.out.AddFlags(f, "yaml", cmd.DefaultFormatters)
}

// Info implements Command.Info.
func (c *publishTermCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "publish-term",
		Args:    "<term id>",
		Purpose: publishTermPurpose,
		Doc:     publishTermDoc,
	}
}

// Description returns a one-line description of the command.
func (c *publishTermCommand) Description() string {
	return publishTermPurpose
}

// Init reads and verifies the arguments.
func (c *publishTermCommand) Init(args []string) error {
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
func (c *publishTermCommand) Run(ctx *cmd.Context) error {
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
	if termsId.Owner == "" {
		c.out.Write(ctx, "only terms with owners require publishing")
		return nil
	}

	response, err := termsClient.Publish(
		termsId.Owner,
		termsId.Name,
		termsId.Revision,
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
