package template

import (
	"bytes"
	"fmt"

	"github.com/docker/swarmkit/api"
	"github.com/docker/swarmkit/api/naming"
)

// Platform holds information about the underlying platform of the node
type Platform struct {
	Architecture string
	OS           string
}

// Context defines the strict set of values that can be injected into a
// template expression in SwarmKit data structure.
type Context struct {
	Service struct {
		ID     string
		Name   string
		Labels map[string]string
	}

	Node struct {
		ID       string
		Hostname string
		Platform Platform
	}

	Task struct {
		ID   string
		Name string
		Slot string

		// NOTE(stevvooe): Why no labels here? Tasks don't actually have labels
		// (from a user perspective). The labels are part of the container! If
		// one wants to use labels for templating, use service labels!
	}
}

// NewContext returns a new template context from the data available in the
// task and the node where it is scheduled to run.
// The provided context can then be used to populate runtime values in a
// ContainerSpec.
func NewContext(n *api.NodeDescription, t *api.Task) (ctx Context) {
	ctx.Service.ID = t.ServiceID
	ctx.Service.Name = t.ServiceAnnotations.Name
	ctx.Service.Labels = t.ServiceAnnotations.Labels

	ctx.Node.ID = t.NodeID

	// Add node information to context only if we have them available
	if n != nil {
		ctx.Node.Hostname = n.Hostname
		ctx.Node.Platform = Platform{
			Architecture: n.Platform.Architecture,
			OS:           n.Platform.OS,
		}
	}
	ctx.Task.ID = t.ID
	ctx.Task.Name = naming.Task(t)

	if t.Slot != 0 {
		ctx.Task.Slot = fmt.Sprint(t.Slot)
	} else {
		// fall back to node id for slot when there is no slot
		ctx.Task.Slot = t.NodeID
	}

	return
}

// Expand treats the string s as a template and populates it with values from
// the context.
func (ctx *Context) Expand(s string) (string, error) {
	tmpl, err := newTemplate(s)
	if err != nil {
		return s, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return s, err
	}

	return buf.String(), nil
}
