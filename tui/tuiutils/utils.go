package tuiutils

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"syscall"

	archethic "github.com/archethic-foundation/libgo"
	"github.com/spf13/pflag"
	"github.com/tyler-smith/go-bip39"
	"github.com/tyler-smith/go-bip39/wordlists"
	"github.com/ybbus/jsonrpc/v3"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v3"
)

func GetHashAlgorithmName(h archethic.HashAlgo) string {
	switch h {
	case archethic.SHA256:
		return "SHA256"
	case archethic.SHA512:
		return "SHA512"
	case archethic.SHA3_256:
		return "SHA3_256"
	case archethic.SHA3_512:
		return "SHA3_512"
	case archethic.BLAKE2B:
		return "BLAKE2B"
	}
	panic("Unknown hash algorithm")
}

func GetCurveName(h archethic.Curve) string {
	switch h {
	case archethic.ED25519:
		return "ED25519"
	case archethic.P256:
		return "P256"
	case archethic.SECP256K1:
		return "SECP256K1"
	}
	panic("Unknown curve")
}

func CreateKeychain(url string, accessSeed []byte) (string, string, string, string, error) {
	originPrivateKey, _ := hex.DecodeString("01019280BDB84B8F8AEDBA205FE3552689964A5626EE2C60AA10E3BF22A91A036009")

	publicKey, _, err := archethic.DeriveKeypair(accessSeed, 0, archethic.ED25519)
	if err != nil {
		return "", "", "", "", err
	}

	randomSeed := make([]byte, 32)
	rand.Read(randomSeed)

	keychain := archethic.NewKeychain(randomSeed)
	keychain.AddService("uco", "m/650'/0/0", archethic.ED25519, archethic.SHA256)
	keychain.AddAuthorizedPublicKey(publicKey)

	accessAddress, err := archethic.DeriveAddress(accessSeed, 1, archethic.ED25519, archethic.SHA256)
	if err != nil {
		return "", "", "", "", err
	}
	keychainAddress, err := archethic.DeriveAddress(randomSeed, 1, archethic.ED25519, archethic.SHA256)
	if err != nil {
		return "", "", "", "", err
	}

	keychainTx, err := archethic.NewKeychainTransaction(keychain, 0)
	if err != nil {
		return "", "", "", "", err
	}
	keychainTx.OriginSign(originPrivateKey)

	client := archethic.NewAPIClient(url)
	accessKeychain, _ := archethic.GetKeychain(accessSeed, *client)
	if accessKeychain != nil {
		err = errors.New("keychain access already exists")
		return "", "", "", "", err
	}

	var returnedError error
	feedback := ""
	keychainSeed := ""
	keychainTransactionAddress := ""
	keychainAccessTransactionAddress := ""
	ts := archethic.NewTransactionSender(client)
	ts.AddOnRequiredConfirmation(func(nbConf uint) {
		feedback += "\nKeychain's transaction confirmed."

		keychainSeed = hex.EncodeToString(randomSeed)
		keychainTransactionAddress = fmt.Sprintf("%s/explorer/transaction/%x", url, keychainAddress)

		accessTx, err := archethic.NewAccessTransaction(accessSeed, keychainAddress)
		if err != nil {
			returnedError = err
			feedback = err.Error()
		}
		accessTx.OriginSign(originPrivateKey)
		ts2 := archethic.NewTransactionSender(client)
		ts2.AddOnRequiredConfirmation(func(nbConf uint) {
			feedback += "\nKeychain access transaction confirmed."
			ts2.Unsubscribe("confirmation")
			keychainAccessTransactionAddress = fmt.Sprintf("%s/explorer/transaction/%x", url, accessAddress)
		})
		ts2.AddOnError(func(senderContext string, error archethic.ErrorDetails) {
			feedback += handleTransactionError(error).Error()
			ts.Unsubscribe("error")
		})
		ts2.SendTransaction(accessTx, 100, 60)
		ts.Unsubscribe("confirmation")
	})
	ts.AddOnError(func(senderContext string, error archethic.ErrorDetails) {
		returnedError = handleTransactionError(error)
		feedback += returnedError.Error()
		ts.Unsubscribe("error")
	})
	ts.SendTransaction(keychainTx, 100, 60)
	return feedback, keychainSeed, keychainTransactionAddress, keychainAccessTransactionAddress, returnedError
}

func AccessKeychain(endpoint string, seed []byte) (*archethic.Keychain, error) {
	client := archethic.NewAPIClient(endpoint)
	return archethic.GetKeychain(seed, *client)
}

