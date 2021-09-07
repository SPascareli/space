package application

import "time"

type Launch struct {
	ID           string
	Name         string
	NET          time.Time
	ProviderName string
	Status       string
}
