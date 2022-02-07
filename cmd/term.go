// Copyright 2022 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/juju/errors"
)

var validTermName = regexp.MustCompile(`^[a-z](-?[a-z0-9]+)+$`)

// TermsId represents a single term id. The term can either be owned
// or "public" (meaning there is no owner).
// The Revision starts at 1. Therefore a value of 0 means the revision
// is unset.
type TermsId struct {
	Tenant   string
	Owner    string
	Name     string
	Revision int
}

var (
	validUserNameSnippet = "[a-zA-Z0-9][a-zA-Z0-9.+-]*[a-zA-Z0-9]"
	validName            = regexp.MustCompile(fmt.Sprintf("^(?P<name>%s)(?:@(?P<domain>%s))?$", validUserNameSnippet, validUserNameSnippet))
)

// Validate returns an error if the Term contains invalid data.
func (t *TermsId) Validate() error {
	if t.Tenant != "" && t.Tenant != "cs" {
		if !validTermName.MatchString(t.Tenant) {
			return errors.Errorf("wrong term tenant format %q", t.Tenant)
		}
	}
	if t.Owner != "" && !validName.MatchString(t.Owner) {
		return errors.Errorf("wrong owner format %q", t.Owner)
	}
	if !validTermName.MatchString(t.Name) {
		return errors.Errorf("wrong term name format %q", t.Name)
	}
	if t.Revision < 0 {
		return errors.Errorf("negative term revision")
	}
	return nil
}

// String returns the term in canonical form.
// This would be one of:
//   tenant:owner/name/revision
//   tenant:name
//   owner/name/revision
//   owner/name
//   name/revision
//   name
func (t *TermsId) String() string {
	id := make([]byte, 0, len(t.Tenant)+1+len(t.Owner)+1+len(t.Name)+4)
	if t.Tenant != "" {
		id = append(id, t.Tenant...)
		id = append(id, ':')
	}
	if t.Owner != "" {
		id = append(id, t.Owner...)
		id = append(id, '/')
	}
	id = append(id, t.Name...)
	if t.Revision != 0 {
		id = append(id, '/')
		id = strconv.AppendInt(id, int64(t.Revision), 10)
	}
	return string(id)
}

// ParseTerm takes a termID as a string and parses it into a Term.
// A complete term is in the form:
// tenant:owner/name/revision
// This function accepts partially specified identifiers
// typically in one of the following forms:
// name
// owner/name
// owner/name/27 # Revision 27
// name/283 # Revision 283
// cs:owner/name # Tenant cs
func ParseTerm(s string) (*TermsId, error) {
	tenant := ""
	termid := s
	if t := strings.SplitN(s, ":", 2); len(t) == 2 {
		tenant = t[0]
		termid = t[1]
	}

	tokens := strings.Split(termid, "/")
	var term TermsId
	switch len(tokens) {
	case 1: // "name"
		term = TermsId{
			Tenant: tenant,
			Name:   tokens[0],
		}
	case 2: // owner/name or name/123
		termRevision, err := strconv.Atoi(tokens[1])
		if err != nil { // owner/name
			term = TermsId{
				Tenant: tenant,
				Owner:  tokens[0],
				Name:   tokens[1],
			}
		} else { // name/123
			term = TermsId{
				Tenant:   tenant,
				Name:     tokens[0],
				Revision: termRevision,
			}
		}
	case 3: // owner/name/123
		termRevision, err := strconv.Atoi(tokens[2])
		if err != nil {
			return nil, errors.Errorf("invalid revision number %q %v", tokens[2], err)
		}
		term = TermsId{
			Tenant:   tenant,
			Owner:    tokens[0],
			Name:     tokens[1],
			Revision: termRevision,
		}
	default:
		return nil, errors.Errorf("unknown term id format %q", s)
	}
	if err := term.Validate(); err != nil {
		return nil, errors.Trace(err)
	}
	return &term, nil
}
