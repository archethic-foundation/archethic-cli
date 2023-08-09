# Archethic Command Line Interpreter

This command line interface enables you to interact with the Archethic's blockchain within the terminal by providing snap features:
- Generate transaction address 
- Build & send transaction
- Manage keychain (decentralized wallet)
    - Create and connect to your keychain (decentralized wallet)
    - Build and send transaction from your connected keychain
    - Update your keychain by adding and removing services

![](intro.gif)

## Installation

Download a pre-compiled binary or package from the [releases](https://github.com/archethic-foundation/archethic-cli/releases) page.

Or install it with go:

```bash
go install github.com/archethic-foundation/archethic-cli@latest
```

Or just build it yourself (requires Go 1.18+):

```bash
git clone https://github.com/archethic-foundation/archethic-cli.git
cd archethic-cli
go build .
```

On UNIX system, please note that you would need to install a clipboard utility if you want to be able to paste text/code (like xsel, xclip, wl-clipboard or Termux).

## Usage

By default the CLI works as TUI (terminal user interface) application allowing the application to be interactive. 

### TUI
To launch the archetic-cli with a TUI (terminal user interface), you need to call the executable without any flag. You could additionnally pass the `--ssh` flag (to use `~/.ssh/id_ed25519` or `~/.ssh/id_rsa`) or you can pass `--ssh-path` (with the location of your ssh key file). If a passphrase is needed, a prompt will appear to enter it), and the ssh key will be used as a seed.
When launching the Archethic TUI you will access to the main menu that allows you to select an action. 

- Generate an address
- Build and send a transaction
    - send uco
    - send tokens
    - interact with smart contract (recipients)
    - add ownerships and secret delegation
    - add abritraty content 
    - add smart contract's code
- Manage keychains
    - create a keychain with a given seed
    - access a keychain
    - add and remove services from a keychain
    - send a keychain transaction for a specific service

### CLI
It is also possible to call the archethic cli tool using the command line.

#### Generate address
`generate-address`Get the address of a transaction based on parameters

Arguments:
- `--seed`  (string) the seed
- `--ssh` (bool) enables ssh option for the seed. If the `--ssh-path` flag is not set, it tries to open the default key files: first `~/.ssh/id_ed25519` and if it doesn't exist, then it tries `~/.ssh/id_rsa`. If `--ssh-path` is passed, then provided value is used. If a passphrase is needed, a prompt will appear to enter it. Can't be set if `seed` is set.
- `--ssh-path` (string) path to ssh key to generate a seed, if a passphrase is needed, a prompt will appear to enter it. Can't be set if `--seed` is set.
- `--index` (integer) index of the transaction
- `--hash-algorithm`  (SHA256|SHA512|SHA3_256|SHA3_512|BLAKE2B) the hash algorithm. Default value is `SHA256`
- `--elliptic-curve` (ED25519|P256|SECP256K1) the elliptic curve. The default value is `ED25519`

#### Send transaction
`send-transaction`
Send a transaction. It’s also possible to send a transaction for a specific service of a keychain (by passing the service parameter). You can either pass the parameter as flags in the command line or decide to put the parameters in a YAML file and pass the path of the config file using the `config` flag.

Arguments:
- `--config` (string) the path of the yaml configuration file (see below for the explanation of the parameters). It is possible to use a combination of configuration with a file and flags (the flags are described below). But if a given value is defined both in the file and a flag, the value from the file will be ignored.
- `--endpoint`  (local|testnet|mainnet|[custom url]) the endpoint to use, you can write your own URL. Default value is `local`.
- `--access-seed`(string) the access seed. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
- `--ssh` (bool) enables ssh option for the seed. If the `--ssh-path` flag is not set, it tries to open the default key files: first `~/.ssh/id_ed25519` and if it doesn't exist, then it tries `~/.ssh/id_rsa`. If `--ssh-path` is passed, then provided value is used. If a passphrase is needed, a prompt will appear to enter it. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`..
- `--ssh-path` (string) path to ssh key to generate a seed, if a passphrase is needed, a prompt will appear to enter it. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`..
- `mnemonic` (boolean) enable use of mnemonic (BIP39) for seed. If set a prompt asking for the list of words is displayed (default is false). You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
- `--index` (integer) the index of the new transaction. The default value is the last transaction index (which is fetched).
- `--elliptic-curve` (ED25519|P256|SECP256K1) the elliptic curve. The default value is `ED25519`
- `transaction-type`  (keychain_access|keychain|transfer|hosting|token|data|contract|code_proposal|code_approval) the transaction type. The default value is `transfer`.
- `--uco-transfer` (destinationAddress(string)=amount(float)) the UCO transfers. You can create several UCO transfers in a transaction by passing the `uco-transfer` flag several times. The amount passed will be multiplied by 10^8.
- `--token-transfer`  (to(string)=amount(float),token_address(string),token_id(integer)) the token transfers. You can create several token transfers in a transaction by passing the `token-transfer` flag several times. The amount passed will be multiplied by 10^8.
- `--recipients` (string) the recipients. You can create several recipients in a transaction by passing the `recipients` flag several times. 
- `--ownerships` (secret(string)=authorization_key(string)) the ownerships. You can create several ownerships in a transaction by passing the `ownerships` flag several times. In the sent transaction, the ownerships will be grouped by `secret`.
- `--content` (string) the path of the file containing the `content` of the transaction.
- `--smart-contract` (string) the path of the file containing the `smart-contract` of the transaction.
- `--serviceName` (string) the name of the service of the keychain. You want to use to create the transaction

YAML configuration file:

The parameters you can use in the YAML file are basically the same, only the format of the `ownerships`, `recipients`, `token-transfers` and `uco-transfers` changes. Also, regarding the `smart-contract`
and the `content` parameters, it is the value of the param that is passed (not the path of the file containing the value).
```yaml
endpoint: local
access_seed: testtest
index: 7
elliptic_curve: ED25519
transaction_type: contract
uco_transfers:
  - to: 0000D574D171A484F8DEAC2D61FC3F7CC984BEB52465D69B3B5F670090742CBF5CCA
    amount: 1
  - to: 0000D574D171A484F8DEAC2D61FC3F7CC984BEB52465D69B3B5F670090742CBF5CCA
    amount: 2
