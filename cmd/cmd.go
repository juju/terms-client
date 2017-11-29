// Copyright 2016 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"github.com/juju/terms-client/api"
)

var (
	clientNew = func(options ...api.ClientOption) (api.Client, error) {
		return api.NewClient(options...)
	}
)
