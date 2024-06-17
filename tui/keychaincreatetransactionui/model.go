package keychaincreatetransactionui

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	archethic "github.com/archethic-foundation/libgo"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BackMsg bool

type CreateTransactionType int

type CreateTransactionMsg struct {
	ServiceName string
	Seed        string
	Url         string
	PvKeyBytes  []byte
}

type TransactionSent struct {
	Error error
	Model Model
}
type TransactionFeeSent struct {
	Error error
	Model Model
}

var (
	focusedStyle                   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle                   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle                    = focusedStyle.Copy()
	noStyle                        = lipgloss.NewStyle()
	helpStyle                      = blurredStyle.Copy()
	focusedAddAuthKey              = focusedStyle.Copy().Render("[ Add authorization key ]")
	blurredAddAuthKey              = fmt.Sprintf("[ %s ]", blurredStyle.Render("Add authorization key"))
	focusedLoadStorageNouncePK     = focusedStyle.Copy().Render("[ Load Storage Nounce Public Key ]")
	blurredLoadStorageNouncePK     = fmt.Sprintf("[ %s ]", blurredStyle.Render("Load Storage Nounce Public Key"))
	focusedButton                  = focusedStyle.Copy().Render("[ Add ]")
	blurredButton                  = fmt.Sprintf("[ %s ]", blurredStyle.Render("Add"))
	focusedResetButton             = focusedStyle.Copy().Render("[ Reset ]")
	blurredResetButton             = fmt.Sprintf("[ %s ]", blurredStyle.Render("Reset"))
	focusedGetTransactionFeeButton = focusedStyle.Copy().Render("[ Get Transaction Fee ]")
	blurredGetTransactionFeeButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Get Transaction Fee"))
)

type RenderFunc func(m Model) string
type createTransactionTab int

const (
	MAIN_TAB           createTransactionTab = 0
	UCO_TAB            createTransactionTab = 1
	TOKEN_TAB          createTransactionTab = 2
	RECIPIENTS_TAB     createTransactionTab = 3
	OWNERSHIPS_TAB     createTransactionTab = 4
	CONTENT_TAB        createTransactionTab = 5
	SMART_CONTRACT_TAB createTransactionTab = 6
)

const (
	MAIN_ADD_BUTTON_INDEX                 = 17
	MAIN_GET_TRANSACTION_FEE_BUTTON_INDEX = 18
	MAIN_RESET_BUTTON_INDEX               = 19
	FIRST_TRANSACTION_TYPE_INDEX          = 8
	URL_INDEX                             = 4
	SEED_INDEX                            = 5
	CURVE_INDEX                           = 6
	TRANSACTION_INDEX_FIELD_INDEX         = 7
)

type SwitchTab struct{}
type Model struct {
	Tabs                   []string
	activeTab              createTransactionTab
	mainModel              MainModel
	ucoTransferModel       UcoTransferModel
	tokenTransferModel     TokenTransferModel
	recipientsModel        RecipientsModel
	ownershipsModel        OwnershipsModel
	contentModel           ContentModel
	smartContractModel     SmartContractModel
	transaction            archethic.TransactionBuilder
	secretKey              []byte
	storageNouncePublicKey string
	serviceName            string
	serviceMode            bool
	feedback               string
	url                    string
	seed                   string
	transactionIndex       uint
	showSpinner            bool
	Spinner                spinner.Model
	IsInit                 bool
	pvKeyBytes             []byte
}

func New(pvKeyBytes []byte) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	key := make([]byte, 32)
	rand.Read(key)
	m := Model{
		activeTab:   MAIN_TAB,
		transaction: *archethic.NewTransaction(archethic.KeychainAccessType),
		secretKey:   key,
		Spinner:     s,
		pvKeyBytes:  pvKeyBytes,
	}

	m.Tabs = []string{"Main", "UCO Transfers", "Token Transfers", "Recipients", "Ownerships", "Content", "Smart Contract"}
	m.resetInterface(pvKeyBytes)
	return m
}

func (m *Model) resetInterface(pvKeyBytes []byte) {
	m.transaction = *archethic.NewTransaction(archethic.KeychainAccessType)
	m.mainModel = NewMainModel(pvKeyBytes)
	m.ucoTransferModel = NewUcoTransferModel(&m.transaction)
	m.tokenTransferModel = NewTokenTransferModel(&m.transaction)
	m.recipientsModel = NewRecipientsModel(&m.transaction)
	m.ownershipsModel = NewOwnershipsModel(m.secretKey, &m.transaction)
	m.contentModel = NewContentModel()
	m.smartContractModel = NewSmartContractModel()
	if m.serviceMode {
		w, _ := m.mainModel.Update(CreateTransactionMsg{
			ServiceName: m.serviceName,
			Url:         m.url,
			Seed:        m.seed,
		})
		m.mainModel = w.(MainModel)
		m.ownershipsModel.SetUrl(m.url)
	}
}

