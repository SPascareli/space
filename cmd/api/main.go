package main

import (
	"context"

	"github.com/spascareli/space/api"
	"github.com/spascareli/space/application"
	"github.com/spascareli/space/pkg/retry"
)

func main() {
	repo := application.NewLaunchRepo("lldev.thespacedevs.com")
	repoWithLog := application.NewLaunchRepoWithLog(repo)
	repoWithRetry := application.NewLaunchRepoWithRetrier(repoWithLog, retry.Default())
	service := application.NewLaunchService(repoWithRetry)
	api := api.NewAdapter(service)
	panic(api.Run(context.Background()))
}