func AddServiceToKeychain(accessSeed []byte, endpoint string, serviceName string, serviceDerivationPath string) (string, error) {
	return updateKeychain(accessSeed, endpoint, func(keychain *archethic.Keychain) {
		keychain.AddService(serviceName, serviceDerivationPath, archethic.ED25519, archethic.SHA256)
	})
}

func RemoveServiceFromKeychain(accessSeed []byte, endpoint string, serviceName string) (string, error) {
	return updateKeychain(accessSeed, endpoint, func(keychain *archethic.Keychain) {
		keychain.RemoveService(serviceName)
	})
}

func updateKeychain(accessSeed []byte, endpoint string, updateFunc func(*archethic.Keychain)) (string, error) {
	client := *archethic.NewAPIClient(endpoint)
	keychain, err := archethic.GetKeychain(accessSeed, client)
	if err != nil {
		return "", err
	}
	updateFunc(keychain)

	keychainGenesisAddress, err := archethic.DeriveAddress(keychain.Seed, 0, archethic.ED25519, archethic.SHA256)
	if err != nil {
		return "", err
	}
	addressHex := hex.EncodeToString(keychainGenesisAddress)
	transactionChainIndex := client.GetLastTransactionIndex(addressHex)
	transaction, err := archethic.NewKeychainTransaction(keychain, uint32(transactionChainIndex))
	if err != nil {
		return "", err
	}
	originPrivateKey, _ := hex.DecodeString("01019280BDB84B8F8AEDBA205FE3552689964A5626EE2C60AA10E3BF22A91A036009")
	transaction.OriginSign(originPrivateKey)

	var returnedError error
	var returnedFeedback = ""
	returnedError = nil

	ts := archethic.NewTransactionSender(&client)
	ts.AddOnRequiredConfirmation(func(nbConf uint) {
		returnedFeedback = "\nKeychain's transaction confirmed."
	})
	ts.AddOnError(func(senderContext string, error archethic.ErrorDetails) {
		returnedError = handleTransactionError(error)
		ts.Unsubscribe("error")
	})
	ts.SendTransaction(transaction, 100, 60)

	return returnedFeedback, returnedError
}

func SendTransaction(transaction *archethic.TransactionBuilder, secretKey []byte, curve archethic.Curve, serviceMode bool, endpoint string, transactionIndex uint, serviceName string, storageNouncePublicKey string, seed []byte) (string, error) {
	err := buildTransactionToSend(transaction, secretKey, curve, serviceMode, endpoint, transactionIndex, serviceName, storageNouncePublicKey, seed)
	if err != nil {
		return "", err
	}
	feedback := ""
	client := archethic.NewAPIClient(endpoint)
	ts := archethic.NewTransactionSender(client)
	ts.AddOnSent(func() {
		feedback = endpoint + "/explorer/transaction/" + strings.ToUpper(hex.EncodeToString(transaction.Address))
	})

	ts.AddOnError(func(sender string, error archethic.ErrorDetails) {
		feedback = handleTransactionError(error).Error()
	})

	ts.SendTransaction(transaction, 100, 60)
	return feedback, nil
}

func GetTransactionFeeJson(transaction *archethic.TransactionBuilder, secretKey []byte, curve archethic.Curve, serviceMode bool, endpoint string, transactionIndex uint, serviceName string, storageNouncePublicKey string, seed []byte) (string, error) {
	fee, err := GetTransactionFee(transaction, secretKey, curve, serviceMode, endpoint, transactionIndex, serviceName, storageNouncePublicKey, seed)
	if err != nil {
		return "", err
	}
	feeBytes, err := json.Marshal(fee)
	if err != nil {
		return "", err
	}
	return string(feeBytes), nil
}

func GetTransactionFee(transaction *archethic.TransactionBuilder, secretKey []byte, curve archethic.Curve, serviceMode bool, endpoint string, transactionIndex uint, serviceName string, storageNouncePublicKey string, seed []byte) (archethic.Fee, error) {
	err := buildTransactionToSend(transaction, secretKey, curve, serviceMode, endpoint, transactionIndex, serviceName, storageNouncePublicKey, seed)
	if err != nil {
		return archethic.Fee{}, err
	}
	client := archethic.NewAPIClient(endpoint)
	fee, err := client.GetTransactionFee(transaction)
	if err != nil {
		return archethic.Fee{}, handleTransactionError(err)
	}
	return fee, nil
}

