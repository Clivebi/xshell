package xshell

import (
	"errors"
	"strings"
)

type Host struct {
	Host     string
	User     string
	Password string
}

func (o Host) String() string {
	return o.User + "@" + o.Host + "@" + o.Password
}

func NewHost(text string) (*Host, error) {
	host := &Host{}
	n := strings.Index(text, "@")
	if n+2 > len(text) {
		return nil, errors.New("invliad host " + text)
	}
	host.User = text[:n]
	text = text[n+1:]
	n = strings.Index(text, "@")
	if n == -1 {
		host.Host = text
		if len(host.Host) == 0 || len(host.User) == 0 {
			return nil, errors.New("invalid host " + text)
		}
		if !strings.Contains(host.Host, ":") {
			host.Host += ":22"
		}
		return host, nil
	}
	if n+2 > len(text) {
		return nil, errors.New("invliad host " + text)
	}
	host.Host = text[:n]
	host.Password = text[n+1:]
	if len(host.Host) == 0 || len(host.User) == 0 {
		return nil, errors.New("invalid host " + text)
	}
	if !strings.Contains(host.Host, ":") {
		host.Host += ":22"
	}
	return host, nil
}
