package rungroup

import (
	"context"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type Group struct {
	jobs []func(context.Context) error
}

func (g *Group) AddJob(job func(context.Context) error) {
	g.jobs = append(g.jobs, job)
}

func (g *Group) AddJobs(jobs ...func(context.Context) error) {
	g.jobs = append(g.jobs, jobs...)
}

func (g *Group) RunAndWait(ctx context.Context) error {
	var retErr error

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Collecting errors from all jobs until all of them are down
	errChan := make(chan error)
	wgErr := sync.WaitGroup{}
	wgErr.Add(1)
	go func() {
		defer wgErr.Done()

		for err := range errChan {
			if err != nil {
				retErr = multierror.Append(retErr, err)
			}
		}
	}()

	// Running jobs
	wgRun := sync.WaitGroup{}
	wgRun.Add(len(g.jobs))
	for _, job := range g.jobs {
		job := job

		go func() {
			defer wgRun.Done()

			if err := job(runCtx); err != nil {
				errChan <- err
			}
			cancel()
		}()
	}

	// Waiting until all jobs are finished
	wgRun.Wait()

	// Everyone who could send write to channel are down, we cal safely close it
	close(errChan)

	// Terminating error-collecting routine
	wgErr.Wait()

	// Voilà!
	return retErr //nolint:wrapcheck // nothing to add
}
