package scheduler

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
)

type CronOption func(option *cron.Option)

type Cron struct {
	c   *cron.Cron
	loc *time.Location
}

func NewCron(loc *time.Location) *Cron {
	if loc == nil {
		loc = time.Local
	}
	c := cron.New(cron.WithLocation(loc), cron.WithChain(cron.Recover(cron.DefaultLogger)))
	return &Cron{c: c, loc: loc}
}

func (cr *Cron) Start() { cr.c.Start() }
func (cr *Cron) Stop()  { ctx := cr.c.Stop(); <-ctx.Done() }

func (cr *Cron) Add(expr string, job Job) (cron.EntryID, error) {
	return cr.c.AddFunc(expr, func() { job.Run(context.Background()) })
}

func (cr *Cron) AddWithCtx(expr string, fn func(ctx context.Context)) (cron.EntryID, error) {
	return cr.c.AddFunc(expr, func() { fn(context.Background()) })
}

func (cr *Cron) Entries() []cron.Entry { return cr.c.Entries() }
