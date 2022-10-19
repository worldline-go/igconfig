package loader

import (
	"context"
	"fmt"
	"path"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/hashicorp/go-hclog"
	"github.com/rs/zerolog/log"
)

// DynamicValue allows to get dynamically updated values at a runtime.
//
// ---
//
// If specified key has new value - runner will be called.
//
// If specified key deleted or consul server closed nothing happen and you can restart consul server safetly.
// Remove specified key not trigger function!
// This function will spin up Goroutine to track changes in background.
//
// Developers will use a context. After that you can cancel that context to close listening and channel.
// Don't close channel manually.
// It return a channel to get new value as []byte.
//
// Example:
//
//	ch, err := loader.Consul{}.DynamicValue(ctx, "test/dynamic")
//	if err != nil {
//		log.Logger.Debug().Err(err).Msg("uupps")
//
//		return
//	}
//	for v := range ch {
//		// use v here
//	}
func (l Consul) DynamicValue(ctx context.Context, key string) (<-chan []byte, error) {
	if err := l.EnsureClient(); err != nil {
		return nil, err
	}

	if l.Plan == nil {
		plan, err := watch.Parse(map[string]interface{}{
			"type": "key",
			"key":  path.Join(ConsulConfigPathPrefix, key),
		})
		if err != nil {
			return nil, fmt.Errorf("wath.Parse %w", err)
		}

		// set plan to watch
		l.Plan = &Watch{
			Plan: plan,
		}
	}

	// not add any buffer, this is useful for getting latest change only
	vChannel := make(chan []byte)

	l.Plan.SetHandler(func(b []byte) {
		vChannel <- b
	})

	go func() {
		runCh := make(chan error, 1)

		go func() {
			runCh <- l.Plan.Run(l.Client)

			// close channel if plan stopped.
			close(vChannel)
		}()

		// this select-case for listen ctx done and plan run result same time
		select {
		case <-ctx.Done():
			l.Plan.Stop()
			log.Ctx(ctx).Debug().Msg("plan stopped")
		case err := <-runCh:
			log.Ctx(ctx).Error().Err(err).Msg("plan watching error")
		}
	}()

	return vChannel, nil
}

// Planer for dynamically get changes interface.
type Planer interface {
	SetHandler(func([]byte))
	Run(*api.Client) error
	Stop()
}

// Watch struct is adapter for consul watch api to Plan.
type Watch struct {
	Plan *watch.Plan
}

// SetHandler call function if new changes happen.
// If Plan value is nil, raise panic.
func (w *Watch) SetHandler(fn func([]byte)) {
	w.Plan.HybridHandler = func(_ watch.BlockingParamVal, raw interface{}) {
		if raw == nil {
			return
		}

		v, ok := raw.(*api.KVPair)
		if ok {
			fn(v.Value)

			return
		}
	}
}

// Run start to get changes.
// If Plan value is nil, raise panic.
func (w *Watch) Run(c *api.Client) error {
	if err := w.Plan.RunWithClientAndHclog(c, hclog.NewNullLogger()); err != nil {
		return fmt.Errorf("plan run; %w", err)
	}

	return nil
}

// Stop close run function.
// If Plan value is nil, raise panic.
func (w *Watch) Stop() {
	w.Plan.Stop()
}

var _ Planer = &Watch{}
