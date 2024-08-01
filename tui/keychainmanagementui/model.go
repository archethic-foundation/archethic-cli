package keychainmanagementui

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/archethic-foundation/archethic-cli/tui/keychaincreatetransactionui"
	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	archethic "github.com/archethic-foundation/libgo"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BackMsg change state back to project view
type BackMsg bool
type SendNewKeychainTransaction struct {
	Model Model
}
type SendAccessKeychain struct {
	Model Model
}
type SendRemoveService struct {
	Model Model
}
type SendCreateService struct {
	Model Model
}

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	accessFocusedButton = focusedStyle.Copy().Render("[ Access Keychain ]")
	accessBlurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Access Keychain"))

	createFocusedButton = focusedStyle.Copy().Render("[ Create Keychain ]")
	createBlurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Create Keychain"))

	createTransactionFocusedButton       = focusedStyle.Copy().Render("[ Create Transaction for Service ]")
	createTransactionaccessBlurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Create Transaction for Service"))
	createServiceFocusedButton           = focusedStyle.Copy().Render("[ Create Service ]")
	createServiceBlurredButton           = fmt.Sprintf("[ %s ]", blurredStyle.Render("Create Service"))
	urlType                              = []string{"Local", "Testnet", "Mainnet", "Custom"}
	urls                                 = map[string]string{
		"Local":   "http://localhost:4000",
		"Testnet": "https://testnet.archethic.net",
		"Mainnet": "https://mainnet.archethic.net",
		"Custom":  ""}
)

type Model struct {
	IsInit                           bool
	focusIndex                       int
	inputs                           []textinput.Model
	newServiceInputs                 []textinput.Model
	keychain                         *archethic.Keychain
	serviceNames                     []string
	selectedUrl                      string
	selectedService                  int
	keychainSeed                     string
	keychainTransactionAddress       string
	keychainAccessTransactionAddress string
	feedback                         string
	showSpinnerCreate                bool
	showSpinnerAccess                bool
	showSpinnerDeleteService         bool
	showSpinnerCreateService         bool
	Spinner                          spinner.Model
	pvKeyBytes                       []byte
}