func numberValidator(s string) error {
	if s == "" {
		return nil
	}
	_, err := strconv.ParseFloat(s, 64)
	return err
}

func curveValidator(s string) error {
	val, err := strconv.ParseInt(s, 10, 32)
	if err == nil && (val < 0 || val > 2) {
		return errors.New("number should be >0 and <=2")
	}
	return err
}

func (m Model) Init() tea.Cmd {
	return m.Spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case CreateTransactionMsg:
		m.seed = msg.Seed
		m.serviceName = msg.ServiceName
		m.serviceMode = m.serviceName != ""
		m.url = msg.Url
		w, cmds := m.mainModel.Update(msg)
		m.mainModel = w.(MainModel)
		m.ownershipsModel.SetUrl(m.url)
		return m, cmds
	case UpdateTransactionIndex:
		m.transactionIndex = msg.Index
		cmds = msg.cmds
	case UpdateUrl:
		m.url = msg.Url
		m.storageNouncePublicKey = ""
		m.ownershipsModel.SetUrl(msg.Url)
		cmds = msg.cmds
	case UpdateTransactionType:
		m.transaction.SetType(msg.TransactionType)
	case TransactionSent:
		m.showSpinner = false
		if msg.Error != nil {
			m.feedback = msg.Error.Error()
		} else {
			m.feedback = msg.Model.feedback
		}
		return m, nil
	case TransactionFeeSent:
		m.showSpinner = false
		if msg.Error != nil {
			m.feedback = msg.Error.Error()
		} else {
			m.feedback = msg.Model.feedback
		}
		return m, nil
	case SendTransaction:
		m.showSpinner = true
		return m, func() tea.Msg {
			return sendTransaction(&m, msg.Curve, msg.Seed)
		}
	case GetTransactionFee:
		m.showSpinner = true
		return m, func() tea.Msg {
			return getTransactionFee(&m, msg.Curve, msg.Seed)
		}
	case ResetInterface:
		m.resetInterface(m.pvKeyBytes)
	case AddUcoTransfer:
		m.transaction.AddUcoTransfer(msg.To, msg.Amount)
		m.ucoTransferModel.transaction = &m.transaction
		cmds = msg.cmds
	case AddTokenTransfer:
		m.transaction.AddTokenTransfer(msg.To, msg.TokenAddress, msg.Amount, msg.TokenId)
		m.tokenTransferModel.transaction = &m.transaction
		cmds = msg.cmds
	case AddRecipient:
		m.feedback = ""
		if msg.Action == "" && msg.ArgsJson == "" {
			m.transaction.AddRecipient(msg.Address)
		} else {
			var args []interface{}
			err := json.Unmarshal([]byte(msg.ArgsJson), &args)
			if err == nil {
				m.transaction.AddRecipientWithNamedAction(msg.Address, []byte(msg.Action), args)
			} else {
				m.feedback = "Invalid JSON"
			}
		}
		m.recipientsModel.transaction = &m.transaction
		cmds = msg.cmds
	case AddOwnership:
		m.transaction.AddOwnership(msg.Cipher, msg.AuthorizedKeys)
		m.ownershipsModel.transaction = &m.transaction
		cmds = msg.cmds
	case UpdateStorageNouncePublicKey:
		m.storageNouncePublicKey = msg.StorageNouncePublicKey
		// need to send back the message to ownerships model to update the spinner
		w, _ := m.ownershipsModel.Update(msg)
		m.ownershipsModel = w.(OwnershipsModel)
		return m, nil
	case DeleteUcoTransfer:
		m.transaction.Data.Ledger.Uco.Transfers = append(m.transaction.Data.Ledger.Uco.Transfers[:msg.IndexToDelete], m.transaction.Data.Ledger.Uco.Transfers[msg.IndexToDelete+1:]...)
		m.ucoTransferModel.transaction = &m.transaction
		return m, nil
	case DeleteTokenTransfer:
		m.transaction.Data.Ledger.Token.Transfers = append(m.transaction.Data.Ledger.Token.Transfers[:msg.IndexToDelete], m.transaction.Data.Ledger.Token.Transfers[msg.IndexToDelete+1:]...)
		m.tokenTransferModel.transaction = &m.transaction
		return m, nil
	case DeleteRecipient:
		m.transaction.Data.Recipients = append(m.transaction.Data.Recipients[:msg.IndexToDelete], m.transaction.Data.Recipients[msg.IndexToDelete+1:]...)
		m.recipientsModel.transaction = &m.transaction
		return m, nil
	case DeleteOwnership:
		m.transaction.Data.Ownerships = append(m.transaction.Data.Ownerships[:msg.IndexToDelete], m.transaction.Data.Ownerships[msg.IndexToDelete+1:]...)
		m.ownershipsModel.transaction = &m.transaction
		return m, nil
	case UpdateContent:
		m.transaction.SetContent(msg.Content)
	case UpdateSmartContract:
		m.transaction.SetCode(msg.Code)
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "esc":
			if m.activeTab == SMART_CONTRACT_TAB && m.smartContractModel.smartContractTextAreaInput.Focused() {
				w, cmds := m.smartContractModel.Update(msg)
				m.smartContractModel = w.(SmartContractModel)
				return m, cmds
			} else if m.activeTab == CONTENT_TAB && m.contentModel.contentTextAreaInput.Focused() {
				w, cmds := m.contentModel.Update(msg)
				m.contentModel = w.(ContentModel)
				return m, cmds
			} else {
				return New(m.pvKeyBytes), func() tea.Msg {
					return BackMsg(true)
				}
			}
		case "ctrl+c":
			return m, tea.Quit
		case "right", "tab":
			// switch to the next tab except if the user is editing the content or the smart contract
			if (m.activeTab == CONTENT_TAB && !m.contentModel.contentTextAreaInput.Focused()) ||
				(m.activeTab == SMART_CONTRACT_TAB && !m.smartContractModel.smartContractTextAreaInput.Focused()) ||
				(m.activeTab != CONTENT_TAB && m.activeTab != SMART_CONTRACT_TAB) {
				m.activeTab = getNewTab(&m, int(m.activeTab)+1)
				cmds = focusOnTab(&m)
			} else if m.activeTab == CONTENT_TAB {
				w, cmds := m.contentModel.Update(msg)
				m.contentModel = w.(ContentModel)
				return m, cmds
			} else if m.activeTab == SMART_CONTRACT_TAB {
				w, cmds := m.smartContractModel.Update(msg)
				m.smartContractModel = w.(SmartContractModel)
				return m, cmds
			}

		case "left", "shift+tab":
			// switch to the previous tab except if the user is editing the content or the smart contract
			if (m.activeTab == CONTENT_TAB && !m.contentModel.contentTextAreaInput.Focused()) ||
				(m.activeTab == SMART_CONTRACT_TAB && !m.smartContractModel.smartContractTextAreaInput.Focused()) ||
				(m.activeTab != CONTENT_TAB && m.activeTab != SMART_CONTRACT_TAB) {
				m.activeTab = getNewTab(&m, int(m.activeTab)-1)
				cmds = focusOnTab(&m)
			} else if m.activeTab == CONTENT_TAB {
				w, cmds := m.contentModel.Update(msg)
				m.contentModel = w.(ContentModel)
				return m, cmds
			} else if m.activeTab == SMART_CONTRACT_TAB {
				w, cmds := m.smartContractModel.Update(msg)
				m.smartContractModel = w.(SmartContractModel)
				return m, cmds
			}
		default:
			switch m.activeTab {
			case MAIN_TAB:
				w, cmds := m.mainModel.Update(msg)
				m.mainModel = w.(MainModel)
				return m, cmds
			case UCO_TAB:
				w, cmds := m.ucoTransferModel.Update(msg)
				m.ucoTransferModel = w.(UcoTransferModel)
				return m, cmds
			case TOKEN_TAB:
				w, cmds := m.tokenTransferModel.Update(msg)
				m.tokenTransferModel = w.(TokenTransferModel)
				return m, cmds
			case RECIPIENTS_TAB:
				w, cmds := m.recipientsModel.Update(msg)
				m.recipientsModel = w.(RecipientsModel)
				return m, cmds
			case OWNERSHIPS_TAB:
				w, cmds := m.ownershipsModel.Update(msg)
				m.ownershipsModel = w.(OwnershipsModel)
				return m, cmds
			case CONTENT_TAB:
				w, cmds := m.contentModel.Update(msg)
				m.contentModel = w.(ContentModel)
				return m, cmds
			case SMART_CONTRACT_TAB:
				w, cmds := m.smartContractModel.Update(msg)
				m.smartContractModel = w.(SmartContractModel)
				return m, cmds
			}
		}
	default:
		var cmds []tea.Cmd
		spinner, newCmd := m.Spinner.Update(msg)
		m.Spinner = spinner
		cmds = append(cmds, newCmd)
		// also update the spinner from the ownerships tab
		spinnerOwnership, cmd2 := m.ownershipsModel.Spinner.Update(msg)
		m.ownershipsModel.Spinner = spinnerOwnership
		cmds = append(cmds, cmd2)
		return m, tea.Batch(cmds...)
	}
	return m, tea.Batch(cmds...)
}

