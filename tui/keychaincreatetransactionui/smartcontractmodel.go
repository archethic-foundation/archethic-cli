package keychaincreatetransactionui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	focusedPasteSmartContractButton = focusedStyle.Copy().Render("[ Paste ]")
	blurredPasteSmartContractButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Paste"))
)

type SmartContractModel struct {
	smartContractTextAreaInput textarea.Model
	focusInput                 int
	enablePaste                bool
}

type UpdateSmartContract struct {
	Code string
}

func NewSmartContractModel() SmartContractModel {
	m := SmartContractModel{}
	m.smartContractTextAreaInput = textarea.New()
	m.smartContractTextAreaInput.CharLimit = 0
	m.smartContractTextAreaInput.MaxHeight = 0
	m.smartContractTextAreaInput.SetHeight(20)
	m.smartContractTextAreaInput.SetWidth(150)
	_, err := clipboard.ReadAll()
	if err != nil {
		m.enablePaste = false
	} else {
		m.enablePaste = true
	}
	return m
}

func (m SmartContractModel) Init() tea.Cmd {
	return nil
}

func (m SmartContractModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "esc":

			if m.smartContractTextAreaInput.Focused() {
				m.smartContractTextAreaInput.Blur()
			}
			return m, nil

		case "up", "down":
			if !m.smartContractTextAreaInput.Focused() && m.enablePaste {
				updateSmartContractFocusInput(&m, keypress)
			} else {
				return updateSmartContractValue(&m, msg)
			}

		case "enter":
			// Paste button
			if m.focusInput == 1 && m.enablePaste {
				if !m.smartContractTextAreaInput.Focused() {
					m.smartContractTextAreaInput.Focus()
					m.focusInput = 0
				}

				newText := textarea.Paste()
				return updateSmartContractValue(&m, newText)
			} else {
				return updateSmartContractValue(&m, msg)
			}
		default:

			if !m.smartContractTextAreaInput.Focused() {
				m.smartContractTextAreaInput.Focus()
				m.focusInput = 0
			}
			return updateSmartContractValue(&m, msg)
		}
	}

	return m, nil
}

func updateSmartContractValue(m *SmartContractModel, msg tea.Msg) (SmartContractModel, func() tea.Msg) {
	m.smartContractTextAreaInput, _ = m.smartContractTextAreaInput.Update(msg)
	return *m, func() tea.Msg {
		return UpdateSmartContract{Code: m.smartContractTextAreaInput.Value()}
	}
}

func updateSmartContractFocusInput(m *SmartContractModel, keypress string) {
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

func (m *SmartContractModel) SwitchTab() (SmartContractModel, []tea.Cmd) {
	return *m, nil
}

func (m SmartContractModel) View() string {
	var b strings.Builder

	b.WriteString(m.smartContractTextAreaInput.View())
	if m.smartContractTextAreaInput.Focused() {
		b.WriteString(helpStyle.Render("\npress 'esc' to exit edit mode "))
	}
	button := &blurredPasteSmartContractButton
	if m.focusInput == 1 {
		button = &focusedPasteSmartContractButton
	}
	if m.enablePaste {
		fmt.Fprintf(&b, "\n\n%s\n\n", *button)
	}
	return b.String()
}
