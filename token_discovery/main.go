package main

import (
  "context"
	"fmt"
  //"encoding/binary"
  //"bytes"
  "github.com/davecgh/go-spew/spew"
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
  //"math/big"
  //"math"
)


func main() {
	client := rpc.New("https://api.mainnet-beta.solana.com")
	poolID := solana.MustPublicKeyFromBase58("2oiqD2upcVZggDjoCk6KwiBHLojgRQnZNTGzAGPRd4bA")
	resp, err := client.GetAccountInfo(context.TODO(), poolID)
  if err != nil{
    panic(err)
  }
  res := getPoolInfo(resp.GetBinary())
  fmt.Println()
  spew.Dump(res)
	//if err != nil {
	//	fmt.Errorf("Error: %v", err)
	//}
  //
  //out := resp.GetBinary()
  //pool := getPoolInfo(out) 
  //
  //baseToken,err := client.GetBalance(context.TODO(), pool.BaseVault, rpc.CommitmentFinalized)
  //if err != nil{
  //  fmt.Errorf("error: %v\n",err)
  //}

  //quoteToken, err := client.GetTokenAccountBalance(context.TODO(),pool.QuoteVault, rpc.CommitmentFinalized)
  //if err != nil{
  //  fmt.Errorf("Error %v\n",err)
  //}
  //spew.Dump(quoteToken)
  //
  //var lamportsBase =  new(big.Float).SetUint64(baseToken.Value)
  //var solBalance = new(big.Float).Quo(lamportsBase, new(big.Float).SetUint64(solana.LAMPORTS_PER_SOL))
  //fmt.Printf("Pooled SOL: %.2f\n",solBalance)
}