func getNewTab(m *Model, tabNb int) createTransactionTab {
	switch {
	case tabNb > len(m.Tabs)-1:
		return 0
	case tabNb < 0:
		return createTransactionTab(len(m.Tabs) - 1)
	default:
		return createTransactionTab(tabNb)
	}
}
func focusOnTab(m *Model) []tea.Cmd {
	var cmds []tea.Cmd
	switch m.activeTab {
	case MAIN_TAB:
		m.mainModel, cmds = m.mainModel.SwitchTab()
	case UCO_TAB:
		m.ucoTransferModel, cmds = m.ucoTransferModel.SwitchTab()
	case TOKEN_TAB:
		m.tokenTransferModel, cmds = m.tokenTransferModel.SwitchTab()
	case RECIPIENTS_TAB:
		m.recipientsModel, cmds = m.recipientsModel.SwitchTab()
	case OWNERSHIPS_TAB:
		if !m.ownershipsModel.IsInit {
			cmds = append(cmds, m.ownershipsModel.Init())
			m.ownershipsModel.IsInit = true
		}
		ownershipsModel, switchTabCmd := m.ownershipsModel.SwitchTab()
		m.ownershipsModel = ownershipsModel
		cmds = append(cmds, switchTabCmd...)
	case SMART_CONTRACT_TAB:
		m.smartContractModel, cmds = m.smartContractModel.SwitchTab()
	case CONTENT_TAB:
		m.contentModel, cmds = m.contentModel.SwitchTab()
	}
	return cmds
}
func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 5)
	activeTabStyle    = inactiveTabStyle.Copy().Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Left).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

