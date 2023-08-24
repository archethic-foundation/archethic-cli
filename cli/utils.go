package cli

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"

	archethic "github.com/archethic-foundation/libgo"
	"github.com/spf13/pflag"
)

var (
	hashAlgo        = SHA256
	ellipticCurve   = ED25519
	endpoint        = mainnet
	transactionType = TransferType
)

type SendTransactionData struct {
	Endpoint        string          `yaml:"endpoint"`
	AccessSeed      string          `yaml:"access_seed"`
	Index           int             `yaml:"index"`
	EllipticCurve   string          `yaml:"elliptic_curve"`
	TransactionType string          `yaml:"transaction_type"`
	UcoTransfers    []UCOTransfer   `yaml:"uco_transfers,omitempty"`
	TokenTransfers  []TokenTransfer `yaml:"token_transfers,omitempty"`
	Recipients      []Recipient     `yaml:"recipients,omitempty"`
	Ownerships      []Ownership     `yaml:"ownerships,omitempty"`
	Content         string          `yaml:"content,omitempty"`
	SmartContract   string          `yaml:"smart_contract,omitempty"`
	ServiceName     string          `yaml:"serviceName,omitempty"`
}

type UCOTransfer struct {
	To     string  `yaml:"to"`
	Amount float64 `yaml:"amount"`
}

type TokenTransfer struct {
	To           string  `yaml:"to"`
	Amount       float64 `yaml:"amount"`
	TokenAddress string  `yaml:"token_address"`
	TokenID      int     `yaml:"token_id"`
}

type Ownership struct {
	Secret         string   `yaml:"secret"`
	AuthorizedKeys []string `yaml:"authorized_keys"`
}

type Recipient struct {
	Address  string `yaml:"to"`
	Action   string `yaml:"action"`
	ArgsJson string `yaml:"args_json"`
}

type ConfiguredTransaction struct {
	accessSeed     []byte
	index          int
	ucoTransfers   []UCOTransfer
	tokenTransfers []TokenTransfer
	recipients     []Recipient
	ownerships     []Ownership
	content        []byte
	smartContract  string
	serviceName    string
}

type EndpointCLI string

const (
	local   EndpointCLI = "local"
	testnet EndpointCLI = "testnet"
	mainnet EndpointCLI = "mainnet"
)

func (e *EndpointCLI) String() string {
	switch *e {
	case local:
		return "http://localhost:4000"
	case testnet:
		return "https://testnet.archethic.net"
	case mainnet:
		return "https://mainnet.archethic.net"
	default:
		_, err := url.Parse(string(*e))
		if err != nil {
			return ""
		}
		return string(*e)
	}
}

func (e *EndpointCLI) Set(value string) error {
	switch value {
	case "local":
		*e = "local"
	case "testnet":
		*e = "testnet"
	case "mainnet":
		*e = "mainnet"
	default:
		_, err := url.Parse(string(*e))
		if err != nil {
			return errors.New("invalid endpoint value")
		}
		*e = EndpointCLI(value)
	}
	return nil
}

func (e *EndpointCLI) Type() string {
	return "EndpointCLI"
}

type HashAlgoCLI uint8

const (
	SHA256   HashAlgoCLI = 0
	SHA512   HashAlgoCLI = 1
	SHA3_256 HashAlgoCLI = 2
	SHA3_512 HashAlgoCLI = 3
	BLAKE2B  HashAlgoCLI = 4
)

func (ha *HashAlgoCLI) String() string {
	switch *ha {
	case SHA256:
		return "SHA256"
	case SHA512:
		return "SHA512"
	case SHA3_256:
		return "SHA3_256"
	case SHA3_512:
		return "SHA3_512"
	case BLAKE2B:
		return "BLAKE2B"
	default:
		return ""
	}
}

func (ha *HashAlgoCLI) Set(value string) error {
	switch value {
	case "SHA256":
		*ha = SHA256
	case "SHA512":
		*ha = SHA512
	case "SHA3_256":
		*ha = SHA3_256
	case "SHA3_512":
		*ha = SHA3_512
	case "BLAKE2B":
		*ha = BLAKE2B
	default:
		return errors.New("invalid HashAlgo value")
	}
	return nil
}

func (ha *HashAlgoCLI) GetHashAlgo() (archethic.HashAlgo, error) {
	switch *ha {
	case SHA256:
		return archethic.SHA256, nil
	case SHA512:
		return archethic.SHA512, nil
	case SHA3_256:
		return archethic.SHA3_256, nil
	case SHA3_512:
		return archethic.SHA3_512, nil
	case BLAKE2B:
		return archethic.BLAKE2B, nil
	default:
		return archethic.SHA256, errors.New("invalid HashAlgo value")
	}
}

