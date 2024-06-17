package keychaincreatetransactionui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	archethic "github.com/archethic-foundation/libgo"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	urlType = []string{"Local", "Testnet", "Mainnet", "Custom"}
	urls    = map[string]string{
		"Local":   "http://localhost:4000",
		"Testnet": "https://testnet.archethic.net",
		"Mainnet": "https://mainnet.archethic.net",
		"Custom":  ""}
	transactionTypesList = []string{
		"Keychain Access",
		"Keychain",
		"Transfer",
		"Hosting",
		"Token",
		"Data",
		"Contract",
		"Code Proposal",
		"Code Approval",
	}
	transactionTypes = map[string]archethic.TransactionType{
		"Keychain Access": archethic.KeychainAccessType,
		"Keychain":        archethic.KeychainType,
		"Transfer":        archethic.TransferType,
		"Hosting":         archethic.HostingType,
		"Token":           archethic.TokenType,
		"Data":            archethic.DataType,
		"Contract":        archethic.ContractType,
		"Code Proposal":   archethic.CodeProposalType,
		"Code Approval":   archethic.CodeApprovalType,
	}
)

type MainModel struct {
	mainInputs              []textinput.Model
	selectedUrl             string
	serviceMode             bool
	serviceName             string
	selectedTransactionType string
	focusInput              int
	feedback                string
	pvKeyBytes              []byte
}

type UpdateTransactionIndex struct {
	Index uint
	cmds  []tea.Cmd
}

type UpdateUrl struct {
	Url  string
	cmds []tea.Cmd
}

type UpdateTransactionType struct {
	TransactionType archethic.TransactionType
	cmds            []tea.Cmd
}

type SendTransaction struct {
	Curve archethic.Curve
	Seed  []byte
}

type GetTransactionFee struct {
	Curve archethic.Curve
	Seed  []byte
}

type ResetInterface struct{}

