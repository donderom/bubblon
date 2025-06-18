# Bubblon

[![Release](https://img.shields.io/github/release/donderom/bubblon.svg)](https://github.com/donderom/bubblon/releases)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/donderom/bubblon)
[![build](https://github.com/donderom/bubblon/actions/workflows/build.yml/badge.svg)](https://github.com/donderom/bubblon/actions/workflows/build.yml)
[![ReportCard](https://goreportcard.com/badge/donderom/bubblon)](https://goreportcard.com/report/donderom/bubblon)
[![License](https://img.shields.io/github/license/donderom/bubblon)](https://github.com/donderom/bubblon/blob/main/LICENSE)

<p align="center">
  <img src="logo.png" width="200" alt="The Bubblon Logo">
</p>

Bubblon is a solution for managing nested [Bubble Tea](https://github.com/charmbracelet/bubbletea) models (or views/screens). This is a common use case in TUIs—for example, navigating from a main list to a sublist when an item is selected. The "canonical" way to structure this is with a view-switching model, where you keep track which view you're in (e.g., `main` or `sub`) and what item is selected, then render the appropriate model(s) for that view.

 By contrast, Bubblon uses a "model stack" architecture, where the controller determines the current model. Instead of bloating a single `Model` with state for everything, you encapsulate each view in its own `tea.Model` with its own `Update()`, `View()`, and logic. The controller then pushes/pops models on a stack as the user navigates.

### Benefits
* 📦 **Modular**: Each view is self-contained.
* 🔁 **Reusability** of sub-models.
* 🧠 **Easier to reason about**, especially when state gets complex.
* 🚫 **No new interfaces**: Keeps complexity low by avoiding new abstractions.

## Example
To run the controller, update the Bubble Tea program initialization from:

```go
mainModel := MainModel.New()
program := tea.NewProgram(mainModel, tea.WithAltScreen()) 
```

to:

```go
...
import "github.com/donderom/bubblon"
 
mainModel := MainModel.New()
controller, err := bubblon.New(mainModel)
program := tea.NewProgram(controller, tea.WithAltScreen()) 
```

At any point within the `MainModel`, you can open a new model by sending a `bubblon.Open()` command:

```go
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
...
  return m, bubblon.Open(SubModel.New())
```

There are no requirements other than for `MainModel` and `SubModel` to implement the `tea.Model` interface.

To close the current view and return to the previous one, send the `bubblon.Close` command:

```go
func (m SubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
...
  return m, bubblon.Close
```

When the `SubModel` is closed, the `MainModel` will receive a `bubblon.Closed` message.

The whole navigation is based on these two commands.

## License

[MIT](https://github.com/donderom/bubblon/raw/main/LICENSE)
