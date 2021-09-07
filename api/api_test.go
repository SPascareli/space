package api

import (
	"context"
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/spascareli/space/application"
	"github.com/stretchr/testify/assert"
)

type errMissing struct{}

func (e errMissing) Error() string { return "error" }
func (e errMissing) Missing() bool { return true }

type errTemporary struct{}

func (e errTemporary) Error() string   { return "error" }
func (e errTemporary) Temporary() bool { return true }

type errUnauthorized struct{}

func (e errUnauthorized) Error() string      { return "error" }
func (e errUnauthorized) Unauthorized() bool { return true }

type errInvalid string

func (e errInvalid) Error() string { return string(e) }
func (e errInvalid) Invalid() bool { return true }

func TestTodayError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code int
	}{
		{
			name: "should 404 when is missing error",
			err:  &errMissing{},
			code: 404,
		},
		{
			name: "should 401 when is unauthorized error",
			err:  &errUnauthorized{},
			code: 401,
		},
		{
			name: "should 503 when is temporary error",
			err:  &errTemporary{},
			code: 503,
		},
		{
			name: "should 500 when is unknown error",
			err:  errors.New("what?"),
			code: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := &Adapter{
				launch: &mockLaunchService{
					LaunchesTodayError: test.err,
				},
			}
			req := httptest.NewRequest("GET", "/today", nil)

			w := httptest.NewRecorder()
			a.Today(w, req)

			assert.Equal(t, test.code, w.Result().StatusCode)
		})
	}
}

func TestTodaySuccess(t *testing.T) {
	a := &Adapter{
		launch: &mockLaunchService{},
	}
	req := httptest.NewRequest("GET", "/today", nil)

	w := httptest.NewRecorder()
	a.Today(w, req)

	assert.Equal(t, 200, w.Result().StatusCode)
}

func TestFailed(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		start string
		end   string
		code  int
	}{
		{
			name: "should fail with 400 when start is missing",
			end:  "pipoca",
			code: 400,
		},
		{
			name:  "should fail with 400 when end is missing",
			start: "pipoca",
			code:  400,
		},
		{
			name:  "should fail with 400 when start is not a valid date",
			start: "pipoca",
			end:   "pipoca",
			code:  400,
		},
		{
			name:  "should fail with 400 when end is not a valid date",
			start: "2021-09-03T00:00:00Z",
			end:   "pipoca",
			code:  400,
		},
		{
			name:  "should 503 when is temporary error",
			start: "2021-09-03T00:00:00Z",
			end:   "2021-09-03T00:00:00Z",
			err:   &errTemporary{},
			code:  503,
		},
		{
			name:  "should 500 when is unknown error",
			start: "2021-09-03T00:00:00Z",
			end:   "2021-09-03T00:00:00Z",
			err:   errors.New("what?"),
			code:  500,
		},
		{
			name:  "should 400 when is invalid error",
			start: "2021-09-03T00:00:00Z",
			end:   "2021-09-03T00:00:00Z",
			err:   errInvalid("err"),
			code:  400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := &Adapter{
				launch: &mockLaunchService{
					GetFailedError: test.err,
				},
			}
			q := url.Values{
				"start": []string{test.start},
				"end":   []string{test.end},
			}
			uri := url.URL{
				Path:     "/failed",
				RawQuery: q.Encode(),
			}
			req := httptest.NewRequest("GET", uri.String(), nil)

			w := httptest.NewRecorder()
			a.Failed(w, req)

			assert.Equal(t, test.code, w.Result().StatusCode)
		})
	}
}

func TestFailedSuccess(t *testing.T) {
	a := &Adapter{
		launch: &mockLaunchService{},
	}
	q := url.Values{
		"start": []string{"2021-09-03T00:00:00Z"},
		"end":   []string{"2021-09-03T00:00:00Z"},
	}
	uri := url.URL{
		Path:     "/failed",
		RawQuery: q.Encode(),
	}
	req := httptest.NewRequest("GET", uri.String(), nil)

	w := httptest.NewRecorder()
	a.Failed(w, req)

	assert.Equal(t, 200, w.Result().StatusCode)
}

type mockLaunchService struct {
	LaunchesTodayResponse []application.Launch
	LaunchesTodayError    error
	GetFailedResponse     []application.Launch
	GetFailedError        error
}

func (m *mockLaunchService) LaunchesToday(_ context.Context) ([]application.Launch, error) {
	return m.LaunchesTodayResponse, m.LaunchesTodayError
}

func (m *mockLaunchService) GetFailed(_ context.Context, _ time.Time, _ time.Time) ([]application.Launch, error) {
	return m.GetFailedResponse, m.GetFailedError
}
