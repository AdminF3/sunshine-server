package sentry

import (
	"fmt"
	"log"
	"net/http"

	raven "github.com/getsentry/raven-go"
)

// Extra data for an error intended to be passed to Sentry. Arbitrary unstructured
// data which is stored with an event sample.
type Extra = raven.Extra

// Tags are sentry tags for an error intended to be passed to Sentry. Key-value
// tags which generate breakdowns charts and search filters.
type Tags map[string]string

// Report sends an error to Sentry. If err is nil, this func is a noop.
//
// Optionally additional context could be passed. Allowed context values are
// string (error message), Tags, Extra and func(Extra) (modifying Extra). If
// any other type is provided it will be passed as extra data under
// "__untyped__" key. This func expects not more than one of Extra and Tags,
// otherwise any subsequent of these would override previous argument of its
// type. Several func(Extra) arguments are accumulating without overwriting.
//
// See: https://docs.sentry.io/learn/context/
func Report(err error, context ...interface{}) error {
	if err == nil {
		return nil
	}

	var (
		tags    Tags
		extra   Extra
		untyped []interface{}
	)

	for _, ctx := range context {
		switch c := ctx.(type) {
		case string:
			err = fmt.Errorf("%s: %w", c, err)
		case Tags:
			tags = c
		case Extra:
			extra = c
		case func(Extra):
			if extra == nil {
				extra = make(Extra)
			}
			c(extra)
		default:
			if untyped == nil {
				untyped = make([]interface{}, 0, 1)
			}
			untyped = append(untyped, c)
		}
	}

	if len(untyped) > 0 {
		if extra == nil {
			extra = make(Extra)
		}
		extra["__untyped__"] = untyped
	}

	log.Printf("ERROR: %s, tags: %v, extra: %#v", err, tags, extra)
	ex := &raven.Exception{
		Stacktrace: raven.NewStacktrace(1, 10, []string{"stageai.tech"}),
		Type:       err.Error(),
	}
	raven.Capture(raven.NewPacketWithExtra(err.Error(), extra, ex), tags)

	return err
}

// CaptureRequest inserts HTTP headers and URL to extra tags of a Sentry capture.
func CaptureRequest(r *http.Request) func(Extra) {
	return func(e Extra) {
		e["http_headers"] = r.Header
		e["http_method"] = r.Method
		e["url"] = r.URL.String()
		e["content_length"] = r.ContentLength
		e["remote_addr"] = r.RemoteAddr
	}
}