func NewMainModel(pvKeyBytes []byte) MainModel {

	m := MainModel{
		mainInputs: make([]textinput.Model, 5),
		pvKeyBytes: pvKeyBytes,
	}

	for i := range m.mainInputs {
		t := textinput.New()
		t.CursorStyle = cursorStyle
		switch i {
		case 0:
			t.Prompt = ""
		case 1:
			t.Prompt = "> Access seed\n"
			if pvKeyBytes != nil {
				t.Placeholder = "(Imported SSH key)"
			} else {
				t.EchoMode = textinput.EchoPassword
				t.EchoCharacter = '•'
			}
		case 2:
			t.Prompt = "> Elliptic curve\n"
			t.Placeholder = "(default 0)"
			t.CharLimit = 1
			t.Validate = curveValidator
		case 3:
			t.Prompt = "> Index\n"
			t.Placeholder = "(default 0)"
			t.Validate = numberValidator
		case 4:
			t.Prompt = ""
		}
		m.mainInputs[i] = t
	}

	return m
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case CreateTransactionMsg:
		m.mainInputs[1].SetValue(msg.Seed)
		m.mainInputs[0].SetValue(msg.Url)
		m.serviceName = msg.ServiceName
		m.serviceMode = m.serviceName != ""
		if m.serviceMode {
			m.focusInput = FIRST_TRANSACTION_TYPE_INDEX
		}
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {

		case "up", "down":
			previousFocus := m.focusInput
			updateMainFocusInput(&m, keypress)
			// if the seed or the curve are blured, they are probably updated, so update the transaction index
			if previousFocus == SEED_INDEX || previousFocus == CURVE_INDEX {
				var seed []byte
				if m.pvKeyBytes != nil {
					seed = m.pvKeyBytes
				} else {
					var err error
					seed, err = archethic.MaybeConvertToHex(m.mainInputs[1].Value())
					if err != nil {
						m.feedback = err.Error()
						return m, nil
					}
				}

				index, err := tuiutils.GetLastTransactionIndex(m.mainInputs[0].Value(), getCurve(&m), seed)
				if err != nil {
					m.feedback = err.Error()
					return m, nil
				}
				m.mainInputs[3].SetValue(fmt.Sprint(index))
				m, cmds := updateMainFocus(m)
				cmds = append(cmds, m.updateMainInputs(msg)...)
				return m, func() tea.Msg {
					return UpdateTransactionIndex{Index: index, cmds: cmds}
				}
			}
		case "enter":

			if m.focusInput < URL_INDEX {
				u := urlType[m.focusInput]
				m.mainInputs[0].SetValue(urls[u])
				m.selectedUrl = u
				m.focusInput = URL_INDEX
				m, cmds := updateMainFocus(m)
				cmds = append(cmds, m.updateMainInputs(msg)...)
				return m, func() tea.Msg {
					return UpdateUrl{Url: urls[u], cmds: cmds}
				}
			} else if m.focusInput > TRANSACTION_INDEX_FIELD_INDEX && m.focusInput < MAIN_ADD_BUTTON_INDEX {
				m.selectedTransactionType = transactionTypesList[m.focusInput-FIRST_TRANSACTION_TYPE_INDEX]
				m.focusInput = MAIN_ADD_BUTTON_INDEX
				m, cmds := updateMainFocus(m)
				cmds = append(cmds, m.updateMainInputs(msg)...)
				return m, func() tea.Msg {
					return UpdateTransactionType{TransactionType: transactionTypes[m.selectedTransactionType], cmds: cmds}
				}
			} else if m.focusInput == MAIN_ADD_BUTTON_INDEX {
				return m, func() tea.Msg {
					var seed []byte
					if m.pvKeyBytes != nil {
						seed = m.pvKeyBytes
					} else {
						var err error
						seed, err = archethic.MaybeConvertToHex(m.mainInputs[1].Value())
						if err != nil {
							m.feedback = err.Error()
							return m
						}
					}
					return SendTransaction{Curve: getCurve(&m), Seed: seed}
				}
			} else if m.focusInput == MAIN_RESET_BUTTON_INDEX {
				return m, func() tea.Msg {
					return ResetInterface{}
				}
			} else if m.focusInput == MAIN_GET_TRANSACTION_FEE_BUTTON_INDEX {
				return m, func() tea.Msg {
					var seed []byte
					if m.pvKeyBytes != nil {
						seed = m.pvKeyBytes
					} else {
						var err error
						seed, err = archethic.MaybeConvertToHex(m.mainInputs[1].Value())
						if err != nil {
							m.feedback = err.Error()
							return m
						}
					}
					return GetTransactionFee{Curve: getCurve(&m), Seed: seed}
				}
			}

		default:
			if m.focusInput == URL_INDEX {
				return m, func() tea.Msg {
					return UpdateUrl{Url: m.mainInputs[0].Value()}
				}
			}
		}
	}
	m, cmds := updateMainFocus(m)
	cmds = append(cmds, m.updateMainInputs(msg)...)

	return m, tea.Batch(cmds...)
}

func getCurve(m *MainModel) archethic.Curve {
	curveInt, err := strconv.Atoi(m.mainInputs[2].Value())
	if err != nil {
		curveInt = 0
	}
	return archethic.Curve(curveInt)
}

func updateMainFocusInput(m *MainModel, keypress string) {
	if keypress == "up" {
		m.focusInput--
	} else {
		m.focusInput++
	}

	if m.serviceMode {
		if m.focusInput > MAIN_RESET_BUTTON_INDEX {
			m.focusInput = FIRST_TRANSACTION_TYPE_INDEX
		} else if m.focusInput < FIRST_TRANSACTION_TYPE_INDEX {
			m.focusInput = MAIN_RESET_BUTTON_INDEX
		}
	} else {
		if m.focusInput > MAIN_RESET_BUTTON_INDEX {
			m.focusInput = 0
		} else if m.focusInput < 0 {
			m.focusInput = MAIN_RESET_BUTTON_INDEX
		}
	}
}

