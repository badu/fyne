package repository

import (
	"bufio"
	"mime"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type TheURI struct {
	scheme    string
	authority string
	path      string
	query     string
	fragment  string
}

func (u *TheURI) Extension() string {
	return filepath.Ext(u.path)
}

func (u *TheURI) Name() string {
	return filepath.Base(u.path)
}

func (u *TheURI) MimeType() string {
	mimeTypeFull := mime.TypeByExtension(u.Extension())
	if mimeTypeFull == "" {
		mimeTypeFull = "text/plain"

		repo, err := ForURI(u)
		if err != nil {
			return "application/octet-stream"
		}

		readCloser, err := repo.Reader(u)
		if err == nil {
			defer readCloser.Close()
			scanner := bufio.NewScanner(readCloser)
			if scanner.Scan() && !utf8.Valid(scanner.Bytes()) {
				mimeTypeFull = "application/octet-stream"
			}
		}
	}

	mimeType, _, _ := strings.Cut(mimeTypeFull, ";")
	return mimeType
}

func (u *TheURI) Scheme() string {
	return u.scheme
}

func (u *TheURI) String() string {
	// NOTE: this string reconstruction is mandated by IETF RFC3986,
	// section 5.3, pp. 35.

	s := u.scheme + "://" + u.authority + u.path
	if len(u.query) > 0 {
		s += "?" + u.query
	}
	if len(u.fragment) > 0 {
		s += "#" + u.fragment
	}
	return s
}

func (u *TheURI) Authority() string {
	return u.authority
}

func (u *TheURI) Path() string {
	return u.path
}

func (u *TheURI) Query() string {
	return u.query
}

func (u *TheURI) Fragment() string {
	return u.fragment
}
