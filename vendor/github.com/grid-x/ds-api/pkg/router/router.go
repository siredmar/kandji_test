package router

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

// acceptRe retrieves the API version from custom mediatype.
//
// The Accept-header should be like the following:
//	* application/vnd.gridx.ai.2018-12-07
//	* application/vnd.gridx.ai.2018-12-07+json
var acceptRe = regexp.MustCompile(`application\/vnd\.gridx.ai(\.20[0-9]{2,}-[0-1][0-9]-[0-3][0-9])?(?:\+json)?`)

func extractVersion(r *http.Request) (string, error) {
	var version string
	// http.Header's keys are in canonical format which means the first letter
	// and any letter following a hypen is upper case; the rest are lower case.
	accepts := r.Header["Accept"]
	if len(accepts) == 0 {
		return "", fmt.Errorf("no version selected")
	}
	for _, accept := range accepts {
		ver := acceptRe.FindStringSubmatch(accept)
		if len(ver) == 2 && ver[1] != "" {
			version = ver[1][1:] // First char is a dot.
		}
	}
	if version == "" {
		return "", fmt.Errorf("no version found")
	}
	return version, nil
}

// Middleware represents a middleware (what else?)
type Middleware func(http.Handler) http.Handler

// Populate adds the version routes to the mux router
func Populate(routes *Routes, router *mux.Router) {
	endpoints := make(VersionedEndpoints)
	for _, version := range routes.Endpoints {
		for _, action := range version {
			endpoint, ok := endpoints.Get(action.Action)
			if !ok {
				endpoint = &VersionedEndpoint{
					indices: make(map[string]int),
				}
				router.Path(action.Path).Name(action.Action).Methods(action.Methods...).Handler(endpoint)
				endpoints.Add(action.Action, endpoint)
			}
			f := action.Func
			// handler specific middlewares
			// the last one in the list will be applied first
			for _, m := range action.Middlewares {
				f = m(http.HandlerFunc(f)).ServeHTTP
			}
			// routes global middlewares. The last one in the list
			// will in the end be applied first
			for _, m := range routes.Middlewares {
				f = m(http.HandlerFunc(f)).ServeHTTP
			}
			endpoints.Set(action.Action, action.Version, f)
		}
	}
}

// Endpoint maps to a specific handler for a specific version
type Endpoint struct {
	Version     string
	Action      string
	Resource    string
	Path        string
	Middlewares []Middleware
	Methods     []string
	Func        func(w http.ResponseWriter, r *http.Request)
}

// Routes describes a set of endpoints for a set of versions
type Routes struct {
	// Middlewares are globally applied middlewares for all routes, e.g.
	// logging
	Middlewares []Middleware
	// Endpoints contains a set of API versions sorted by the publish date,
	// with the most recent versions appearing first. Each version only
	// contains the modified endpoint identified by their action.
	Endpoints []map[string]*Endpoint
}

// VersionedEndpoints is a map of versions to a versioned endpoint
type VersionedEndpoints map[string]*VersionedEndpoint

// Get returns the versioned endpoint with the given key
func (ee VersionedEndpoints) Get(key string) (*VersionedEndpoint, bool) {
	e, ok := ee[key]
	return e, ok
}

// Add adds the new versioned endpoint for the given key
func (ee VersionedEndpoints) Add(key string, e *VersionedEndpoint) {
	ee[key] = e
}

// Set adds a given version of a given handler for a given key
func (ee VersionedEndpoints) Set(key, version string, fn func(w http.ResponseWriter, r *http.Request)) {
	// Set version index
	ee[key].indices[version] = len(ee[key].versions)
	// Add version
	ee[key].versions = append(ee[key].versions, fn)
}

// VersionedEndpoint handles the version of own endpoints.
type VersionedEndpoint struct {
	// Versions contains a a set of endpoint versions sorted by
	// publish date, with the most recent versions appearing first.
	versions []http.HandlerFunc

	// indices contains the version index to lookup the nearest version.
	indices map[string]int
}

// ServeHTTP servers the versioned endpoint by picking the right handler to serve
func (e *VersionedEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(e.versions) == 0 {
		http.NotFound(w, r)
		return
	}

	version, err := extractVersion(r)
	if err != nil {
		http.Error(w, "version error", http.StatusBadRequest)
		return
	}

	e.pickHandler(version)(w, r)
}

func (e *VersionedEndpoint) pickHandler(requestVersion string) func(http.ResponseWriter, *http.Request) {
	// Speed improvement: Usage of prefix trie might be a better version router.
	return func(w http.ResponseWriter, r *http.Request) {
		for version, i := range e.indices {
			if version <= requestVersion {
				e.versions[i](w, r)
				return
			}
		}
		e.versions[0](w, r)
	}
}
