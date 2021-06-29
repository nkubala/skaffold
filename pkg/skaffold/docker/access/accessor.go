package access

import (
	"context"
	"io"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker/tracker"
)

type Accessor struct{}

func NewAccessor(t *tracker.ContainerTracker) *Accessor {
	return &Accessor{}
}

// Start starts the resource accessor.
func (a *Accessor) Start(context.Context, io.Writer, []string) error { return nil }

// Stop stops the resource accessor.
func (a *Accessor) Stop() {}