func (m Model) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == int(m.activeTab)
		if isActive {
			style = activeTabStyle.Copy()
		} else {
			style = inactiveTabStyle.Copy()
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")

	tabContent := ""
	var b strings.Builder

	switch m.activeTab {
	case MAIN_TAB:
		b.WriteString(m.mainModel.View())
		b.WriteString("\n\n")
		if m.showSpinner {
			b.WriteString("\n\n")
			b.WriteString(m.Spinner.View())
		}
		b.WriteString(m.feedback)
	case UCO_TAB:
		b.WriteString(m.ucoTransferModel.View())
	case TOKEN_TAB:
		b.WriteString(m.tokenTransferModel.View())
	case RECIPIENTS_TAB:
		b.WriteString(m.recipientsModel.View())
	case OWNERSHIPS_TAB:
		b.WriteString(m.ownershipsModel.View())
	case CONTENT_TAB:
		b.WriteString(m.contentModel.View())
	case SMART_CONTRACT_TAB:
		b.WriteString(m.smartContractModel.View())
	}
	b.WriteString("\n\n")
	tabContent = b.String()
	doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(tabContent))
	doc.WriteString("\n\n")
	doc.WriteString(helpStyle.Render("press 'esc' to go back "))
	return docStyle.Render(doc.String())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func sendTransaction(m *Model, curve archethic.Curve, seed []byte) TransactionSent {
	m.feedback = ""
	feedback, error := tuiutils.SendTransaction(&m.transaction, m.secretKey, curve, m.serviceMode, m.url, m.transactionIndex, m.serviceName, m.storageNouncePublicKey, seed)
	m.feedback = fmt.Sprintf("Transaction sent: %s", feedback)
	if error != nil {
		return TransactionSent{Model: *m, Error: error}
	}
	return TransactionSent{Model: *m, Error: nil}
}

func getTransactionFee(m *Model, curve archethic.Curve, seed []byte) TransactionFeeSent {
	m.feedback = ""
	fee, error := tuiutils.GetTransactionFee(&m.transaction, m.secretKey, curve, m.serviceMode, m.url, m.transactionIndex, m.serviceName, m.storageNouncePublicKey, seed)
	humanReadableFee, _ := strconv.ParseFloat(archethic.FormatBigInt(fee.Fee, 8), 64)
	usdEquivalent := humanReadableFee * float64(fee.Rates.Usd)
	eurEquivanlent := humanReadableFee * float64(fee.Rates.Eur)
	m.feedback = fmt.Sprintf("Transaction fee: %f UCO (~ $%f) (~ %f€)", humanReadableFee, usdEquivalent, eurEquivanlent)
	if error != nil {
		return TransactionFeeSent{Model: *m, Error: error}
	}
	return TransactionFeeSent{Model: *m, Error: nil}
}
