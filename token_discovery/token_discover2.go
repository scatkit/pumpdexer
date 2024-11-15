package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const solanaAPIURL = "https://api.mainnet-beta.solana.com" // Solana mainnet URL

type RpcRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Id      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type tokenAccount struct{
  //Pubkey string `json:"pubkey"`
  Account struct{
    Data struct {
      Parsed struct{
        Info struct{
          Mint string `json:"mint"`
        } `json:"info"`
      } `json:"parsed"`
    } `json:"data"`
  } `json:"account"`
}

type RpcResponse struct { 
	Result struct{
    Value []tokenAccount `json:"value"`
    } `json:"result"`
    Error  interface{} `json:"error"`
}

func fetchTokenAccountsByOwner(owner string) ([]tokenAccount, error) {
	// Solana JSON-RPC call to get token accounts by owner
  params := []interface{}{
    owner,
    map[string]string{
      "programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" ,
    },
    map[string]interface{}{
      "encoding": "jsonParsed",
    },
  }
  
  // JSON-RPC payload to be sent to Solana API 
	request := RpcRequest{
		Jsonrpc: "2.0",
		Id:      1,
		Method:  "getTokenAccountsByOwner",
		Params:  params,
	}

	// Marshalling request into JSON
	requestData, err := json.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}

	// Sending the request to Solana API
	resp, err := http.Post(solanaAPIURL, "application/json", bytes.NewBuffer(requestData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Reading response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parsing JSON response
	var rpcResp RpcResponse
  err = json.Unmarshal(body, &rpcResp)
	if err != nil{
		return nil, err
	}

	// Return the list of token accounts
  if rpcResp.Error != nil{
    return nil, fmt.Errorf("API error: %v", rpcResp.Error)
  }
  
  return rpcResp.Result.Value, nil
  
}

func main() {
	// Example owner address (replace with a real address)
	owner := "3w9kbaohDgz4jm62W935hT2i1LW5EWZWuHHcDB6oWNoN"

	// Fetch the token accounts
	tokenAccounts, err := fetchTokenAccountsByOwner(owner)
	if err != nil {
		log.Fatal(err)
	}
  fmt.Println(len(tokenAccounts))

	// Print out the list of token accounts
	//for _, account := range tokenAccounts {
	//	// You would implement additional filtering or processing to list only meme tokens
	//	// For now, just printing all token accounts
  //  fmt.Println("Account:",account.Account.Data.Parsed.Info.Mint)
	//}
}