func (ha *HashAlgoCLI) Type() string {
	return "HashAlgoCLI"
}

type CurveCLI uint8

const (
	ED25519   CurveCLI = 0
	P256      CurveCLI = 1
	SECP256K1 CurveCLI = 2
)

func (c *CurveCLI) String() string {
	switch *c {
	case ED25519:
		return "ED25519"
	case P256:
		return "P256"
	case SECP256K1:
		return "SECP256K1"
	default:
		return ""
	}
}

func (c *CurveCLI) Set(value string) error {
	switch value {
	case "ED25519":
		*c = ED25519
	case "P256":
		*c = P256
	case "SECP256K1":
		*c = SECP256K1
	default:
		return errors.New("invalid Curve value")
	}
	return nil
}

func (c *CurveCLI) GetCurve() (archethic.Curve, error) {
	switch *c {
	case ED25519:
		return archethic.ED25519, nil
	case P256:
		return archethic.P256, nil
	case SECP256K1:
		return archethic.SECP256K1, nil
	default:
		return archethic.ED25519, errors.New("invalid Curve value")
	}
}

func (c *CurveCLI) Type() string {
	return "CurveCLI"
}

type TransactionTypeCLI uint8

const (
	KeychainAccessType TransactionTypeCLI = 254
	KeychainType       TransactionTypeCLI = 255
	TransferType       TransactionTypeCLI = 253
	HostingType        TransactionTypeCLI = 252
	TokenType          TransactionTypeCLI = 251
	DataType           TransactionTypeCLI = 250
	ContractType       TransactionTypeCLI = 249
	CodeProposalType   TransactionTypeCLI = 5
	CodeApprovalType   TransactionTypeCLI = 6
)

func (tt *TransactionTypeCLI) String() string {
	switch *tt {
	case KeychainAccessType:
		return "keychain_access"
	case KeychainType:
		return "keychain"
	case TransferType:
		return "transfer"
	case HostingType:
		return "hosting"
	case TokenType:
		return "token"
	case DataType:
		return "data"
	case ContractType:
		return "contract"
	case CodeProposalType:
		return "code_proposal"
	case CodeApprovalType:
		return "code_approval"
	default:
		return ""
	}
}

func (tt *TransactionTypeCLI) Set(value string) error {
	switch value {
	case "keychain_access":
		*tt = KeychainAccessType
	case "keychain":
		*tt = KeychainType
	case "transfer":
		*tt = TransferType
	case "hosting":
		*tt = HostingType
	case "token":
		*tt = TokenType
	case "data":
		*tt = DataType
	case "contract":
		*tt = ContractType
	case "code_proposal":
		*tt = CodeProposalType
	case "code_approval":
		*tt = CodeApprovalType
	default:
		return errors.New("invalid TransactionType value")
	}
	return nil
}

func (tt *TransactionTypeCLI) Type() string {
	return "TransactionTypeCLI"
}

func (tt *TransactionTypeCLI) GetTransactionType() (archethic.TransactionType, error) {
	switch *tt {
	case KeychainAccessType:
		return archethic.KeychainAccessType, nil
	case KeychainType:
		return archethic.KeychainType, nil
	case TransferType:
		return archethic.TransferType, nil
	case HostingType:
		return archethic.HostingType, nil
	case TokenType:
		return archethic.TokenType, nil
	case DataType:
		return archethic.DataType, nil
	case ContractType:
		return archethic.ContractType, nil
	case CodeProposalType:
		return archethic.CodeProposalType, nil
	case CodeApprovalType:
		return archethic.CodeApprovalType, nil
	default:
		return archethic.TransferType, errors.New("invalid TransactionType value")
	}
}

func validateRequiredFlags(flags *pflag.FlagSet, sshFlagKey, sshPathFlagKey, seedKey, mnemonicFlag string) error {
	// validate if sshFlagKey or sshPathFlagKey or seedKey is set
	if !flags.Changed(sshFlagKey) && !flags.Changed(sshPathFlagKey) && !flags.Changed(seedKey) && !flags.Changed(mnemonicFlag) {
		errorMessage := fmt.Sprintf("required flag(s) \"%s\" or \"%s\" or \"%s\" or \"%s\" not set", sshFlagKey, sshPathFlagKey, seedKey, mnemonicFlag)
		return errors.New(errorMessage)
	}
	return nil
}

func GetFirstSshKeyDefaultPath() string {
	home, _ := os.UserHomeDir()
	return home + "/.ssh/id_ed25519"
}

func ToBigInt(number float64, decimals int) uint64 {
	return uint64(number * math.Pow(10, float64(decimals)))
}

func FromBigInt(number uint64, decimals int) float64 {
	return float64(number) / math.Pow(10, float64(decimals))
}
