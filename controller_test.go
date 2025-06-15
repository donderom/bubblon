package bubblon_test

import (
	"bytes"
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/donderom/bubblon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultCheckInterval = 20 * time.Millisecond
	defaultDuration      = 3 * time.Second
	defaultView          = "view"
	secondView           = "view 2"
)

type viewUpdateMsg struct{}

type openModelMsg struct {
	model tea.Model
}

type replaceModelMsg struct {
	model tea.Model
}

type model struct {
	view string
	init bool
}

func (m *model) Init() tea.Cmd {
	m.init = true

	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case openModelMsg:
		return m, bubblon.Open(msg.model)

	case replaceModelMsg:
		return m, bubblon.Replace(msg.model)

	case viewUpdateMsg:
		m.view += " updated"

	case bubblon.Closed:
		m.view += " closed"
	}

	return m, cmd
}

func (m *model) View() string {
	return m.view
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("with valid model", func(t *testing.T) {
		t.Parallel()

		_, err := bubblon.New(newDefaultModel())
		assert.NoError(t, err)
	})

	t.Run("with nil model", func(t *testing.T) {
		t.Parallel()

		_, err := bubblon.New(nil)
		assert.ErrorIs(t, err, bubblon.ErrNilModel)
	})
}

func TestInitialModel(t *testing.T) {
	t.Parallel()

	m := newDefaultModel()
	c, _ := bubblon.New(m)

	tm := teatest.NewTestModel(t, c)
	waitForView(t, tm.Output(), defaultView)
	assert.True(t, m.init)

	tm.Send(viewUpdateMsg{})
	waitForView(t, tm.Output(), defaultView+" updated")

	require.NoError(t, tm.Quit())
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestOpen(t *testing.T) {
	t.Parallel()

	t.Run("new model", func(t *testing.T) {
		t.Parallel()

		// Init controller with the first model
		m1 := newDefaultModel()
		c, _ := bubblon.New(m1)

		m2 := newModel(secondView)

		// Open a new model and init it immediately
		tm := teatest.NewTestModel(t, c)
		tm.Send(openModelMsg{m2})
		waitForView(t, tm.Output(), secondView)
		assert.True(t, m2.init)

		// Update only the new model
		tm.Send(viewUpdateMsg{})
		waitForView(t, tm.Output(), secondView+" updated")
		assert.Equal(t, secondView+" updated", m2.view)

		// The first model is not updated
		assert.Equal(t, defaultView, m1.view)

		require.NoError(t, tm.Quit())
		tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
	})

	t.Run("nil model", func(t *testing.T) {
		t.Parallel()

		c := newController()
		c2, cmd := c.Update(bubblon.Open(nil)())
		assert.Equal(t, defaultView, c2.View())
		assert.Nil(t, cmd)
	})
}

func TestClose(t *testing.T) {
	t.Parallel()

	t.Run("initial model", func(t *testing.T) {
		t.Parallel()

		c := newController()
		assert.Equal(t, defaultView, c.View())

		c2, cmd := c.Update(bubblon.Close())
		// No more models - no more messages
		assert.Nil(t, cmd)
		assert.Empty(t, c2.View())

		c2, _ = c2.Update(viewUpdateMsg{})
		assert.Empty(t, c2.View())
	})

	t.Run("new model", func(t *testing.T) {
		t.Parallel()

		c := newController()

		m2 := newModel(secondView)

		tm := teatest.NewTestModel(t, c)
		tm.Send(openModelMsg{m2})
		waitForView(t, tm.Output(), secondView)

		tm.Send(bubblon.Close())
		// The parent model should be notified that model closed
		waitForView(t, tm.Output(), defaultView+" closed")

		require.NoError(t, tm.Quit())
		tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
	})

	t.Run("multiple times", func(t *testing.T) {
		t.Parallel()

		c := newController()
		c2, _ := c.Update(bubblon.Close())

		assert.NotPanics(t, func() { c2.Update(bubblon.Close()) })
		assert.NotPanics(t, func() { c2.Update(bubblon.Close()) })
		assert.Empty(t, c2.View())
	})
}

func TestReplace(t *testing.T) {
	t.Parallel()

	c := newController()
	m2 := newModel(secondView)
	view3 := "view 3"
	m3 := newModel(view3)

	tm := teatest.NewTestModel(t, c)
	tm.Send(openModelMsg{m2})
	tm.Send(replaceModelMsg{m3})
	waitForView(t, tm.Output(), view3)
	assert.Equal(t, secondView, m2.view)
	require.NoError(t, tm.Quit())
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func waitForView(t *testing.T, output io.Reader, view string) {
	t.Helper()

	checkInterval := teatest.WithCheckInterval(defaultCheckInterval)
	duration := teatest.WithDuration(defaultDuration)

	teatest.WaitFor(t, output, func(bts []byte) bool {
		return bytes.Contains(bts, []byte(view))
	}, checkInterval, duration)
}

func newModel(view string) *model {
	return &model{
		view: view,
		init: false,
	}
}

func newDefaultModel() *model {
	return newModel(defaultView)
}

func newController() bubblon.Controller {
	c, _ := bubblon.New(newDefaultModel())

	return c
}
