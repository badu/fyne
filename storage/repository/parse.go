package repository

import (
	"errors"
	"path/filepath"
	"runtime"
	"strings"

	uriParser "github.com/fredbi/uri"

	"fyne.io/fyne/v2"
)

// NewFileURI implements the back-end logic to storage.NewFileURI, which you
// should use instead. This is only here because other functions in repository
// need to call it, and it prevents a circular import.
//
// Since: 2.0
func NewFileURI(path string) *TheURI {
	// URIs are supposed to use forward slashes. On Windows, it
	// should be OK to use the platform native filepath with UNIX
	// or NT style paths, with / or \, but when we reconstruct
	// the URI, we want to have / only.
	if runtime.GOOS == "windows" {
		// seems that sometimes we end up with
		// double-backslashes
		path = filepath.ToSlash(path)
	}

	return &TheURI{
		scheme: "file",
		path:   path,
	}
}

// ParseURI implements the back-end logic for storage.ParseURI, which you
// should use instead. This is only here because other functions in repository
// need to call it, and it prevents a circular import.
//
// Since: 2.0
func ParseURI(s string) (fyne.URI, error) {
	// Extract the scheme.
	scheme, _, ok := strings.Cut(s, ":")
	if !ok {
		return nil, errors.New("invalid URI, scheme must be present")
	}

	if strings.EqualFold(scheme, "file") {
		// Does this really deserve to be special? In principle, the
		// purpose of this check is to pass it to NewFileURI, which
		// allows platform path seps in the URI (against the RFC, but
		// easier for people building URIs naively on Windows). Maybe
		// we should punt this to whoever generated the URI in the
		// first place?

		if len(s) <= 7 {
			return nil, errors.New("not a valid URI")
		}
		path := s[5:] // everything after file:
		if len(path) > 2 && path[:2] == "//" {
			path = path[2:]
		}

		// Windows files can break authority checks, so just return the parsed file URI
		return NewFileURI(path), nil
	}

	scheme = strings.ToLower(scheme)
	repo, err := ForScheme(scheme)
	if err == nil {
		// If the repository registered for this scheme implements a parser
		if c, ok := repo.(CustomURIRepository); ok {
			return c.ParseURI(s)
		}
	}

	// There was no repository registered, or it did not provide a parser

	l, err := uriParser.Parse(s)
	if err != nil {
		return nil, err
	}

	authority := ""

	if userInfo := l.Authority().UserInfo(); len(userInfo) > 0 {
		authority += userInfo + "@"
	}

	authority += l.Authority().Host()

	if port := l.Authority().Port(); len(port) > 0 {
		authority += ":" + port
	}

	return &TheURI{
		scheme:    scheme,
		authority: authority,
		path:      l.Authority().Path(),
		query:     l.Query().Encode(),
		fragment:  l.Fragment(),
	}, nil
}
