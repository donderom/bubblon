package bubblon_test

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultDuration = 3 * time.Second
	defaultView     = "view 1"
	secondView      = "view 2"
	updatedPrefix   = " updated"
	closedPrefix    = " closed"
)

var err = errors.New("fail")

type viewUpdateMsg struct{}

type model struct {
	view string
	init bool
}

func (m *model) Init() tea.Cmd {
	m.init = true

	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case viewUpdateMsg:
		m.view += updatedPrefix

	case bubblon.Closed:
		m.view += closedPrefix
	}

	return m, nil
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
	c.Init()

	assert.Equal(t, defaultView, c.View())
	assert.True(t, m.init)

	cm, _ := c.Update(viewUpdateMsg{})
	assert.Equal(t, updated(defaultView), cm.View())
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
		cm, _ := c.Update(bubblon.Open(m2)())
		assert.Equal(t, secondView, cm.View())
		assert.True(t, m2.init)

		// Update only the new model
		cm, _ = cm.Update(viewUpdateMsg{})
		assert.Equal(t, updated(secondView), cm.View())
		assert.Equal(t, updated(secondView), m2.view)

		// The first model is not updated
		assert.Equal(t, defaultView, m1.view)
	})

	t.Run("nil model", func(t *testing.T) {
		t.Parallel()

		c := newController()
		cm, cmd := c.Update(bubblon.Open(nil)())
		assert.Equal(t, defaultView, cm.View())
		assert.Nil(t, cmd)
	})
}

func TestClose(t *testing.T) {
	t.Parallel()

	t.Run("initial model", func(t *testing.T) {
		t.Parallel()

		c := newController()
		assert.Equal(t, defaultView, c.View())

		cm, cmd := c.Update(bubblon.Close())
		// No more models - no more messages
		assert.Nil(t, cmd)
		assert.Empty(t, cm.View())

		cm, _ = cm.Update(viewUpdateMsg{})
		assert.Empty(t, cm.View())
	})

	t.Run("new model", func(t *testing.T) {
		t.Parallel()

		c := newController()
		m2 := newModel(secondView)

		cm, _ := c.Update(bubblon.Open(m2)())
		assert.Equal(t, secondView, cm.View())

		// The parent model should be notified that model closed
		cm = closeModel(cm)
		assert.Equal(t, closed(defaultView), cm.View())
	})

	t.Run("multiple times", func(t *testing.T) {
		t.Parallel()

		c := newController()
		cm, _ := c.Update(bubblon.Close())

		assert.NotPanics(t, func() { cm.Update(bubblon.Close()) })
		assert.NotPanics(t, func() { cm.Update(bubblon.Close()) })
		assert.Empty(t, cm.View())
	})
}

func TestReplace(t *testing.T) {
	t.Parallel()

	c := newController()
	m2 := newModel(secondView)
	view3 := "view 3"
	m3 := newModel(view3)

	cm, _ := c.Update(bubblon.Open(m2)())
	cm, _ = cm.Update(bubblon.Replace(m3)())
	assert.Equal(t, view3, cm.View())
	assert.Equal(t, secondView, m2.view)

	cm = closeModel(cm)
	assert.Equal(t, closed(defaultView), cm.View())
}

func TestReplaceAll(t *testing.T) {
	t.Parallel()

	c := newController()
	m2 := newModel(secondView)
	models := 3
	tempSuffix := "tempview "

	var cm tea.Model = c
	for i := range models {
		cm, _ = cm.Update(bubblon.Open(newModel(tempSuffix + strconv.Itoa(i)))())
	}
	assert.Equal(t, tempSuffix+strconv.Itoa(models-1), cm.View())

	cm, _ = cm.Update(bubblon.Replace(m2)())
	assert.Equal(t, secondView, cm.View())
}

func TestFail(t *testing.T) {
	t.Parallel()

	c := newController()
	require.NoError(t, c.Err)

	cm, cmd := c.Update(bubblon.Fail(err)())
	assert.IsType(t, tea.QuitMsg{}, cmd())

	fc, ok := cm.(bubblon.Controller)
	assert.True(t, ok)
	assert.Equal(t, err, fc.Err)
}

func TestInterrupt(t *testing.T) {
	t.Parallel()

	done := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), defaultDuration)
	defer cancel()
	c := newController()
	p := tea.NewProgram(c, tea.WithAltScreen(), tea.WithContext(ctx))
	go func() {
		_, err := p.Run()
		done <- err
	}()

	p.Send(bubblon.Close())
	p.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.Error(t, <-done)
	assert.Nil(t, ctx.Err())
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

func updated(view string) string {
	return view + updatedPrefix
}

func closed(view string) string {
	return view + closedPrefix
}

func closeModel(ctrl tea.Model) tea.Model {
	ctrl, cmd := ctrl.Update(bubblon.Close())

	switch msg := cmd().(type) {
	case tea.BatchMsg:
		for _, subcmd := range msg {
			ctrl, _ = ctrl.Update(subcmd())
		}
	}

	return ctrl
}
