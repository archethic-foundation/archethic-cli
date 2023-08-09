package keychaincreatetransactionui

import (
	"encoding/hex"
	"fmt"
	"strings"

	archethic "github.com/archethic-foundation/libgo"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type RecipientsModel struct {
	focusInput      int
	recipientsInput textinput.Model
	transaction     *archethic.TransactionBuilder
	feedback        string
}
type AddRecipient struct {
	Recipient []byte
	cmds      []tea.Cmd
}
type DeleteRecipient struct {
	IndexToDelete int
}

func NewRecipientsModel(transaction *archethic.TransactionBuilder) RecipientsModel {
	m := RecipientsModel{transaction: transaction}
	m.recipientsInput = textinput.New()
	m.recipientsInput.CursorStyle = cursorStyle
	m.recipientsInput.Prompt = "> Recipient address:\n"
	return m

}

func (m RecipientsModel) Init() tea.Cmd {
	return nil
}

func (m RecipientsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {

		case "up", "down":
			updateRecipientsFocusInput(&m, keypress)

		case "enter":

			if m.focusInput == 1 || m.focusInput == 0 {
				m.feedback = ""
				recipientHex := m.recipientsInput.Value()
				recipient, err := hex.DecodeString(recipientHex)
				if err != nil || recipientHex == "" {
					m.feedback = "Invalid recipient address"
					return m, nil
				}
				m.recipientsInput.SetValue("")
				m, cmds := updateRecipientsFocus(m)
				cmds = append(cmds, m.updateRecipientsInputs(msg)...)
				return m, func() tea.Msg {
					return AddRecipient{Recipient: recipient, cmds: cmds}
				}
			}

		case "d":

			if m.focusInput > 1 {
				indexToDelete := m.focusInput - 2
				m.focusInput--
				return m, func() tea.Msg {
					return DeleteRecipient{IndexToDelete: indexToDelete}
				}
			}

		}
	}
	m, cmds := updateRecipientsFocus(m)
	cmds = append(cmds, m.updateRecipientsInputs(msg)...)

	return m, tea.Batch(cmds...)
}

func (m *RecipientsModel) updateRecipientsInputs(msg tea.Msg) []tea.Cmd {

	cmds := make([]tea.Cmd, 1)
	m.recipientsInput, cmds[0] = m.recipientsInput.Update(msg)

	return cmds
}

func updateRecipientsFocus(m RecipientsModel) (RecipientsModel, []tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	if m.focusInput == 0 {
		cmds = append(cmds, m.recipientsInput.Focus())
	} else {
		m.recipientsInput.Blur()
		m.recipientsInput.PromptStyle = noStyle
		m.recipientsInput.TextStyle = noStyle
	}

	return m, cmds
}

func updateRecipientsFocusInput(m *RecipientsModel, keypress string) {
	if keypress == "up" {
		m.focusInput--
	} else {
		m.focusInput++
	}
	// 1 because : first input [0] is the recipient address and second [1] is the button
	if m.focusInput > 1+len(m.transaction.Data.Recipients) {
		m.focusInput = 0
	} else if m.focusInput < 0 {
		m.focusInput = 1 + len(m.transaction.Data.Recipients)
	}
}

func (m *RecipientsModel) SwitchTab() (RecipientsModel, []tea.Cmd) {
	m.recipientsInput.Focus()
	m2, cmds := updateRecipientsFocus(*m)
	return m2, cmds
}

func (m RecipientsModel) View() string {
	var b strings.Builder
	b.WriteString(m.recipientsInput.View())
	b.WriteRune('\n')
	b.WriteString(m.feedback)
	b.WriteRune('\n')
	button := &blurredButton
	if m.focusInput == 1 {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	startCount := 2 // 1 for the input, 1 for the button
	for i, t := range m.transaction.Data.Recipients {
		recipient := fmt.Sprintf("%s\n", hex.EncodeToString(t))
		if m.focusInput == startCount+i {
			b.WriteString(focusedStyle.Render(recipient))
			continue
		} else {
			b.WriteString(recipient)
		}
	}
	if len(m.transaction.Data.Recipients) > 0 {
		b.WriteString(helpStyle.Render("\npress 'd' to delete the selected recipient "))
	}
	return b.String()
}
