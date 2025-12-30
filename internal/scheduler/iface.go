package scheduler

import "context"

type IScheduler interface {
	CreateOrUpdateCronJobs(ctx context.Context, jobs []CronJob) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
