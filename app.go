package main

import (
	"context"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

// Called by Wails when the app starts; keep ctx to use runtime APIs later
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}