func buildTransactionToSend(transaction *archethic.TransactionBuilder, secretKey []byte, curve archethic.Curve, serviceMode bool, endpoint string, transactionIndex uint, serviceName string, storageNouncePublicKey string, seed []byte) error {
	if len(transaction.Data.Code) > 0 {
		ownershipIndex := -1
		for i, ownership := range transaction.Data.Ownerships {
			decryptSecret, err := archethic.AesDecrypt(ownership.Secret, secretKey)
			if err != nil {
				return err
			}
			decodedSecret := string(decryptSecret)

			if reflect.DeepEqual(decodedSecret, string(seed)) {
				ownershipIndex = i
				break
			}
		}

		if ownershipIndex == -1 {
			return errors.New("you need to create an ownership with the transaction seed as secret and authorize node public key to let nodes generate new transaction from your smart contract")
		} else {
			authorizedKeyIndex := -1
			for i, authKey := range transaction.Data.Ownerships[ownershipIndex].AuthorizedKeys {
				if reflect.DeepEqual(strings.ToUpper(hex.EncodeToString(authKey.PublicKey)), storageNouncePublicKey) {
					authorizedKeyIndex = i
					break
				}
			}

			if authorizedKeyIndex == -1 {
				return errors.New("you need to create an ownership with the transaction seed as secret and authorize node public key to let nodes generate new transaction from your smart contract")
			}
		}
	}

	client := archethic.NewAPIClient(endpoint)

	if serviceMode {
		err := buildKeychainTransaction(seed, client, transaction, serviceName)
		if err != nil {
			return err
		}
	} else {
		err := transaction.Build(seed, uint32(transactionIndex), curve, archethic.SHA256)
		if err != nil {
			return err
		}
	}

	originPrivateKey, _ := hex.DecodeString("01019280BDB84B8F8AEDBA205FE3552689964A5626EE2C60AA10E3BF22A91A036009")
	transaction.OriginSign(originPrivateKey)
	return nil
}

func buildKeychainTransaction(seed []byte, client *archethic.APIClient, transaction *archethic.TransactionBuilder, serviceName string) error {
	keychain, err := archethic.GetKeychain(seed, *client)
	if err != nil {
		return err
	}

	genesisAddress, err := keychain.DeriveAddress(serviceName, 0)
	if err != nil {
		return err
	}

	index := client.GetLastTransactionIndex(hex.EncodeToString(genesisAddress))

	err = keychain.BuildTransaction(transaction, serviceName, uint8(index))
	if err != nil {
		return err
	}
	return nil
}

func GetSSHPrivateKey(privateKeyPath string) ([]byte, error) {
	var pvKeyBytes []byte
	// Read the private key file
	privateBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	pvKey, err := ssh.ParseRawPrivateKey(privateBytes)

	if _, ok := err.(*ssh.PassphraseMissingError); ok {
		passphrase := promptSecret(fmt.Sprintf("Enter passphrase for key '%s': ", privateKeyPath))
		pvKey, err = ssh.ParseRawPrivateKeyWithPassphrase(privateBytes, []byte(passphrase))
		if err != nil {
			return nil, errors.New("Failed to parse private key: " + err.Error())
		}
	}

	switch pvKey := pvKey.(type) {
	case *rsa.PrivateKey:
		pvKeyBytes = pvKey.D.Bytes()
	case *ecdsa.PrivateKey:
		pvKeyBytes = pvKey.D.Bytes()
	case *dsa.PrivateKey:
		pvKeyBytes = pvKey.X.Bytes()
	case *ed25519.PrivateKey:
		pvKeyBytes = *pvKey
	default:
		return nil, errors.New("Only RSA, ECDSA and DSA keys are supported, got " + reflect.TypeOf(pvKey).String())
	}
	return pvKeyBytes, nil

}

func promptSecret(message string) string {
	fmt.Printf(message)
	passphrase, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Failed to read secret: %v", err)
	}
	fmt.Println()
	return string(passphrase)
}

func GetLastTransactionIndex(url string, curve archethic.Curve, seed []byte) (uint, error) {
	client := archethic.NewAPIClient(url)
	address, err := archethic.DeriveAddress(seed, 0, curve, archethic.SHA256)
	if err != nil {
		return 0, err
	}
	addressHex := hex.EncodeToString(address)
	index := client.GetLastTransactionIndex(addressHex)
	return index, nil
}