func (m *MainModel) updateMainInputs(msg tea.Msg) []tea.Cmd {

	cmds := make([]tea.Cmd, len(m.mainInputs))
	for i := range m.mainInputs {
		m.mainInputs[i], cmds[i] = m.mainInputs[i].Update(msg)
	}

	return cmds
}

func updateMainFocus(m MainModel) (MainModel, []tea.Cmd) {

	cmds := make([]tea.Cmd, len(m.mainInputs))
	for i := 0; i <= len(m.mainInputs)-1; i++ {
		// the first 4 inputs are not focusable fields (node endpoints for URL)

		if i == m.focusInput-len(urlType) {
			// Set focused state
			cmds[i] = m.mainInputs[i].Focus()
			continue
		}
		// Remove focused state
		m.mainInputs[i].Blur()
		m.mainInputs[i].PromptStyle = noStyle
		m.mainInputs[i].TextStyle = noStyle
	}

	return m, cmds
}

func (m *MainModel) SwitchTab() (MainModel, []tea.Cmd) {
	m.focusInput = 0
	m2, cmds := updateMainFocus(*m)
	return m2, cmds
}

func (m MainModel) View() string {
	var b strings.Builder
	// Only display the node endpoint, seed, curve and index fields if we are not building a transaction for a service
	if m.serviceMode {
		b.WriteString("Creating transaction for service " + m.serviceName + "\n\n")
	} else {
		b.WriteString("> Node endpoint:\n")
		b.WriteString(urlView(m))
		// url field
		b.WriteString(m.mainInputs[0].View() + "\n\n")
		// seed field
		b.WriteString(m.mainInputs[1].View() + "\n\n")
		// curve field
		b.WriteString(m.mainInputs[2].View() + "\n\n")
		for j := 0; j <= 2; j++ {
			b.WriteString("\t (" + strconv.Itoa(j) + ") " + tuiutils.GetCurveName(archethic.Curve(j)) + "\n")
		}
		// index field
		b.WriteString(m.mainInputs[3].View() + "\n\n")
	}

	// transaction type field
	b.WriteString("> Transaction type:\n")
	b.WriteString(transactionTypeView(m))

	// send transaction button
	button := &blurredButton
	if m.focusInput == MAIN_ADD_BUTTON_INDEX {
		button = &focusedButton
	}

	// get transaction fee button
	getTransactionFeeButton := &blurredGetTransactionFeeButton
	if m.focusInput == MAIN_GET_TRANSACTION_FEE_BUTTON_INDEX {
		getTransactionFeeButton = &focusedGetTransactionFeeButton
	}

	if m.feedback != "" {
		b.WriteString(m.feedback + "\n\n")
	}

	fmt.Fprintf(&b, "\n\n%s\n\n", *button)
	fmt.Fprintf(&b, "%s\n\n", *getTransactionFeeButton)

	// reset button
	resetButton := &blurredResetButton
	if m.focusInput == MAIN_RESET_BUTTON_INDEX {
		resetButton = &focusedResetButton
	}
	fmt.Fprintf(&b, "%s\n\n", *resetButton)

	return b.String()
}

// URL part of the main tab
// looping through the urlType array to display the different options
func urlView(m MainModel) string {
	s := strings.Builder{}

	for i := 0; i < len(urlType); i++ {
		var u string

		if m.selectedUrl == urlType[i] {
			u = "(•) "
		} else {
			u = "( ) "
		}
		u += urlType[i]

		if i == m.focusInput {
			s.WriteString(focusedStyle.Render(u))
		} else {
			s.WriteString(u)
		}
		s.WriteString("\n")
	}

	return s.String()
}

// Transaction type part of the main tab
func transactionTypeView(m MainModel) string {
	s := strings.Builder{}

	for i, t := range transactionTypesList {
		var u string
		if m.selectedTransactionType == t {
			u = "(•) "
		} else {
			u = "( ) "
		}
		u += t
		if m.focusInput == i+FIRST_TRANSACTION_TYPE_INDEX {
			s.WriteString(focusedStyle.Render(u))
		} else {
			s.WriteString(u)
		}
		s.WriteString("\n")
	}

	return s.String()
}