token_transfers:
  - to: 0000D574D171A484F8DEAC2D61FC3F7CC984BEB52465D69B3B5F670090742CBF5CCA
    amount: 1
    token_id: 1
    token_address: 0000D574D171A484F8DEAC2D61FC3F7CC984BEB52465D69B3B5F670090742CBF5CCA
smart_contract: |
  condition inherit: [
    type: transfer,
      uco_transfers: %{
        "0000D574D171A484F8DEAC2D61FC3F7CC984BEB52465D69B3B5F670090742CBF5CCA" => 100000000
      }
  ]
  actions triggered_by: interval, at: "0 0 1 * *" do
    set_type transfer
    add_uco_transfer to: "0000D574D171A484F8DEAC2D61FC3F7CC984BEB52465D69B3B5F670090742CBF5CCA", amount: 100000000
  end
ownerships:
  - secret: testtest
    authorized_keys:
      - 000150D4592BD0AC74BA6B5BAC49E505FB878F14DEED1692E5017ABFEFE49D060B6E
```

#### Get transaction fee
`get-transaction-fee`
Gets the transaction fee, in the following format `{"Fee":16617375,"Rates":{"Eur":0.05518,"Usd":0.0602}}`.
The flags are the same as those used for the `send-transaction` command.


#### Create keychain
`create-keychain` creates a new keychain

Arguments:
- `--endpoint`  (local|testnet|mainnet|[custom url]) the endpoint to use, you can write your own URL. Default value is `local`.
- `--access-seed`(string) the access seed of the keychain. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
- `--ssh` (bool) enables ssh option for the seed. If the `--ssh-path` flag is not set, it tries to open the default key files: first `~/.ssh/id_ed25519` and if it doesn't exist, then it tries `~/.ssh/id_rsa`. If `--ssh-path` is passed, then provided value is used. If a passphrase is needed, a prompt will appear to enter it. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
- `--ssh-path` (string) path to ssh key to generate a seed, if a passphrase is needed, a prompt will appear to enter it. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
- `mnemonic` (boolean) enable use of mnemonic (BIP39) for seed. If set a prompt asking for the list of words is displayed (default is false). You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.

#### Get keychain
`get-keychain` access the details of the keychain (list of services)

Arguments:
- `--endpoint`  (local|testnet|mainnet|[custom url]) the endpoint to use, you can write your own URL. Default value is `local`.
- `--access-seed`(string) the access seed of the keychain. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
- `--ssh` (bool) enables ssh option for the seed. If the `--ssh-path` flag is not set, it tries to open the default key files: first `~/.ssh/id_ed25519` and if it doesn't exist, then it tries `~/.ssh/id_rsa`. If `--ssh-path` is passed, then provided value is used. If a passphrase is needed, a prompt will appear to enter it. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
- `--ssh-path` (string) path to ssh key to generate a seed, if a passphrase is needed, a prompt will appear to enter it. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
- `mnemonic` (boolean) enable use of mnemonic (BIP39) for seed. If set a prompt asking for the list of words is displayed (default is false). You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.

#### Add service to keychain
`add-service-to-keychain` add a service to a keychain

Arguments:
- `--endpoint`  (local|testnet|mainnet|[custom url]) the endpoint to use, you can write your own URL. Default value is `local`.
- `--access-seed`(string) the access seed of the keychain. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
- `--service-name` (string) the name of the service to add
- `--derivation-path` (string) the derivation path of the service to add
- `--ssh` (bool) enables ssh option for the seed. If the `--ssh-path` flag is not set, it tries to open the default key files: first `~/.ssh/id_ed25519` and if it doesn't exist, then it tries `~/.ssh/id_rsa`. If `--ssh-path` is passed, then provided value is used. If a passphrase is needed, a prompt will appear to enter it. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
- `--ssh-path` (string) path to ssh key to generate a seed, if a passphrase is needed, a prompt will appear to enter it. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
- `mnemonic` (boolean) enable use of mnemonic (BIP39) for seed. If set a prompt asking for the list of words is displayed (default is false). You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.

#### Delete service from keychain
`delete-service-from-keychain` delete a service from a keychain

Arguments:
- `--endpoint`  (local|testnet|mainnet|[custom url]) the endpoint to use, you can write your own URL. Default value is `local`.
- `--access-seed`(string) the access seed of the keychain. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
- `--service-name` (string) the name of the service to delete
- `--ssh` (bool) enables ssh option for the seed. If the `--ssh-path` flag is not set, it tries to open the default key files: first `~/.ssh/id_ed25519` and if it doesn't exist, then it tries `~/.ssh/id_rsa`. If `--ssh-path` is passed, then provided value is used. If a passphrase is needed, a prompt will appear to enter it. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
- `--ssh-path` (string) path to ssh key to generate a seed, if a passphrase is needed, a prompt will appear to enter it. You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
- `mnemonic` (boolean) enable use of mnemonic (BIP39) for seed. If set a prompt asking for the list of words is displayed (default is false). You can only pass either `--access-seed`, or a combination of `--ssh`/`--ssh-path` or `--mnemonic`.
  
## License
[AGPL-3](/LICENCE)
