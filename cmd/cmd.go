// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/go-macaroon-bakery/macaroon-bakery/v3/httpbakery"
	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/gnuflag"
	"github.com/juju/idmclient/v2/ussologin"
	"github.com/juju/juju/juju/osenv"
	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/juju/utils/v2"
	"gopkg.in/juju/environschema.v1/form"

	"github.com/juju/terms-client/v2/api"
)

var (
	readFile = ioutil.ReadFile
)

type baseCommand struct {
	cmd.CommandBase

	ServiceURL string

	// NoBrowser specifies that web-browser-based auth should
	// not be used when authenticating.
	NoBrowser bool
}

// NewClient returns a new http bakery client for terms commands
// and a cleanup method that updates the stored cookie.
func (s *baseCommand) NewClient(ctx *cmd.Context) (*httpbakery.Client, func(), error) {
	jujuXDGDataHome := osenv.JujuXDGDataHomeDir()
	if jujuXDGDataHome == "" {
		return nil, func() {}, errors.Errorf("cannot determine juju data home, required environment variables are not set")
	}
	osenv.SetJujuXDGDataHome(jujuXDGDataHome)
	client := httpbakery.NewClient()
	if s.NoBrowser {
		filler := &form.IOFiller{
			In:  os.Stdin,
			Out: os.Stdout,
		}
		store := ussologin.NewFileTokenStore(osenv.JujuXDGDataHomePath("store-usso-token"))
		interactor := ussologin.NewInteractor(ussologin.StoreTokenGetter{
			Store: store,
			TokenGetter: ussologin.FormTokenGetter{
				Filler: filler,
				Name:   "juju",
			},
		})
		client.AddInteractor(interactor)
	} else {
		client.AddInteractor(httpbakery.WebBrowserInteractor{})
	}
	if jar, err := cookiejar.New(&cookiejar.Options{
		Filename: cookieFile(),
	}); err == nil {
		client.Jar = jar
		return client, func() {
			err := jar.Save()
			if err != nil {
				ctx.Warningf("failed to save cookie jar: %v", err)
			}
		}, nil
	} else {
		ctx.Warningf("failed to create cookie jar")
		return client, func() {}, nil
	}
}

// newBaseCommand creates a new baseCommand with the default service
// url set.
func newBaseCommand() *baseCommand {
	return &baseCommand{
		ServiceURL: api.BaseURL(),
	}
}

// SetFlags implements the Command interface.
func (c *baseCommand) SetFlags(f *gnuflag.FlagSet) {
	f.BoolVar(&c.NoBrowser, "B", false, "Do not use web browser for authentication")
	f.BoolVar(&c.NoBrowser, "no-browser-login", false, "")
	f.StringVar(&c.ServiceURL, "url", api.BaseURL(), "host and port of the terms service")
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
