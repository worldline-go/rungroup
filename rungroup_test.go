package rungroup

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type TestService struct {
	name          string
	closeDuration time.Duration
	chErrorTerm   chan struct{}
}

func NewTestService(name string, closeDuration time.Duration) *TestService {
	return &TestService{
		name:          name,
		closeDuration: closeDuration,
		chErrorTerm:   make(chan struct{}),
	}
}

func (svc *TestService) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		fmt.Printf("[%s] finishing Run\n", svc.name)
		time.Sleep(svc.closeDuration)

		return nil

	case <-svc.chErrorTerm:
		return fmt.Errorf("[%s] something went wrong", svc.name)
	}
}

func (svc *TestService) Close() error {
	close(svc.chErrorTerm)

	fmt.Printf("[%s] is closed\n", svc.name)

	return nil
}

func (svc *TestService) RaiseError() {
	fmt.Printf("[%s] got raise an error\n", svc.name)
	svc.chErrorTerm <- struct{}{}
}

func TestGroup_RunAndWait(t *testing.T) {
	t.Parallel()

	svc1 := NewTestService("svc1", 2*time.Second)
	svc2 := NewTestService("svc2", 0*time.Second)

	runGroup := Group{}

	runGroup.AddJob(func(ctx context.Context) error {
		defer func() { _ = svc1.Close() }()

		return svc1.Run(ctx)
	})

	runGroup.AddJob(func(ctx context.Context) error {
		defer func() { _ = svc2.Close() }()

		return svc2.Run(ctx)
	})

	go func() {
		time.Sleep(3 * time.Second)

		svc2.RaiseError()
	}()

	ctx := context.Background()
	if err := runGroup.RunAndWait(ctx); err != nil {
		fmt.Println("error:", err)
	}

}
