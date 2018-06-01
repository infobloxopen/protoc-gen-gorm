package pep

import (
	"fmt"

	"google.golang.org/grpc/naming"
)

type staticResolver struct {
	name  string
	addrs []string
}

func newStaticResolver(name string, addrs ...string) naming.Resolver {
	return &staticResolver{
		name:  name,
		addrs: addrs,
	}
}

func (r *staticResolver) Resolve(target string) (naming.Watcher, error) {
	if target != r.name {
		return nil, fmt.Errorf("%q is an invalid target for resolver %q", target, r.name)
	}

	return newStaticWatcher(r.addrs), nil
}
