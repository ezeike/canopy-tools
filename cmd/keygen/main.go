package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/canopy-network/canopy/lib/crypto"
)

const password = "test"
const dataDirPath = "keys/"
const nickPrefix = "nick"

type KeyPair struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
	Address    string `json:"address"`
}

type KeyOutput struct {
	Timestamp string    `json:"timestamp"`
	Keys      []KeyPair `json:"keys"`
}

func main() {
	var keys []KeyPair

	os.Remove(dataDirPath + "/keystore.json")

	// load the keystore from file
	k, e := crypto.NewKeystoreFromFile(dataDirPath)
	if e != nil {
		panic(e)
	}

	for i := 0; i < 12; i++ {
		blsKey, _ := crypto.NewBLS12381PrivateKey()
		blsPub := blsKey.PublicKey()

		keyPair := KeyPair{
			PrivateKey: blsKey.String(),
			PublicKey:  blsPub.String(),
			Address:    blsPub.Address().String(),
		}
		keys = append(keys, keyPair)

		// import each key to keystore with same password
		address, e := k.ImportRaw(blsKey.Bytes(), password, crypto.ImportRawOpts{
			Nickname: fmt.Sprintf("%s-%d", nickPrefix, i),
		})
		if e != nil {
			log.Fatal(e.Error())
		}
		fmt.Printf("Imported validator key %s to keystore\n", address)
	}

	// save keystore to file once after all imports
	if e = k.SaveToFile(dataDirPath); e != nil {
		panic(e)
	}

	output := KeyOutput{
		Timestamp: time.Now().Format("2006-01-02T15:04:05Z"),
		Keys:      keys,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling to JSON: %v", err)
	}

	fmt.Println(string(jsonData))

	err = ioutil.WriteFile("keys/node-bls.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}

	fmt.Printf("\nKeys saved to: %s\n", "/keys/node-bls.json")
}
