// Copyright 2016 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

var (
	ClientNew    = &clientNew
	ReadFile     = &readFile
	NewIDMClient = &newIDMClient
)

// BaseCommand type is exported for test purposes.
type BaseCommand struct {
	*baseCommand
}

// NewBaseCommand returns a new instance
// of the BaseCommand.
func NewBaseCommand() BaseCommand {
	return BaseCommand{&baseCommand{}}
}
