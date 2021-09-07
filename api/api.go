package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/spascareli/space/application"
)

type LaunchService interface {
	Today(context.Context) ([]application.Launch, error)
	GetFailed(ctx context.Context, start time.Time, end time.Time) ([]application.Launch, error)
}

type Adapter struct {
	launch LaunchService
}

func NewAdapter(launch LaunchService) *Adapter {
	return &Adapter{launch}
}

func (a *Adapter) Run(ctx context.Context) error {
	http.HandleFunc("/today", a.Today)
	http.HandleFunc("/failed", a.Failed)

	return http.ListenAndServe(":8082", nil)
}

func (a *Adapter) Today(w http.ResponseWriter, r *http.Request) {
	data, err := a.launch.Today(r.Context())
	if err != nil {
		handlerErrorHeader(err, w)
		fmt.Fprint(w, err)
		return
	}
	resp, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, string(resp))
}

func (a *Adapter) Failed(w http.ResponseWriter, r *http.Request) {
	var startParam, endParam string
	v := r.URL.Query()
	if startParam = v.Get("start"); startParam == "" {
		w.WriteHeader(400)
		fmt.Fprintf(w, "missing required parameter: %s", "start")
		return
	}
	if endParam = v.Get("end"); endParam == "" {
		w.WriteHeader(400)
		fmt.Fprintf(w, "missing required parameter: %s", "end")
		return
	}
	start, err := time.Parse(time.RFC3339, startParam)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "badly formated parameter: %s", "start")
		return
	}
	end, err := time.Parse(time.RFC3339, endParam)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "badly formated parameter: %s", "end")
		return
	}
	data, err := a.launch.GetFailed(r.Context(), start, end)
	if err != nil {
		handlerErrorHeader(err, w)
		fmt.Fprint(w, err)
		return
	}
	resp, _ := json.Marshal(data)
	fmt.Fprint(w, string(resp))
}

func handlerErrorHeader(err error, w http.ResponseWriter) {
	switch {
	case isMissing(err):
		w.WriteHeader(404)
	case isTemporary(err):
		w.WriteHeader(503)
	case isUnauthorized(err):
		w.WriteHeader(401)
	case isInvalid(err):
		w.WriteHeader(400)
	case err != nil:
		w.WriteHeader(500)
	}
}

func isMissing(err error) bool {
	e, ok := errors.Cause(err).(interface {
		Missing() bool
	})
	return ok && e.Missing()
}

func isTemporary(err error) bool {
	e, ok := errors.Cause(err).(interface {
		Temporary() bool
	})
	return ok && e.Temporary()
}

func isUnauthorized(err error) bool {
	e, ok := errors.Cause(err).(interface {
		Unauthorized() bool
	})
	return ok && e.Unauthorized()
}

func isInvalid(err error) bool {
	e, ok := errors.Cause(err).(interface {
		Invalid() bool
	})
	return ok && e.Invalid()
}
