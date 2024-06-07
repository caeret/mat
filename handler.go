package mat

import (
	"context"
	"github.com/caeret/mat/decoder"
	"net/http"
	"reflect"
	"runtime"
	"sync"
)

var handleMap sync.Map

// StatusCoder allows you to customise the HTTP response code.
type StatusCoder interface {
	StatusCode() int
}

// Headerer allows you to customise the HTTP headers.
type Headerer interface {
	Header() http.Header
}

type ContextAware interface {
	WithContext(c *Context) error
}

// Handle is the type for your handlers.
type Handle[T, O any] func(ctx context.Context, request T) (O, error)

// H wraps your handler function with the Go generics magic.
func H[T, O any](handle Handle[T, O]) Handler {
	pool := decoder.NewRequestPool(*new(T))
	decodeRequest := newRequestDecoder(*new(T))

	h := func(c *Context) error {
		var res any

		req := pool.Get()
		err := decodeRequest(req, c)
		if err != nil {
			return err
		} else {
			res, err = handle(c.Context(), *req)
			if err != nil {
				return err
			}
		}
		pool.Put(req)

		if h, ok := res.(Headerer); ok {
			for k, v := range h.Header() {
				c.Response.Header().Add(k, v[0])
			}
		}

		if sc, ok := res.(StatusCoder); ok {
			c.Response.WriteHeader(sc.StatusCode())
		}

		return c.Write(res)
	}

	handleMap.Store(reflect.ValueOf(h).Pointer(), runtime.FuncForPC(reflect.ValueOf(handle).Pointer()).Name())
	return h
}

type requestDecoder[V any] func(v *V, c *Context) error

func newRequestDecoder[V any](v V) requestDecoder[V] {
	path, _ := decoder.NewCached(v, "path")
	query, _ := decoder.NewCached(v, "query")
	header, _ := decoder.NewCached(v, "header")
	param, _ := decoder.NewCached(v, "param")

	if path == nil && query == nil && header == nil && param == nil {
		return decodeBody[V](true)
	}

	return decodeRequest(path, query, header, param)
}

func decodeRequest[V any](path, query, header, param *decoder.CachedDecoder[V]) requestDecoder[V] {
	body := decodeBody[V](false)
	return func(v *V, c *Context) error {
		if err := body(v, c); err != nil {
			return err
		}

		val := reflect.ValueOf(v).Elem()

		if p := c.Params(); path != nil && len(p) > 0 {
			if err := path.DecodeValue((decoder.Params)(p), val); err != nil {
				return err
			}
		}

		if query != nil {
			if q := c.Request.URL.Query(); len(q) > 0 {
				if err := query.DecodeValue((decoder.Args)(q), val); err != nil {
					return err
				}
			}
		}

		if header != nil {
			if err := header.DecodeValue((decoder.Header)(c.Request.Header), val); err != nil {
				return err
			}
		}

		if param != nil {
			p := decoder.Params{"ip": c.RealIP()}
			if err := param.DecodeValue(p, val); err != nil {
				return err
			}
		}

		if ca, ok := any(v).(ContextAware); ok {
			err := ca.WithContext(c)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func decodeBody[V any](enable bool) requestDecoder[V] {
	return func(v *V, c *Context) error {
		if enable {
			if ca, ok := any(v).(ContextAware); ok {
				err := ca.WithContext(c)
				if err != nil {
					return err
				}
			}
		}

		if c.Request.ContentLength == 0 || c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
			return nil
		}

		return c.Read(&v)
	}
}
