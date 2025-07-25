package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Account struct {
	Address string `json:"address" yaml:"address"`
	Amount  int64  `json:"amount" yaml:"amount"`
}

type Validator struct {
	Address         string `json:"address,omitempty"`
	PublicKey       string `json:"publicKey,omitempty"`
	Committees      []int  `json:"committees"`
	NetAddress      string `json:"netAddress,omitempty"`
	StakedAmount    int64  `json:"stakedAmount,omitempty"`
	Output          string `json:"output,omitempty"`
	MaxPausedHeight int64  `json:"maxPausedHeight,omitempty"`
	UnstakingHeight int64  `json:"unstakingHeight,omitempty"`
	Delegate        bool   `json:"delegate,omitempty"`
	Compound        bool   `json:"compound,omitempty"`

	Profile     string `yaml:"profile" json:"-"`
	Key         int    `yaml:"key" json:"-"`
	ChainID     int    `yaml:"chainId" json:"-"`
	RootChainID int    `yaml:"rootChainId" json:"-"`
	Nested      bool   `yaml:"nested" json:"-"`
	EthOracle   bool   `yaml:"eth_oracle" json:"-"`
}

type Genesis struct {
	Time       string      `json:"time"`
	Accounts   []Account   `json:"accounts"`
	NonSigners interface{} `json:"nonSigners"`
	Validators []Validator `json:"validators"`
	Params     interface{} `json:"params"`
}

type KeyPair struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
	Address    string `json:"address"`
}

type KeyOutput struct {
	Timestamp string    `json:"timestamp"`
	Keys      []KeyPair `json:"keys"`
}

type Config struct {
	Accounts   []Account   `yaml:"accounts"`
	Validators []Validator `yaml:"validators"`
}