func New(pvKeyBytes []byte) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	m := Model{
		inputs:           make([]textinput.Model, 2),
		Spinner:          s,
		newServiceInputs: make([]textinput.Model, 2),
		pvKeyBytes:       pvKeyBytes,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
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
		}

		m.inputs[i] = t
	}

	for i := range m.newServiceInputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle

		switch i {
		case 0:
			t.Prompt = "> Service name\n"
		case 1:
			t.Prompt = "> Derivation path\n"
		}

		m.newServiceInputs[i] = t
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return m.Spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	urlBlockSize := len(urlType)
	switch msg := msg.(type) {
	case SendNewKeychainTransaction:
		m.feedback = msg.Model.feedback
		m.keychainSeed = msg.Model.keychainSeed
		m.keychainTransactionAddress = msg.Model.keychainTransactionAddress
		m.keychainAccessTransactionAddress = msg.Model.keychainAccessTransactionAddress
		m.showSpinnerCreate = false
		return m, nil
	case SendAccessKeychain:
		m.keychain = msg.Model.keychain
		m.feedback = msg.Model.feedback
		m.serviceNames = msg.Model.serviceNames
		m.selectedService = msg.Model.selectedService
		m.showSpinnerAccess = false
		return m, nil
	case SendRemoveService:
		m.keychain = msg.Model.keychain
		m.feedback = msg.Model.feedback
		m.serviceNames = msg.Model.serviceNames
		m.selectedService = msg.Model.selectedService
		m.showSpinnerDeleteService = false
		return m, nil
	case SendCreateService:
		m.keychain = msg.Model.keychain
		m.feedback = msg.Model.feedback
		m.serviceNames = msg.Model.serviceNames
		m.selectedService = msg.Model.selectedService
		m.showSpinnerCreateService = false
		cmds := m.updateInputs(msg)
		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			return New(m.pvKeyBytes), func() tea.Msg {
				return BackMsg(true)
			}

		case "enter":
			if m.focusIndex < urlBlockSize {
				u := urlType[m.focusIndex]
				m.inputs[0].SetValue(urls[u])
				m.selectedUrl = u
				m.focusIndex = urlBlockSize
			}

			// create keychain button
			if m.focusIndex == len(m.inputs)+urlBlockSize {
				m.feedback = ""
				accessSeed, err := getAccessKey(m)
				if err != nil {
					m.feedback = err.Error()
					return m, nil
				}
				err = validateInput(m.inputs[0].Value(), accessSeed)
				if err != nil {
					m.feedback = err.Error()
					return m, nil
				}
				m.showSpinnerCreate = true
				return m, func() tea.Msg {
					return SendNewKeychainTransaction{createKeychain(&m)}
				}
			}

			// access keychain button
			if m.focusIndex == len(m.inputs)+5 {
				m.showSpinnerAccess = true
				m.feedback = ""
				m.keychainSeed = ""
				return m, func() tea.Msg {
					return SendAccessKeychain{accessKeychain(&m)}
				}
			}

			// add service
			if m.focusIndex == len(m.inputs)+6+len(m.serviceNames)+1+len(m.newServiceInputs) {
				m.showSpinnerCreateService = true
				return m, func() tea.Msg {
					return SendCreateService{addService(&m)}
				}
			}

			// select service
			if m.focusIndex > len(m.inputs)+urlBlockSize+1 && m.focusIndex < len(m.inputs)+urlBlockSize+2+len(m.serviceNames) {
				m.selectedService = m.focusIndex - len(m.inputs) - urlBlockSize - 2
			}

			// redirect to create transaction
			if m.focusIndex == len(m.inputs)+6+len(m.serviceNames) {
				return m, func() tea.Msg {
					accessKey, err := getAccessKey(m)
					if err != nil {
						m.feedback = err.Error()
						return nil
					}
					return keychaincreatetransactionui.CreateTransactionMsg{
						ServiceName: m.serviceNames[m.selectedService],
						Url:         m.inputs[0].Value(),
						Seed:        hex.EncodeToString(accessKey),
						PvKeyBytes:  m.pvKeyBytes,
					}
				}
			}
		// Set focus to next input
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else if s == "down" || s == "tab" {
				m.focusIndex++
			}

			serviceSize := len(m.serviceNames)
			// if there is at least one service, there is a button to create a transaction
			if serviceSize > 0 {
				serviceSize++
			}
			newServiceFormSize := 0
			// if there is a keychain, there is a form to add a new service (and a button)
			if m.keychain != nil {
				newServiceFormSize = len(m.newServiceInputs) + 1
			}
			if m.focusIndex > len(m.inputs)+urlBlockSize+2+serviceSize+newServiceFormSize {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) + urlBlockSize + 2 + serviceSize + newServiceFormSize
			}
		default:
			// remove the highlighted service
			if msg.String() == "d" && m.focusIndex > len(m.inputs)+urlBlockSize+1 && m.focusIndex < len(m.inputs)+urlBlockSize+2+len(m.serviceNames) {
				selectedService := m.focusIndex - len(m.inputs) - urlBlockSize - 2
				m.showSpinnerDeleteService = true
				return m, func() tea.Msg {
					return SendRemoveService{removeServiceAndRefresh(&m, selectedService)}
				}
			}
			// set a default derivation path
			if m.focusIndex == len(m.inputs)+6+len(m.serviceNames)+1 {
				serviceName := m.newServiceInputs[0].Value()
				derivationPath := "m/650'/" + serviceName + msg.String()
				m.newServiceInputs[1].SetValue(derivationPath)
			}
		}
	default:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}

	// Handle character input
	cmds := m.updateInputs(msg)

	cmds = append(cmds, m.updateFocus(urlBlockSize)...)

	return m, tea.Batch(cmds...)
}

func addService(m *Model) Model {
	accessSeed, err := getAccessKey(*m)
	if err != nil {
		m.feedback = err.Error()
		return *m
	}
	addServiceToKeychain(m, accessSeed, m.inputs[0].Value(), m.newServiceInputs[0].Value(), m.newServiceInputs[1].Value())
	m.newServiceInputs[0].SetValue("")
	m.newServiceInputs[1].SetValue("")
	m.focusIndex++
	return accessKeychain(m)
}

func removeServiceAndRefresh(m *Model, selectedService int) Model {
	accessSeed, err := getAccessKey(*m)
	if err != nil {
		m.feedback = err.Error()
		return *m
	}
	removeServiceFromKeychain(m, accessSeed, m.inputs[0].Value(), m.serviceNames[selectedService])
	return accessKeychain(m)
}

