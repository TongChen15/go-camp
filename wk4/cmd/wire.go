//go:build wireinject
// +build wireinject

package main

// The build tag makes sure the stub is not built in the final build.

import (
	"github.com/TongChen15/go-camp/wk4/internal/datamodel"
	"github.com/google/wire"
)

func initializeCluster(name string) (*datamodel.Cluster, error) {
	wire.Build(datamodel.InitializeAll)
	return &datamodel.Cluster{}, nil
}
