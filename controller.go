// Package bubblon enables a model-stack architecture for Bubble Tea apps.
package bubblon

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
)

// Closed is a message sent to the parent model indicating that the top
// model has been closed. The message is not sent if the model is replaced.
type Closed struct{}

// openMsg is an internal message wrapping a model to open.
type openMsg struct {
	model tea.Model
}

// closeMsg is an internal message for closing the top model.
// notify indicates if the Closed message will be sent.
type closeMsg struct {
	notify bool
}

// failMsg is an internal message for setting the error in the controller
// before quitting the app.
type failMsg struct {
	err error
}

// ErrNilModel is returned when attempting to initialize a Controller with a nil model.
var ErrNilModel = errors.New("model cannot be nil")

// Open is a command to push a new model onto the stack.
// The new model will become the active model receiving updates and rendering.
func Open(model tea.Model) tea.Cmd {
	return Cmd(openMsg{model: model})
}

// Close is a command instructing bubblon to close the current model.
// A notification to the parent model is sent on closure.
func Close() tea.Msg {
	return closeMsg{notify: true}
}

// Replace combines closing the current model and opening a new one in a single command.
func Replace(model tea.Model) tea.Cmd {
	return tea.Sequence(Cmd(closeMsg{notify: false}), Open(model))
}

// Fail is a command to propagate the error to the controller and quit the app.
func Fail(err error) tea.Cmd {
	return Cmd(failMsg{err: err})
}

// Controller implements a stack-based navigation model for Bubble Tea apps.
// It manages a stack of tea.Model instances, where only the top model receives
// updates and renders. In case of the Fail message, Err is set to the error
// from the message.
type Controller struct {
	models []tea.Model
	Err    error
}

// Ensure Controller implements the tea.Model interface.
var _ tea.Model = Controller{}

// New creates a new Controller initialized with the given model.
// Returns an error if the model is nil.
func New(model tea.Model) (Controller, error) {
	if model != nil {
		return Controller{models: []tea.Model{model}, Err: nil}, nil
	}

	return Controller{}, ErrNilModel
}

// Init initializes the initial model, if one exists.
func (c Controller) Init() tea.Cmd {
	if top := c.top(); top != nil {
		return c.top().Init()
	}

	return nil
}

// Update handles incoming messages to update the state and
// delegate messages to the top model.
func (c Controller) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case openMsg:
		if msg.model != nil {
			c.push(msg.model)

			return c, tea.Batch(msg.model.Init(), tea.WindowSize())
		}

	case closeMsg:
		c.pop()

		if len(c.models) > 0 {
			if msg.notify {
				return c, tea.Batch(Cmd(Closed{}), tea.WindowSize())
			}

			return c, tea.WindowSize()
		}

	case failMsg:
		c.models = nil
		c.Err = msg.err

		return c, tea.Quit

	default:
		if top := c.top(); top != nil {
			m, cmd := top.Update(msg)
			c.models[len(c.models)-1] = m

			return c, cmd
		}
	}

	return c, nil
}

// View renders the view of the top model on the stack.
// Returns an empty string if there is no model.
func (c Controller) View() string {
	if top := c.top(); top != nil {
		return c.top().View()
	}

	return ""
}

// Cmd is a helper function that wraps a tea.Msg as a tea.Cmd.
// Should be used only for static messages.
func Cmd(msg tea.Msg) tea.Cmd {
	return func() tea.Msg { return msg }
}

func (c *Controller) push(m tea.Model) {
	c.models = append(c.models, m)
}

func (c *Controller) pop() {
	num := len(c.models)
	if num > 0 {
		c.models[num-1] = nil
		c.models = c.models[:num-1]
	}
}

func (c Controller) top() tea.Model {
	if len(c.models) > 0 {
		return c.models[len(c.models)-1]
	}

	return nil
}
