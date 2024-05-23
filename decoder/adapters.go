package decoder

import (
	"net/http"
	"net/url"
)

type Args url.Values

func (m Args) Get(key string) string {
	if m == nil {
		return ""
	}
	if vs := m[key]; len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func (m Args) Values(key string) []string {
	if m == nil {
		return nil
	}
	return m[key]
}

type Header http.Header

func (m Header) Get(key string) string {
	if m == nil {
		return ""
	}
	if vs := m[key]; len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func (m Header) Values(key string) []string {
	if m == nil {
		return nil
	}
	return m[key]
}

type Params map[string]string

func (m Params) Get(key string) string {
	return m[key]
}

func (m Params) Values(key string) []string {
	v, ok := m[key]
	if !ok {
		return nil
	}
	return []string{v}
}
