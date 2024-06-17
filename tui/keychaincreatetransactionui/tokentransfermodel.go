package keychaincreatetransactionui

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	archethic "github.com/archethic-foundation/libgo"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TokenTransferModel struct {
	tokenInputs []textinput.Model
	focusInput  int
	transaction *archethic.TransactionBuilder
	feedback    string
}

type AddTokenTransfer struct {
	To           []byte
	Amount       *big.Int
	TokenId      uint
	TokenAddress []byte
	cmds         []tea.Cmd
}

type DeleteTokenTransfer struct {
	IndexToDelete int
}

func NewTokenTransferModel(transaction *archethic.TransactionBuilder) TokenTransferModel {
	m := TokenTransferModel{
		tokenInputs: make([]textinput.Model, 4),
		transaction: transaction,
	}
	for i := range m.tokenInputs {
		t := textinput.New()
		t.CursorStyle = cursorStyle
		switch i {
		case 0:
			t.Prompt = "> To:\n"
		case 1:
			t.Prompt = "> Amount:\n"
			t.Validate = numberValidator
		case 2:
			t.Prompt = "> Token Address:\n"
		case 3:
			t.Prompt = "> Token ID:\n"
			t.Validate = numberValidator
		}

		m.tokenInputs[i] = t
	}
	return m
}

func (m TokenTransferModel) Init() tea.Cmd {
	return nil
}

func (m TokenTransferModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "up", "down":
			updateTokenTransferFocusInput(&m, keypress)

		case "enter":
			if m.focusInput == len(m.tokenInputs) {
				toHex := m.tokenInputs[0].Value()
				to, err := hex.DecodeString(toHex)
				if toHex == "" || err != nil {
					m.feedback = "Invalid address"
					return m, nil
				}
				amountStr := m.tokenInputs[1].Value()
				amountBigInt, err := archethic.ParseBigInt(amountStr, 8)
				if err != nil {
					m.feedback = "Invalid amount"
					return m, nil
				}
				tokenAddressHex := m.tokenInputs[2].Value()
				tokenAddress, err := hex.DecodeString(tokenAddressHex)
				if tokenAddressHex == "" || err != nil {
					m.feedback = "Invalid token address"
					return m, nil
				}
				tokenIdStr := m.tokenInputs[3].Value()

				tokenId, err := strconv.Atoi(tokenIdStr)
				if err != nil {
					m.feedback = "Invalid TokenID"
					return m, nil
				}

				m.tokenInputs[0].SetValue("")
				m.tokenInputs[1].SetValue("")
				m.tokenInputs[2].SetValue("")
				m.tokenInputs[3].SetValue("")
				m, cmds := updateTokenTransferFocus(m)
				cmds = append(cmds, m.updateTokenTransferInputs(msg)...)
				return m, func() tea.Msg {
					return AddTokenTransfer{To: to, Amount: amountBigInt, TokenAddress: tokenAddress, TokenId: uint(tokenId), cmds: cmds}
				}
			}
		case "d":
			if m.focusInput > len(m.tokenInputs) {
				indexToDelete := m.focusInput - len(m.tokenInputs) - 1
				m.focusInput--
				return m, func() tea.Msg {
					return DeleteTokenTransfer{IndexToDelete: indexToDelete}
				}
			}
		}
	}
	m, cmds := updateTokenTransferFocus(m)
	cmds = append(cmds, m.updateTokenTransferInputs(msg)...)

	return m, tea.Batch(cmds...)
}

func (m *TokenTransferModel) updateTokenTransferInputs(msg tea.Msg) []tea.Cmd {

	cmds := make([]tea.Cmd, len(m.tokenInputs))
	for i := range m.tokenInputs {
		m.tokenInputs[i], cmds[i] = m.tokenInputs[i].Update(msg)
	}

	return cmds
}

func updateTokenTransferFocus(m TokenTransferModel) (TokenTransferModel, []tea.Cmd) {

	cmds := make([]tea.Cmd, len(m.tokenInputs))
	for i := 0; i <= len(m.tokenInputs)-1; i++ {
		if i == m.focusInput {
			// Set focused state
			cmds[i] = m.tokenInputs[i].Focus()
			continue
		}
		// Remove focused state
		m.tokenInputs[i].Blur()
		m.tokenInputs[i].PromptStyle = noStyle
		m.tokenInputs[i].TextStyle = noStyle
	}

	return m, cmds
}

func updateTokenTransferFocusInput(m *TokenTransferModel, keypress string) {
	if keypress == "up" {
		m.focusInput--
	} else {
		m.focusInput++
	}
	if m.focusInput > len(m.tokenInputs)+len(m.transaction.Data.Ledger.Token.Transfers) {
		m.focusInput = 0
	} else if m.focusInput < 0 {
		m.focusInput = len(m.tokenInputs) + len(m.transaction.Data.Ledger.Token.Transfers)
	}
}

func (m *TokenTransferModel) SwitchTab() (TokenTransferModel, []tea.Cmd) {
	m.focusInput = 0
	m2, cmds := updateTokenTransferFocus(*m)
	return m2, cmds
}

func (m TokenTransferModel) View() string {
	var b strings.Builder
	for i := range m.tokenInputs {
		b.WriteString(m.tokenInputs[i].View())
		if i < len(m.tokenInputs)-1 {
			b.WriteRune('\n')
		}
	}
	b.WriteRune('\n')
	b.WriteString(m.feedback)
	button := &blurredButton
	if m.focusInput == len(m.tokenInputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	startCount := len(m.tokenInputs) + 1 // +1 for the button
	for i, t := range m.transaction.Data.Ledger.Token.Transfers {
		transfer := fmt.Sprintf("%s : %f - %s %d \n", hex.EncodeToString(t.To), archethic.FormatBigInt(t.Amount, 8), hex.EncodeToString(t.TokenAddress), t.TokenId)
		if m.focusInput == startCount+i {
			b.WriteString(focusedStyle.Render(transfer))
			continue
		} else {
			b.WriteString(transfer)
		}
	}
	if len(m.transaction.Data.Ledger.Token.Transfers) > 0 {
		b.WriteString(helpStyle.Render("\npress 'd' to delete the selected token transfer "))
	}
	return b.String()
}
