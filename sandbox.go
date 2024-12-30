package main

import (
	"context"

	"github.com/scatkit/pumpdexer/solana"
	"github.com/scatkit/pumpdexer/ws"
	"github.com/scatkit/pumpdexer/rpc"
	"github.com/scatkit/pumpdexer/dexes"
  "fmt"
	"github.com/davecgh/go-spew/spew"

)

const URL_ENDPOINT = "wss://api.mainnet-beta.solana.com"

func main_changeLater() {
	ws_client, err := ws.ConnectWithOptions(context.Background(), URL_ENDPOINT, nil)
  http_client := rpc.New("https://api.mainnet-beta.solana.com")
  
	if err != nil {
		panic(err)
	}
	defer ws_client.Close()
	poolID := solana.MustPubkeyFromBase58("6tpCWpvihiRkF3G7ZKGE8T3jCbMs8kvxsw8hRz1JJz6Z")

	sub, err := ws_client.AccountSubscribeWithOpts(poolID, "", solana.EncodingBase64)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()
  
  for i := 0; i<10; i++{
    //resp, err := sub.Recv(context.Background())
    resp, err := http_client.GetAccountInfo(context.Background(), poolID)
    if err != nil{
      panic(err)
    }
    spew.Dump(resp)
    fmt.Println("-=-=-=-=-=-=")
    data := dexes.GetPoolInfo(resp.Value.Data.GetBinary())
    spew.Dump(data)
  }
}
  
	//for {
	//	_, err := sub.Recv(context.Background())
	//	if err != nil {
	//		panic(err)
	//	}
  //  
  //  transactions, err := http_client.GetSignaturesForAddress(context.Background(), poolID) 
  //  if err != nil{
  //    log.Printf("Error fetching\n")
  //    spew.Dump(err)
  //  }
  //  if len(transactions) > 0{
  //    spew.Dump(transactions[0])
  //  }
	//}
