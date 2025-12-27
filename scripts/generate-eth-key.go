// +build ignore

package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
)

type KeyOutput struct {
	PrivateKey string `json:"private_key"`
	Address    string `json:"address"`
}

func main() {
	// Generate new private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	// Get private key bytes (without 0x prefix)
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := fmt.Sprintf("%x", privateKeyBytes)

	// Get public key and derive address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	// Output as JSON
	output := KeyOutput{
		PrivateKey: privateKeyHex,
		Address:    address,
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// If filename provided as argument, write to file
	if len(os.Args) > 1 {
		err = os.WriteFile(os.Args[1], jsonBytes, 0600)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Keys saved to %s\n", os.Args[1])
	}

	// Always print to stdout
	fmt.Println(string(jsonBytes))
}
