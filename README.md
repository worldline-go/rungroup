# rungroup

Some services run its parts in parallel (i.e. http server, Kafka consumer). 
Because of this:
- there's no one blocking call in main func
- if one of parallel run parts is down the rest should correctly close before the service goes down

`rungroup` manages correct lifecycle for such cases.

## How to use it?

You can find a working example [here](./rungroup_test.go#L50-L81).

`rungroup` runs jobs in a separate go routine each. 
When one of the jobs returns, `rungroup` cancels a context and waits until all jobs are finished.
It collects all errors and returns it in one piece.

**!!! Don't forget to close correctly every job.**
It might be done in jobs' defer section or after `RunAndWait()` call.

```go
func letsRun(ctx context.Context) error{
    runGroup := Group{}

    svc1 := NewTestService("svc1", 2*time.Second)
    svc2 := NewTestService("svc2", 0*time.Second)
    
    runGroup.AddJob(func(ctx context.Context) error {
        defer func() { _ = svc1.Close() }()
    
        return svc1.Run(ctx)
    })
    
    runGroup.AddJob(func(ctx context.Context) error {
        defer func() { _ = svc2.Close() }()
    
        return svc2.Run(ctx)
    })
    
    return runGroup.RunAndWait(ctx)
}
```
