package main

import (
	"context"
	"log"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/memoprog"
	"github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/types"

	"crypto/rand"
	"encoding/base64"
	"flag"
	"github.com/omarkilani/hide_on_sol/ecies"
)

// FUarP2p5EnxD66vVDL4PWRoWMzA56ZVHG24hpEDFShEz
var feePayer, _ = types.AccountFromBase58("4TMFNY9ntAn3CHzguSAvDNLPRoQTaK3sWbQQXdDXaE6KWRBLufGL6PJdsD2koiEe3gGmMdRK3aAw7sikGNksHJrN")

// 9aE476sH92Vz7DMPyq5WLPkrKWivxeuTKEFKd2sZZcde
var alice, _ = types.AccountFromBase58("4voSPg3tYuWbKzimpQK9EbXHmuyy5fUrtXvpLDMLkmY6TRncaTHAKGD8jUg3maB5Jbrd9CkQg4qjJMyN6sQvnEF2")

func GetEncryptedBytes(dec []byte) []byte {
	prv, err := ecies.GenerateKey(rand.Reader, ecies.DefaultCurve, nil)
	if err != nil {
		log.Fatal(err)
	}

	message := []byte(dec)
	ct, err := ecies.Encrypt(rand.Reader, &prv.PublicKey, message, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	return ct
}

func EncodeBase64(b []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(b))
}

var PlainText string

func InitFlags() {
	flag.StringVar(&PlainText, "text", "",
		"plain text to encrypt and store on chain")
	flag.Parse()

	if PlainText == "" {
		log.Fatal("must specify text")
	}
}

func main() {
	InitFlags()

	c := client.NewClient(rpc.DevnetRPCEndpoint)

	// to fetch recent blockhash
	recentBlockhashResponse, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		log.Fatalf("failed to get recent blockhash, err: %v", err)
	}

	// create a tx
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{feePayer, alice},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        feePayer.PublicKey,
			RecentBlockhash: recentBlockhashResponse.Blockhash,
			Instructions: []types.Instruction{
				// memo instruction
				memoprog.BuildMemo(memoprog.BuildMemoParam{
					SignerPubkeys: []common.PublicKey{alice.PublicKey},
					Memo:          EncodeBase64(GetEncryptedBytes([]byte(PlainText))),
				}),
			},
		}),
	})

	if err != nil {
		log.Fatalf("failed to new a transaction, err: %v", err)
	}

	// send tx
	txhash, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Fatalf("failed to send tx, err: %v", err)
	}

	log.Println("txhash:", txhash)
}
