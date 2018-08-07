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

const showTermDoc = `
show-term is used to show a specific Terms and Conditions document.
Examples
show-term enterprise-plan/1
   shows revision 1 of the enterprise-plan Terms and Conditions.
show-term enterprise-plan
   shows the latest revision of the enterprise plan Terms and Conditions.   
`
const showTermPurpose = "shows the specified term"

// NewShowTermCommand returns a new command that can be used
// to shows Terms and Conditions document.
func NewShowTermCommand() cmd.Command {
	return &showTermCommand{}
}

type showTermCommand struct {
	baseCommand
	out cmd.Output

	TermID      string
	ShowContent bool
}

// SetFlags implements Command.SetFlags.
func (c *showTermCommand) SetFlags(f *gnuflag.FlagSet) {
	c.out.AddFlags(f, "yaml", cmd.DefaultFormatters)
	f.BoolVar(&c.ShowContent, "content", false, "show term contents only")
	c.baseCommand.SetFlags(f)
}

// Info implements Command.Info.
func (c *showTermCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "show-term",
		Args:    "<term id>",
		Purpose: showTermPurpose,
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

// Description returns a one-line description of the command.
func (c *showTermCommand) Description() string {
	return showTermPurpose
}

// Run implements Command.Run.
func (c *showTermCommand) Run(ctx *cmd.Context) error {
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

	response, err := termsClient.GetTerm(termsId.Owner, termsId.Name, termsId.Revision)
	if err != nil {
		return errors.Trace(err)
	}

	if c.ShowContent {
		_, err = ctx.Stdout.Write([]byte(response.Content))
	} else {
		err = c.out.Write(ctx, response)
	}
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
