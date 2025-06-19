<div align="center">

# Bubblon

[![Release](https://img.shields.io/github/v/release/donderom/bubblon.svg?style=flat-square&color=019aca)](https://github.com/donderom/bubblon/releases)
[![GoDoc](https://img.shields.io/badge/go.dev-docs-01ade3?style=flat-square&logo=go)](https://pkg.go.dev/github.com/donderom/bubblon)
[![Build](https://img.shields.io/github/actions/workflow/status/donderom/bubblon/build.yml?style=flat-square&logo=github&color=b199da)](https://github.com/donderom/bubblon/actions/workflows/build.yml)
[![License](https://img.shields.io/badge/license-MIT-fec4e2?style=flat-square)](https://github.com/donderom/bubblon/blob/main/LICENSE)
[![ReportCard](https://goreportcard.com/badge/github.com/donderom/bubblon?style=flat-square)](https://goreportcard.com/report/donderom/bubblon)

<img src="logo.png" width="200" alt="The Bubblon Logo">
</div>

Bubblon is a solution for managing nested [Bubble Tea](https://github.com/charmbracelet/bubbletea) models (or views/screens). This is a common use case in TUIs—for example, navigating from a main list to a sublist when an item is selected. The "canonical" way to structure this is with a view-switching model, where you keep track which view you're in (e.g., `main` or `sub`) and what item is selected, then render the appropriate model(s) for that view.

 By contrast, Bubblon uses a "model stack" architecture, where the controller determines the current model. Instead of bloating a single `Model` with state for everything, you encapsulate each view in its own `tea.Model` with its own `Update()`, `View()`, and logic. The controller then pushes/pops models on a stack as the user navigates.

### Benefits
* 📦 **Modular**: Each view is self-contained.
* 🔁 **Reusability** of sub-models.
* 🧠 **Easier to reason about**, especially when state gets complex.
* 🚫 **No new interfaces**: Keeps complexity low by avoiding new abstractions.

## Installation

To install Bubblon, use `go get`:

```sh
go get github.com/donderom/bubblon
```

Import the `bubblon` package into your code:

```sh
import "github.com/donderom/bubblon"
```

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
