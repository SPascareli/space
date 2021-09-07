package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/spascareli/space/pkg/retry"
)

type errNotFound string

func (e errNotFound) Error() string { return string(e) }
func (e errNotFound) Missing() bool { return true }

type launchDTO struct {
	ID                    string
	Name                  string
	Net                   time.Time
	LaunchServiceProvider struct {
		ID   int
		Name string
	} `json:"launch_service_provider"`
	Status struct {
		ID          int
		Name        string
		Description string
	}
}

type LaunchRepo struct {
	baseURL string
}

func NewLaunchRepo(baseURL string) *LaunchRepo {
	return &LaunchRepo{baseURL}
}

func (l *LaunchRepo) Get(ctx context.Context, params SearchParams) ([]Launch, error) {
	q := url.Values{
		"net__gte": []string{params.NETGreaterThanEqual.Format("2006-01-02T15:04:05Z07:00")},
		"net__lt":  []string{params.NETLessThanEqual.Format("2006-01-02T15:04:05Z07:00")},
	}
	status := parseStatus(params.Status)
	if status != 0 {
		q["status"] = []string{fmt.Sprintf("%d", status)}
	}
	uri := url.URL{
		Scheme:   "https",
		Host:     l.baseURL,
		Path:     "/2.2.0/launch",
		RawQuery: q.Encode(),
	}
	res, err := http.Get(uri.String())
	if err != nil {
		return nil, errors.Wrap(err, "http request failed")
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}
	switch res.StatusCode {
	case 404:
		return nil, errNotFound("no launches found")
	}

	var results struct {
		Count   int
		Results []launchDTO
	}
	err = json.Unmarshal(b, &results)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal launches result")
	}

	var data []Launch
	for _, launch := range results.Results {
		data = append(data, mapLaunchDTOToDomain(launch))
	}
	return data, nil
}

type getter interface {
	Get(context.Context, SearchParams) ([]Launch, error)
}

type LaunchRepoWithRetry struct {
	repo  getter
	retry *retry.Retrier
}

func NewLaunchRepoWithRetrier(repo getter, retrier *retry.Retrier) *LaunchRepoWithRetry {
	return &LaunchRepoWithRetry{repo, retrier}
}

func (l *LaunchRepoWithRetry) Get(ctx context.Context, params SearchParams) (launches []Launch, err error) {
	l.retry.IfTemporary(ctx, func() error {
		launches, err = l.repo.Get(ctx, params)
		return err
	})
	return
}

type LaunchRepoWithLog struct {
	repo getter
}

func NewLaunchRepoWithLog(repo getter) *LaunchRepoWithLog {
	return &LaunchRepoWithLog{repo}
}

func (l *LaunchRepoWithLog) Get(ctx context.Context, params SearchParams) (launches []Launch, err error) {
	launches, err = l.repo.Get(ctx, params)
	type logging struct {
		Message   string
		Timestamp time.Time
		Metadata  map[string]interface{}
	}
	log := logging{"called Get", time.Now(), map[string]interface{}{
		"params": params,
		"results": [2]interface{}{
			launches, err,
		},
	}}
	s, _ := json.Marshal(log)
	fmt.Println(string(s))
	return
}

func mapLaunchDTOToDomain(l launchDTO) Launch {
	return Launch{
		ID:           l.ID,
		Name:         l.Name,
		NET:          l.Net,
		ProviderName: l.LaunchServiceProvider.Name,
		Status:       l.Status.Description,
	}
}

func parseStatus(status LaunchStatus) int {
	switch status {
	case LaunchStatusSucceeded:
		return 3
	case LaunchStatusFailed:
		return 4
	default:
		return 0
	}
}