func accessKeychain(m *Model) Model {
	accessSeed, err := getAccessKey(*m)
	if err != nil {
		m.feedback = err.Error()
		return *m
	}
	err = validateInput(m.inputs[0].Value(), accessSeed)
	if err != nil {
		m.feedback = err.Error()
		return *m
	}

	keychain, err := tuiutils.AccessKeychain(m.inputs[0].Value(), accessSeed)
	if err != nil {
		m.feedback = err.Error()
		return *m
	}
	m.keychain = keychain
	m.serviceNames = make([]string, 0, len(m.keychain.Services))
	for k := range m.keychain.Services {
		m.serviceNames = append(m.serviceNames, k)
	}
	sort.Strings(m.serviceNames)
	m.selectedService = 0
	return *m
}

func validateInput(endoint string, seed []byte) error {
	if endoint == "" {
		return errors.New("please select a node endpoint")
	}
	if len(seed) == 0 {
		return errors.New("please enter a seed")
	}
	return nil
}

func (m *Model) updateInputs(msg tea.Msg) []tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs)+len(m.newServiceInputs))

	for i := range m.inputs {
		if m.pvKeyBytes != nil && i == 1 {
			continue
		}
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	for i := range m.newServiceInputs {
		m.newServiceInputs[i], cmds[i+len(m.inputs)] = m.newServiceInputs[i].Update(msg)
	}
	return cmds
}

func (m *Model) updateFocus(urlBlockSize int) []tea.Cmd {

	cmds := make([]tea.Cmd, len(m.inputs)+len(m.newServiceInputs))
	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focusIndex-urlBlockSize {
			// Set focused state
			cmds[i] = m.inputs[i].Focus()
			continue
		}
		// Remove focused state
		m.inputs[i].Blur()
		m.inputs[i].PromptStyle = noStyle
		m.inputs[i].TextStyle = noStyle
	}

	for i := 0; i < len(m.newServiceInputs); i++ {
		index := len(m.inputs) + i
		if i == m.focusIndex-len(m.inputs)-urlBlockSize-3-len(m.serviceNames) {
			// Set focused state
			cmds[index] = m.newServiceInputs[i].Focus()
			continue
		}
		// Remove focused state
		m.newServiceInputs[i].Blur()
		m.newServiceInputs[i].PromptStyle = noStyle
		m.newServiceInputs[i].TextStyle = noStyle
	}

	return cmds
}

func createKeychain(m *Model) Model {
	accessSeed, err := getAccessKey(*m)
	if err != nil {
		m.feedback = err.Error()
		return *m
	}
	feedback, keychainSeed, keychainTransactionAddress, keychainAccessTransactionAddress, error := tuiutils.CreateKeychain(m.inputs[0].Value(), accessSeed)
	if error != nil {
		m.feedback = error.Error()
	} else {
		m.feedback = feedback
	}
	m.keychainSeed = keychainSeed
	m.keychainTransactionAddress = keychainTransactionAddress
	m.keychainAccessTransactionAddress = keychainAccessTransactionAddress
	return *m
}

func addServiceToKeychain(m *Model, accessSeed []byte, endpoint string, serviceName string, serviceDerivationPath string) {
	feedback, err := tuiutils.AddServiceToKeychain(accessSeed, endpoint, serviceName, serviceDerivationPath)
	if err != nil {
		m.feedback = err.Error()
	} else {
		m.feedback = feedback
	}
}

