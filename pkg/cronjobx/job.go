package cronjobx

import "context"

type Job interface {
	Name() string
	Run(ctx context.Context) error
}
