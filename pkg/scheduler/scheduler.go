package scheduler

import (
	"context"
	"time"
)

type Job interface{ Run(ctx context.Context) }

type FuncJob func(ctx context.Context)

func (f FuncJob) Run(ctx context.Context) { f(ctx) }

type Scheduler struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func New() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{ctx: ctx, cancel: cancel}
}

func (s *Scheduler) Stop() { s.cancel() }

func (s *Scheduler) Every(d time.Duration, job Job) { go s.loopEvery(d, job) }

func (s *Scheduler) DailyAt(hh, mm int, job Job) { go s.loopDaily(hh, mm, job) }

func (s *Scheduler) OnceAfter(d time.Duration, job Job) { go s.onceAfter(d, job) }

func (s *Scheduler) loopEvery(d time.Duration, job Job) {
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-t.C:
			job.Run(s.ctx)
		}
	}
}

func (s *Scheduler) loopDaily(hh, mm int, job Job) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), hh, mm, 0, 0, now.Location())
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(next.Sub(now)):
			job.Run(s.ctx)
		}
	}
}

func (s *Scheduler) onceAfter(d time.Duration, job Job) {
	select {
	case <-s.ctx.Done():
		return
	case <-time.After(d):
		job.Run(s.ctx)
	}
}
