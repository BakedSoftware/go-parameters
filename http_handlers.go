package parameters

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// CORSHeaders adds cross origin resource sharing headers to a response
func CORSHeaders(fn httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		fn(w, r, p)
	}
}

// SendCORS sends a cross origin resource sharing header only
func SendCORS(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	if origin := req.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(http.StatusOK)
}

// JSONResp will set the content-type to application/json
func JSONResp(fn httprouter.Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
		rw.Header().Set("Content-Type", "application/json")
		fn(rw, req, p)
	}
}

// GeneralResponse calls the default wrappers: EnableGZIP, LogRequest,
// CORSHeaders
func GeneralResponse(fn httprouter.Handle) httprouter.Handle {
	return EnableGZIP(MakeHTTPRouterParsedReq(CORSHeaders(fn)))
}

// GeneralJSONRequest calls the default wrappers for a json response:
// EnableGZIP, JSONResp, LogRequest, CORSHeaders
func GeneralJSONResponse(fn httprouter.Handle) httprouter.Handle {
	return EnableGZIP(JSONResp(MakeHTTPRouterParsedReq(CORSHeaders(fn))))
}

var filterReplace = [...]string{"FILTERED"}

// FilteredKeys is a lower case array of keys to filter when logging
var FilteredKeys []string

// filterMap will filter the parameters and not log parameters with sensitive
// data. To add more parameters - see the if in the loop
func filterMap(params *Params) *Params {
	var filtered Params
	filtered.Values = make(map[string]interface{}, len(params.Values))

	for k, v := range params.Values {
		if contains(FilteredKeys, k) {
			filtered.Values[k] = filterReplace[:]
		} else if b, ok := v.([]byte); ok {
			filtered.Values[k] = string(b)
		} else {
			filtered.Values[k] = v
		}
	}
	return &filtered
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	if "" == w.Header().Get("Content-Type") {
		// If no content type, apply sniffing algorithm to un-gzipped body.
		w.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return w.Writer.Write(b)
}

// EnableGZIP will attempt to compress the response if the client has passed a
// header value for Accept-Encoding which allows gzip
func EnableGZIP(fn httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn(w, r, p)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		fn(gzr, r, p)
	}
}
