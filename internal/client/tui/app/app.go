package app

import (
	"github.com/rivo/tview"
)

// App представляет TUI-приложение.
type App struct {
	App   *tview.Application
	Pages *tview.Pages
}

// Primitives - структуры для хранения и передачи экранов.
type Primitives struct {
	name string
	prim func(*App) tview.Primitive
}

// NewApp создаёт новое TUI-приложение.
func NewApp(prims []Primitives) *App {
	tuiApp := &App{
		App:   tview.NewApplication(),
		Pages: tview.NewPages(),
	}

	// Добавляем экраны
	for _, p := range prims {
		tuiApp.Pages.AddPage(p.name, p.prim(tuiApp), true, true)
	}

	tuiApp.App.SetRoot(tuiApp.Pages, true)

	return tuiApp
}

// Run запускает приложение.
func (a *App) Run() error {
	return a.App.Run()
}

// SwitchTo переключает экран.
func (a *App) SwitchTo(page string) {
	a.Pages.SwitchToPage(page)
}
