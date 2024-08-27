package keychaincreatetransactionui

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	archethic "github.com/archethic-foundation/libgo"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type UcoTransferModel struct {
	ucoInputs   []textinput.Model
	focusInput  int
	transaction *archethic.TransactionBuilder
	feedback    string
}

type AddUcoTransfer struct {
	To     []byte
	Amount *big.Int
	cmds   []tea.Cmd
}
type DeleteUcoTransfer struct {
	IndexToDelete int
}

func NewUcoTransferModel(transaction *archethic.TransactionBuilder) UcoTransferModel {

	m := UcoTransferModel{
		ucoInputs:   make([]textinput.Model, 2),
		transaction: transaction,
	}
	for i := range m.ucoInputs {
		t := textinput.New()
		t.CursorStyle = cursorStyle
		switch i {
		case 0:
			t.Prompt = "> To:\n"
		case 1:
			t.Prompt = "> Amount:\n"
			t.Validate = numberValidator
		}

		m.ucoInputs[i] = t
	}

	return m

}

func (m UcoTransferModel) Init() tea.Cmd {
	return nil
}

func (m UcoTransferModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {

		case "up", "down":
			updateUcoTransferFocusInput(&m, keypress)

		case "enter":

			if m.focusInput == len(m.ucoInputs) {
				m.feedback = ""
				toHex := m.ucoInputs[0].Value()
				to, err := hex.DecodeString(toHex)
				if toHex == "" || err != nil {
					m.feedback = "Invalid address"
					return m, nil
				}
				amountStr := m.ucoInputs[1].Value()
				amountBigInt, err := archethic.ParseBigInt(amountStr, 8)
				if err != nil {
					m.feedback = "Invalid amount"
					return m, nil
				}
				m.ucoInputs[0].SetValue("")
				m.ucoInputs[1].SetValue("")

				m, cmds := updateUcoTransferFocus(m)
				cmds = append(cmds, m.updateUcoTransferInputs(msg)...)

				return m, func() tea.Msg {
					return AddUcoTransfer{
						To:     to,
						Amount: amountBigInt,
						cmds:   cmds,
					}
				}
			}
		case "d":

			if m.focusInput > len(m.ucoInputs) {
				indexToDelete := m.focusInput - len(m.ucoInputs) - 1
				m.focusInput--
				return m, func() tea.Msg {
					return DeleteUcoTransfer{
						IndexToDelete: indexToDelete,
					}
				}

			}

		}
	}
	m, cmds := updateUcoTransferFocus(m)
	cmds = append(cmds, m.updateUcoTransferInputs(msg)...)

	return m, tea.Batch(cmds...)
}

func (m *UcoTransferModel) updateUcoTransferInputs(msg tea.Msg) []tea.Cmd {

	cmds := make([]tea.Cmd, len(m.ucoInputs))
	for i := range m.ucoInputs {
		m.ucoInputs[i], cmds[i] = m.ucoInputs[i].Update(msg)
	}

	return cmds
}

func updateUcoTransferFocus(m UcoTransferModel) (UcoTransferModel, []tea.Cmd) {

	cmds := make([]tea.Cmd, len(m.ucoInputs))
	for i := 0; i <= len(m.ucoInputs)-1; i++ {
		if i == m.focusInput {
			// Set focused state
			cmds[i] = m.ucoInputs[i].Focus()
			continue
		}
		// Remove focused state
		m.ucoInputs[i].Blur()
		m.ucoInputs[i].PromptStyle = noStyle
		m.ucoInputs[i].TextStyle = noStyle
	}

	return m, cmds
}

func updateUcoTransferFocusInput(m *UcoTransferModel, keypress string) {
	if keypress == "up" {
		m.focusInput--
	} else {
		m.focusInput++
	}

	if m.focusInput > len(m.ucoInputs)+len(m.transaction.Data.Ledger.Uco.Transfers) {
		m.focusInput = 0
	} else if m.focusInput < 0 {
		m.focusInput = len(m.ucoInputs) + len(m.transaction.Data.Ledger.Uco.Transfers)
	}

}

func (m *UcoTransferModel) SwitchTab() (UcoTransferModel, []tea.Cmd) {
	m.focusInput = 0
	m2, cmds := updateUcoTransferFocus(*m)
	return m2, cmds
}

func (m UcoTransferModel) View() string {
	var b strings.Builder
	for i := range m.ucoInputs {
		b.WriteString(m.ucoInputs[i].View())
		if i < len(m.ucoInputs)-1 {
			b.WriteRune('\n')
		}
	}
	b.WriteRune('\n')
	b.WriteString(m.feedback)
	button := &blurredButton
	if m.focusInput == len(m.ucoInputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	startCount := len(m.ucoInputs) + 1 // +1 for the button
	for i, t := range m.transaction.Data.Ledger.Uco.Transfers {
		transfer := fmt.Sprintf("%s: %s\n", hex.EncodeToString(t.To), archethic.FormatBigInt(t.Amount, 8))
		if m.focusInput == startCount+i {
			b.WriteString(focusedStyle.Render(transfer))
			continue
		} else {
			b.WriteString(transfer)
		}
	}
	if len(m.transaction.Data.Ledger.Uco.Transfers) > 0 {
		b.WriteString(helpStyle.Render("\npress 'd' to delete the selected UCO transfer "))
	}
	return b.String()
}
