// Copyright 2016 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"fmt"

	"github.com/juju/cmd"
	"launchpad.net/gnuflag"

	"github.com/juju/terms-client/api"
)

var (
	clientNew = func(options ...api.ClientOption) (api.Client, error) {
		return api.NewClient(options...)
	}
)

type commandWithDescription interface {
	cmd.Command
	Description() string
}

// WrapPlugin returns a wrapped plugin command.
func WrapPlugin(cmd commandWithDescription) cmd.Command {
	return &pluginWrapper{commandWithDescription: cmd}
}

type pluginWrapper struct {
	commandWithDescription
	Description bool
}

// SetFlags implements Command.SetFlags.
func (c *pluginWrapper) SetFlags(f *gnuflag.FlagSet) {
	c.commandWithDescription.SetFlags(f)
	f.BoolVar(&c.Description, "description", false, "returns command description")
}

// Init implements Command.Init.
func (c *pluginWrapper) Init(args []string) error {
	if c.Description {
		return nil
	}
	return c.commandWithDescription.Init(args)
}

// Run implements Command.Run.
func (c *pluginWrapper) Run(ctx *cmd.Context) error {
	if c.Description {
		fmt.Fprint(ctx.Stdout, c.commandWithDescription.Description())
		return nil
	}
	return c.commandWithDescription.Run(ctx)
}
