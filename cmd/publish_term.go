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

const publishTermDoc = `
release-term is used to release a Terms and Conditions document.
Examples
release-term me/my-terms
`
const publishTermPurpose = "releases the given terms document"

// NewReleaseTermCommand returns a new command that can be
// used to publish existing owner terms
// Conditions documents.
func NewReleaseTermCommand() cmd.Command {
	return &releaseTermCommand{}
}

type releaseTermCommand struct {
	baseCommand
	out cmd.Output

	TermID string
}

// SetFlags implements Command.SetFlags.
func (c *releaseTermCommand) SetFlags(f *gnuflag.FlagSet) {
	c.out.AddFlags(f, "yaml", cmd.DefaultFormatters)
	c.baseCommand.SetFlags(f)
}

// Info implements Command.Info.
func (c *releaseTermCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "release-term",
		Args:    "<term id>",
		Purpose: publishTermPurpose,
		Doc:     publishTermDoc,
	}
}

// Description returns a one-line description of the command.
func (c *releaseTermCommand) Description() string {
	return publishTermPurpose
}

// Init reads and verifies the arguments.
func (c *releaseTermCommand) Init(args []string) error {
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
func (c *releaseTermCommand) Run(ctx *cmd.Context) error {
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

	termsId, err := charm.ParseTerm(c.TermID)
	if err != nil {
		return errors.Annotate(err, "invalid term format")
	}
	if termsId.Owner == "" {
		c.out.Write(ctx, "only terms with owners require releasing")
		return nil
	}
	if termsId.Revision == 0 {
		return errors.New("must specify a term revision")
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