func getPortsForProfile(profile string, chainId int) (string, string, string, string, string, string) {
	listenPort := fmt.Sprintf("%d", 9000+chainId)
	
	switch profile {
	case "node-1":
		return "50000", "50001", "50002", "50003", listenPort, "127.0.0.101"
	case "node-2":
		return "40000", "40001", "40002", "40003", listenPort, "127.0.0.102"
	case "node-3":
		return "30000", "30001", "30002", "30003", listenPort, "127.0.0.103"
	default:
		panic("can't use default ports")
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <chain-profile-name>", os.Args[0])
	}
	chainProfileName := os.Args[1]

	genesisData, err := ioutil.ReadFile("templates/genesis.json")
	if err != nil {
		log.Fatalf("Error reading genesis.json: %v", err)
	}

	configTemplateData, err := ioutil.ReadFile("templates/config.json")
	if err != nil {
		log.Fatalf("Error reading templates/config.json: %v", err)
	}

	configPath := fmt.Sprintf("chain-profiles/%s.yaml", chainProfileName)
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading %s: %v", configPath, err)
	}

	keysData, err := ioutil.ReadFile("keys/node-bls.json")
	if err != nil {
		log.Fatalf("Error reading keys/node-bls.json: %v", err)
	}

	keystoreData, err := ioutil.ReadFile("keys/keystore.json")
	if err != nil {
		log.Fatalf("Error reading keys/keystore.json: %v", err)
	}

	var genesis Genesis
	if err := json.Unmarshal(genesisData, &genesis); err != nil {
		log.Fatalf("Error parsing genesis.json: %v", err)
	}

	var configTemplate map[string]interface{}
	if err := json.Unmarshal(configTemplateData, &configTemplate); err != nil {
		log.Fatalf("Error parsing templates/config.json: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Error parsing default.yaml: %v", err)
	}

	var keyOutput KeyOutput
	if err := json.Unmarshal(keysData, &keyOutput); err != nil {
		log.Fatalf("Error parsing keys/node-bls.json: %v", err)
	}

	genesis.Time = time.Now().Format("2006-01-02 15:04:05")

	// Create accounts from all BLS keys
	var accounts []Account
	for _, key := range keyOutput.Keys {
		account := Account{
			Address: key.Address,
			Amount:  1000000000,
		}
		accounts = append(accounts, account)
	}
	genesis.Accounts = accounts

	if len(config.Validators) > 0 {
		// Build all validators first
		mergedValidators := make([]Validator, len(config.Validators))
		for i, configValidator := range config.Validators {
			validator := Validator{
				Committees:      configValidator.Committees,
				NetAddress:      fmt.Sprintf("tcp://%s", configValidator.Profile),
				StakedAmount:    1000000000,
				MaxPausedHeight: 0,
				UnstakingHeight: 0,
				Delegate:        false,
				Compound:        true,
			}

			// Use the key field to reference the correct key from node-bls.json
			keyIndex := configValidator.Key
			if keyIndex >= 0 && keyIndex < len(keyOutput.Keys) {
				key := keyOutput.Keys[keyIndex]
				validator.Address = key.Address
				validator.PublicKey = key.PublicKey
				validator.Output = key.Address
			}

			mergedValidators[i] = validator
		}

		// Set all validators in the genesis
		genesis.Validators = mergedValidators

		// Generate files for each validator node
		for _, configValidator := range config.Validators {
			// Create directory structure
			dirName := fmt.Sprintf("%s-%s", chainProfileName, configValidator.Profile)
			dirPath := filepath.Join("data-dir", dirName)

			err := os.MkdirAll(dirPath, 0755)
			if err != nil {
				log.Fatalf("Error creating directory %s: %v", dirPath, err)
			}

			// Generate genesis.json (same for all nodes)
			genesisOutput, err := json.MarshalIndent(genesis, "", "  ")
			if err != nil {
				log.Fatalf("Error marshaling genesis output: %v", err)
			}

			genesisFilePath := filepath.Join(dirPath, "genesis.json")
			err = ioutil.WriteFile(genesisFilePath, genesisOutput, 0644)
			if err != nil {
				log.Fatalf("Error writing genesis.json to %s: %v", genesisFilePath, err)
			}

			// Generate config.json (unique for each node)
			nodeConfig := make(map[string]interface{})
			for k, v := range configTemplate {
				nodeConfig[k] = v
			}

			// Set node-specific ports and addresses
			walletPort, explorerPort, rpcPort, adminPort, listenPort, listenAddr := getPortsForProfile(configValidator.Profile, configValidator.ChainID)
			nodeConfig["walletPort"] = walletPort
			nodeConfig["explorerPort"] = explorerPort
			nodeConfig["rpcPort"] = rpcPort
			nodeConfig["adminPort"] = adminPort
			nodeConfig["listenAddress"] = fmt.Sprintf("%s:%s", listenAddr, listenPort)
			nodeConfig["externalAddress"] = configValidator.Profile
			nodeConfig["rpcURL"] = fmt.Sprintf("http://%s:%s", configValidator.Profile, rpcPort)
			nodeConfig["adminRPCUrl"] = fmt.Sprintf("http://%s:%s", configValidator.Profile, adminPort)

			// Set chainId from YAML configuration
			nodeConfig["chainId"] = configValidator.ChainID

			// Set runVDF based on nested flag
			if configValidator.Nested {
				nodeConfig["runVDF"] = false
			}

			// Add eth oracle configuration if enabled
			if configValidator.EthOracle {
				nodeConfig["ethBlockProviderConfig"] = map[string]interface{}{
					"ethNodeUrl":              "http://anvil:8545",
					"ethNodeWsUrl":            "ws://anvil:8545",
					"ethChainId":              1,
					"retryDelay":              5,
					"safeBlockConfirmations":  5,
				}
				nodeConfig["oracleConfig"] = map[string]interface{}{
					"stateSaveFile":       "last_block_height.txt",
					"orderResubmitDelay":  2,
					"committee":           2,
				}
			}

			configOutput, err := json.MarshalIndent(nodeConfig, "", "  ")
			if err != nil {
				log.Fatalf("Error marshaling config output: %v", err)
			}

			configFilePath := filepath.Join(dirPath, "config.json")
			err = ioutil.WriteFile(configFilePath, configOutput, 0644)
			if err != nil {
				log.Fatalf("Error writing config.json to %s: %v", configFilePath, err)
			}

			// Sort config.json with jq
			cmd := exec.Command("jq", "to_entries | sort_by(.key) | from_entries", configFilePath)
			sortedOutput, err := cmd.Output()
			if err != nil {
				log.Fatalf("Error sorting config.json with jq: %v", err)
			}

			err = ioutil.WriteFile(configFilePath, sortedOutput, 0644)
			if err != nil {
				log.Fatalf("Error writing sorted config.json: %v", err)
			}

			// Generate validator.key file with private key
			keyIndex := configValidator.Key
			if keyIndex >= 0 && keyIndex < len(keyOutput.Keys) {
				privateKey := keyOutput.Keys[keyIndex].PrivateKey
				keyContent := fmt.Sprintf("\"%s\"", privateKey)

				keyFilePath := filepath.Join(dirPath, "validator_key.json")
				err = ioutil.WriteFile(keyFilePath, []byte(keyContent), 0644)
				if err != nil {
					log.Fatalf("Error writing validator.key to %s: %v", keyFilePath, err)
				}
			}

			// Copy keystore.json to validator directory
			keystoreFilePath := filepath.Join(dirPath, "keystore.json")
			err = ioutil.WriteFile(keystoreFilePath, keystoreData, 0644)
			if err != nil {
				log.Fatalf("Error writing keystore.json to %s: %v", keystoreFilePath, err)
			}

			fmt.Printf("Generated genesis.json, config.json, validator.key, and keystore.json for %s in %s\n", configValidator.Profile, dirPath)
		}
	}
}
