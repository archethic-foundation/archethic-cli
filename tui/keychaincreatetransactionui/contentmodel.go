package keychaincreatetransactionui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	focusedPasteContentButton = focusedStyle.Copy().Render("[ Paste ]")
	blurredPasteContentButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Paste"))
)

type ContentModel struct {
	contentTextAreaInput textarea.Model
	focusInput           int
	enablePaste          bool
}

type UpdateContent struct {
	Content []byte
}

func NewContentModel() ContentModel {
	m := ContentModel{}
	m.contentTextAreaInput = textarea.New()
	m.contentTextAreaInput.CharLimit = 0
	m.contentTextAreaInput.MaxHeight = 0
	m.contentTextAreaInput.SetHeight(20)
	m.contentTextAreaInput.SetWidth(150)
	_, err := clipboard.ReadAll()
	if err != nil {
		m.enablePaste = false
	} else {
		m.enablePaste = true
	}
	return m
}

func (m ContentModel) Init() tea.Cmd {
	return nil
}

func (m ContentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {

		case "esc":

			if m.contentTextAreaInput.Focused() {
				m.contentTextAreaInput.Blur()
			}
			return m, nil

		case "up", "down":
			if !m.contentTextAreaInput.Focused() && m.enablePaste {
				updateContentFocusInput(&m, keypress)
			} else {
				return updateContentValue(&m, msg)
			}

		case "enter":
			// Paste button
			if m.focusInput == 1 && m.enablePaste {
				if !m.contentTextAreaInput.Focused() {
					m.contentTextAreaInput.Focus()
					m.focusInput = 0
				}

				newText := textarea.Paste()
				return updateContentValue(&m, newText)
			} else {
				return updateContentValue(&m, msg)
			}
		default:

			if !m.contentTextAreaInput.Focused() {
				m.contentTextAreaInput.Focus()
				m.focusInput = 0
			}
			return updateContentValue(&m, msg)
		}
	}

	return m, nil
}

func updateContentValue(m *ContentModel, msg tea.Msg) (ContentModel, func() tea.Msg) {
	m.contentTextAreaInput, _ = m.contentTextAreaInput.Update(msg)
	return *m, func() tea.Msg {
		return UpdateContent{Content: []byte(m.contentTextAreaInput.Value())}
	}
}

func updateContentFocusInput(m *ContentModel, keypress string) {
	if keypress == "up" {
		m.focusInput--
	} else {
		m.focusInput++
	}
	if m.focusInput > 1 {
		m.focusInput = 0
	} else if m.focusInput < 0 {
		m.focusInput = 1
	}
}

func (m *ContentModel) SwitchTab() (ContentModel, []tea.Cmd) {
	return *m, nil
}

func (m ContentModel) View() string {
	var b strings.Builder

	b.WriteString(m.contentTextAreaInput.View())

	if m.contentTextAreaInput.Focused() {
		b.WriteString(helpStyle.Render("\npress 'esc' to exit edit mode "))
	}
	button := &blurredPasteContentButton
	if m.focusInput == 1 {
		button = &focusedPasteContentButton
	}
	if m.enablePaste {
		fmt.Fprintf(&b, "\n\n%s\n\n", *button)
	}
	return b.String()
}
