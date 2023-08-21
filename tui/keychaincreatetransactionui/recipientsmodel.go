package keychaincreatetransactionui

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	archethic "github.com/archethic-foundation/libgo"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type RecipientsModel struct {
	focusInput       int
	recipientsInputs []textinput.Model
	transaction      *archethic.TransactionBuilder
	feedback         string
}
type AddRecipient struct {
	Address  []byte
	Action   string
	ArgsJson string
	cmds     []tea.Cmd
}
type DeleteRecipient struct {
	IndexToDelete int
}

const (
	FIELD_TO     = 0
	FIELD_ACTION = 1
	FIELD_ARGS   = 2
)

func NewRecipientsModel(transaction *archethic.TransactionBuilder) RecipientsModel {
	m := RecipientsModel{
		recipientsInputs: make([]textinput.Model, 3),
		transaction:      transaction,
	}

	toInput := textinput.New()
	toInput.CursorStyle = cursorStyle
	toInput.Prompt = "> To:\n"
	m.recipientsInputs[FIELD_TO] = toInput

	actionInput := textinput.New()
	actionInput.CursorStyle = cursorStyle
	actionInput.Prompt = "> Named action [optional]:\n"
	m.recipientsInputs[FIELD_ACTION] = actionInput

	argsInput := textinput.New()
	argsInput.CursorStyle = cursorStyle
	argsInput.Prompt = "> Arguments (JSON) [optional]:\n"
	m.recipientsInputs[FIELD_ARGS] = argsInput

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
			if m.focusInput == len(m.recipientsInputs) {
				m.feedback = ""

				to := m.recipientsInputs[FIELD_TO].Value()
				action := m.recipientsInputs[FIELD_ACTION].Value()
				argsJson := m.recipientsInputs[FIELD_ARGS].Value()

				toBin, err := hex.DecodeString(to)
				if to == "" || err != nil {
					m.feedback = "Invalid address"
					return m, nil
				}

				m.recipientsInputs[FIELD_TO].SetValue("")
				m.recipientsInputs[FIELD_ACTION].SetValue("")
				m.recipientsInputs[FIELD_ARGS].SetValue("")

				m, cmds := updateRecipientsFocus(m)
				cmds = append(cmds, m.updateRecipientsInputs(msg)...)
				return m, func() tea.Msg {

					return AddRecipient{
						Address:  toBin,
						Action:   action,
						ArgsJson: argsJson,
						cmds:     cmds}
				}
			}

		case "d":

			if m.focusInput > len(m.recipientsInputs) {
				indexToDelete := m.focusInput - len(m.recipientsInputs) - 1
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

	cmds := make([]tea.Cmd, len(m.recipientsInputs))
	for i := range m.recipientsInputs {
		m.recipientsInputs[i], cmds[i] = m.recipientsInputs[i].Update(msg)
	}
	return cmds
}

func updateRecipientsFocus(m RecipientsModel) (RecipientsModel, []tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.recipientsInputs))
	for i := 0; i <= len(m.recipientsInputs)-1; i++ {
		if i == m.focusInput {
			// Set focused state
			cmds[i] = m.recipientsInputs[i].Focus()
			continue
		}
		// Remove focused state
		m.recipientsInputs[i].Blur()
		m.recipientsInputs[i].PromptStyle = noStyle
		m.recipientsInputs[i].TextStyle = noStyle
	}

	return m, cmds
}

func updateRecipientsFocusInput(m *RecipientsModel, keypress string) {
	if keypress == "up" {
		m.focusInput--
	} else {
		m.focusInput++
	}
	if m.focusInput > len(m.recipientsInputs)+len(m.transaction.Data.Recipients) {
		m.focusInput = 0
	} else if m.focusInput < 0 {
		m.focusInput = len(m.recipientsInputs) + len(m.transaction.Data.Recipients)
	}
}

func (m *RecipientsModel) SwitchTab() (RecipientsModel, []tea.Cmd) {
	m.focusInput = 0
	m2, cmds := updateRecipientsFocus(*m)
	return m2, cmds
}

func (m RecipientsModel) View() string {
	var b strings.Builder
	for i := range m.recipientsInputs {
		b.WriteString(m.recipientsInputs[i].View())
		if i < len(m.recipientsInputs)-1 {
			b.WriteRune('\n')
		}
	}
	b.WriteRune('\n')
	b.WriteString(m.feedback)
	b.WriteRune('\n')
	button := &blurredButton
	if m.focusInput == len(m.recipientsInputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	startCount := len(m.recipientsInputs) + 1 // +1 for the button
	for i, r := range m.transaction.Data.Recipients {

		argsJson, err := json.Marshal(r.Args)
		if err != nil {
			panic("invalid recipient's args")
		}

		recipientStr := fmt.Sprintf("address=%s action=%s args=%s\n", hex.EncodeToString(r.Address), r.Action, argsJson)
		if m.focusInput == startCount+i {
			b.WriteString(focusedStyle.Render(recipientStr))
			continue
		} else {
			b.WriteString(recipientStr)
		}
	}
	if len(m.transaction.Data.Recipients) > 0 {
		b.WriteString(helpStyle.Render("\npress 'd' to delete the selected recipient "))
	}
	return b.String()
}