func GetSeedBytes(flags *pflag.FlagSet, sshFlagKey, sshPathFlagKey, seedFlagKey, mnemonicFlag string) ([]byte, error) {
	// if the mnemonic flag is set, get the mnemonic words with a prompt
	if mnemonicFlag != "" {
		mnemonic, _ := flags.GetBool(mnemonicFlag)
		if mnemonic {
			words := promptSecret("Enter mnemonic words:")
			var err error
			accessSeedBytes, err := ExtractSeedFromMnemonic(words)
			if err != nil {
				return nil, err
			}
			return accessSeedBytes, nil
		}
	}
	// if the ssh flag is set, get the ssh key with a prompt
	ssh, _ := flags.GetBool(sshFlagKey)
	isSshPathSet := flags.Lookup(sshPathFlagKey).Changed
	sshEnabled := ssh || isSshPathSet
	if sshEnabled {
		// try to get the ssh key based on the provided path (or the default value)
		privateKeyPath, _ := flags.GetString(sshPathFlagKey)
		key, err := GetSSHPrivateKey(privateKeyPath)
		// if the path is provided but we get an error, return the error
		if flags.Changed(sshPathFlagKey) && err != nil {
			return nil, err
		}
		// but if the key was found, return it
		if key != nil {
			return key, nil
		}

		// otherwise try to get the second default value for ssh key path
		home, _ := os.UserHomeDir()
		key, err = GetSSHPrivateKey(home + "/.ssh/id_rsa")
		if err != nil {
			return nil, err
		}
		return key, nil
	}

	// if no seedFlagKey is provided, return nil
	if seedFlagKey == "" {
		return nil, nil
	}

	// otherwise try to get the seed from the seedFlagKey
	accessSeed, _ := flags.GetString(seedFlagKey)
	// check if the provided seed looks like a mnemonic word list
	potentialWordsList := strings.Fields(accessSeed)
	if len(potentialWordsList) == 24 {
		seed, err := ExtractSeedFromMnemonic(accessSeed)
		if err == nil && seed != nil {
			return seed, nil
		}
	}

	return archethic.MaybeConvertToHex(accessSeed)
}

func ExtractSeedFromMnemonic(words string) ([]byte, error) {
	// check if it's a bip39 word list in English
	if bip39.IsMnemonicValid(words) {
		seed, err := bip39.EntropyFromMnemonic(words)
		if err != nil {
			return nil, err
		}
		return seed, nil
	}
	// check if it's a bip39 word list in French
	bip39.SetWordList(wordlists.French)
	// normalize the string to NFD (as the bip39 French word list is in NFD)
	nfd := norm.NFD.String(words)
	if bip39.IsMnemonicValid(nfd) {
		seed, err := bip39.EntropyFromMnemonic(nfd)
		if err != nil {
			return nil, err
		}
		return seed, nil
	}
	return nil, nil
}

type customError struct {
	Error   string `yaml:"Error"`
	Message string `yaml:"Message,omitempty"`
	Data    any    `yaml:"Data,omitempty"`
}

func customErrorDetails(error archethic.ErrorDetails) string {
	var message string

	var throwData archethic.ErrorDetails

	errDataBytes, err := json.Marshal(error.Data)
	if err != nil {
		return err.Error()
	}
	if err := json.Unmarshal(errDataBytes, &throwData); err != nil {
		return err.Error()
	}

	var customErr customError
	if throwData.Message == "" {
		customErr = customError{
			Error: fmt.Sprintf("%s (%d)", error.Message, error.Code),
		}
	} else {
		message = fmt.Sprintf("%s (%d)", throwData.Message, throwData.Code)

		customErr = customError{
			Error:   fmt.Sprintf("%s (%d)", error.Message, error.Code),
			Message: message,
		}
		if message != "" && throwData.Data != nil {
			customDataDetails, _ := yaml.Marshal(throwData.Data)
			customErr.Data = fmt.Sprintf("%s", customDataDetails)
		}
	}

	b, _ := yaml.Marshal(customErr)

	return fmt.Sprintf("\n%s", b)
}

func handleTransactionError(err error) error {
	if errorDetails, ok := err.(archethic.ErrorDetails); ok {
		return errors.New(customErrorDetails(errorDetails))
	}

	if jsonRpcError, ok := err.(*jsonrpc.RPCError); ok {
		if mapError, ok := jsonRpcError.Data.(map[string]interface{}); ok {
			errorMessage := fmt.Sprintf("Error %d: %s %s.", jsonRpcError.Code, jsonRpcError.Message, flattenNestedMap(mapError, ""))
			return errors.New(errorMessage)
		}
	}

	return err
}

func flattenNestedMap(nestedMap map[string]interface{}, prefix string) string {
	var results string

	for key, value := range nestedMap {
		if innerMap, ok := value.(map[string]interface{}); ok {
			results += flattenNestedMap(innerMap, fmt.Sprintf("%s%s.", prefix, key))
		} else if arr, ok := value.([]interface{}); ok {
			for i, item := range arr {
				if innerMap, ok := item.(map[string]interface{}); ok {
					results += flattenNestedMap(innerMap, fmt.Sprintf("%s%s[%d].", prefix, key, i))
				} else {
					results += fmt.Sprintf("%s%s[%d] => %v\n", prefix, key, i, item)
				}
			}
		} else {
			results += fmt.Sprintf("%s%s => %v\n", prefix, key, value)
		}
	}

	return results
}
