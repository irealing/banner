package main

import (
	"context"
	"sync"
)

type Application struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}