func removeServiceFromKeychain(m *Model, accessSeed []byte, endpoint string, serviceName string) {
	feedback, err := tuiutils.RemoveServiceFromKeychain(accessSeed, endpoint, serviceName)
	if err != nil {
		m.feedback = err.Error()
	} else {
		m.feedback = feedback
	}
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString("> Node endpoint:\n")
	b.WriteString(urlView(m))
	for i := range m.inputs {
		b.WriteRune('\n')
		b.WriteString(m.inputs[i].View())
	}

	if m.showSpinnerCreate {
		b.WriteString("\n\n")
		b.WriteString(m.Spinner.View())
	}

	createButton := &createBlurredButton
	if m.focusIndex == len(m.inputs)+len(urlType) {
		createButton = &createFocusedButton
	}

	if m.feedback != "" {
		b.WriteString("\n\n")
		b.WriteString(m.feedback)
		b.WriteString("\n\n")
	}

	fmt.Fprintf(&b, "\n\n%s", *createButton)
	b.WriteRune('\n')
	if m.keychainSeed != "" {
		fmt.Fprintf(&b, "Keychain seed: %s\n", m.keychainSeed)
		fmt.Fprintf(&b, "Keychain transaction: %s\n", m.keychainTransactionAddress)
		fmt.Fprintf(&b, "Keychain access transaction: %s\n", m.keychainAccessTransactionAddress)
	}

	if m.showSpinnerAccess {
		b.WriteString("\n\n")
		b.WriteString(m.Spinner.View())
	}
	button := &accessBlurredButton
	if m.focusIndex == len(m.inputs)+5 {
		button = &accessFocusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	if m.keychain != nil {
		var b2 strings.Builder
		b2.WriteString("--------------------\n")
		b2.WriteString("KEYCHAIN MANAGEMENT:\n")
		b2.WriteString("--------------------\n\n")
		b2.WriteString("List of services:\n")
		for i, k := range m.serviceNames {

			var u string
			if m.selectedService == i {
				u = "(•) "
			} else {
				u = "( ) "
			}

			keychainDerivedAddress, _ := m.keychain.DeriveAddress(k, 0)
			u += k + " : " + m.keychain.Services[k].DerivationPath + " (" + hex.EncodeToString(keychainDerivedAddress) + ")\n"
			if m.focusIndex == i+len(m.inputs)+6 {
				b2.WriteString(focusedStyle.Render(u))
			} else {
				b2.WriteString(u)
			}
			b2.WriteString("\n")
		}
		b2.WriteString(helpStyle.Render("press 'enter' to select or 'd' to delete "))

		if m.showSpinnerDeleteService {
			b2.WriteString("\n\n")
			b2.WriteString(m.Spinner.View())
		}

		if len(m.serviceNames) > 0 {
			button := &createTransactionaccessBlurredButton
			if m.focusIndex == len(m.inputs)+len(m.serviceNames)+6 {
				button = &createTransactionFocusedButton
			}
			fmt.Fprintf(&b2, "\n\n\n%s\n\n", *button)
		} else {
			b2.WriteString("No service")
		}

		b2.WriteString("Add a service:\n")
		// add fields for service name and derivation path
		for i := range m.newServiceInputs {
			b2.WriteRune('\n')
			b2.WriteString(m.newServiceInputs[i].View())
		}

		if m.showSpinnerCreateService {
			b2.WriteString("\n\n")
			b2.WriteString(m.Spinner.View())
		}

		createServiceButton := &createServiceBlurredButton
		if m.focusIndex == len(m.inputs)+len(m.serviceNames)+6+len(m.newServiceInputs)+1 {
			createServiceButton = &createServiceFocusedButton
		}
		fmt.Fprintf(&b2, "\n\n%s\n\n", *createServiceButton)

		dialogBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 10).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

		dialog := lipgloss.Place(150, 9,
			lipgloss.Left, lipgloss.Center,
			dialogBoxStyle.Render(b2.String()),
		)
		b.WriteString(dialog)
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("press 'esc' to go back "))

	return b.String()
}

func urlView(m Model) string {
	s := strings.Builder{}

	for i := 0; i < len(urlType); i++ {
		var u string
		if m.selectedUrl == urlType[i] {
			u = "(•) "
		} else {
			u = "( ) "
		}
		u += urlType[i]
		if i == m.focusIndex {
			s.WriteString(focusedStyle.Render(u))
		} else {
			s.WriteString(u)
		}
		s.WriteString("\n")
	}

	return s.String()
}

func getAccessKey(m Model) ([]byte, error) {
	potentialWordsList := strings.Fields(m.inputs[1].Value())
	if len(potentialWordsList) == 24 {
		seed, err := tuiutils.ExtractSeedFromMnemonic(m.inputs[1].Value())
		if err != nil {
			return nil, err
		}
		if seed != nil {
			return seed, nil
		}
	}

	accessSeed, err := archethic.MaybeConvertToHex(m.inputs[1].Value())
	if m.pvKeyBytes != nil {
		accessSeed = m.pvKeyBytes
	}
	return accessSeed, err
}
