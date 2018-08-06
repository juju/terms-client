// Copyright 2016 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"strings"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/gnuflag"
	"gopkg.in/juju/charm.v6-unstable"

	"github.com/juju/terms-client/api"
)

var (
	clientNew = func(options ...api.ClientOption) (api.Client, error) {
		return api.NewClient(options...)
	}
)

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
	baseCommand
	out cmd.Output

	TermID       string
	TermFilename string
}

// SetFlags implements Command.SetFlags.
func (c *pushTermCommand) SetFlags(f *gnuflag.FlagSet) {
	c.out.AddFlags(f, "yaml", cmd.DefaultFormatters)
	c.baseCommand.SetFlags(f)
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

	bakeryClient, cleanup, err := c.NewClient(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	defer cleanup()
	termsClient, err := clientNew(
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
