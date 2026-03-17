package actor

import (
	"context"
	"sync"
)

// Coordinator manages the lifecycle of multiple actors.
// It provides a central point for starting and stopping all actors
// in a coordinated fashion.
//
// Example:
//
//	coord := actor.NewCoordinator()
//	coord.Register("cache", cacheActor)
//	coord.Register("tracker", trackerActor)
//	coord.Start(context.Background())
//	defer coord.Stop()
type Coordinator struct {
	mu     sync.Mutex
	actors map[string]Actor
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewCoordinator creates a new coordinator with no actors registered.
func NewCoordinator() *Coordinator {
	return &Coordinator{
		actors: make(map[string]Actor),
	}
}

// Register adds an actor to the coordinator with a unique name.
// Actors must be registered before Start is called.
// If an actor with the same name exists, it is replaced.
func (c *Coordinator) Register(name string, actor Actor) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.actors[name] = actor
}

// Unregister removes an actor from the coordinator.
// This has no effect if the coordinator is already running.
func (c *Coordinator) Unregister(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.actors, name)
}

// Start launches all registered actors in separate goroutines.
// Each actor runs until the context is cancelled.
// Start can only be called once; subsequent calls are no-ops.
func (c *Coordinator) Start(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Prevent double start
	if c.cancel != nil {
		return
	}

	ctx, c.cancel = context.WithCancel(ctx)
	for name, actor := range c.actors {
		c.wg.Add(1)
		go func(n string, a Actor) {
			defer c.wg.Done()
			a.Run(ctx)
		}(name, actor)
	}
}

// Stop gracefully shuts down all actors.
// It cancels the context and waits for all actors to finish.
// Stop is idempotent and can be called multiple times.
func (c *Coordinator) Stop() {
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
	c.mu.Unlock()
	c.wg.Wait()
}

// ActorNames returns the names of all registered actors.
// This is useful for debugging and monitoring.
func (c *Coordinator) ActorNames() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	names := make([]string, 0, len(c.actors))
	for name := range c.actors {
		names = append(names, name)
	}
	return names
}

// ActorCount returns the number of registered actors.
func (c *Coordinator) ActorCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.actors)
}
