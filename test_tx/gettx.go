package main
import(
  "context"
  "fmt"
  "github.com/scatkit/pumpdexer/rpc"
  "github.com/scatkit/pumpdexer/solana"
  "github.com/davecgh/go-spew/spew"

)

func main(){
  cl := rpc.New("https://api.mainnet-beta.solana.com")
  
  //sig := solana.MustSignatureFromBase58("PMkj5Rj6pwrVyAbmc99bWujHTcmHL1LSVDTD6PSEq8nSSji5X4YwVHYxSng9GYEAeUBS8XE8HVBhsPoQNiQhh2Y")
  //version := uint64(1)
  //tx, err := cl.GetTransaction(context.Background(), sig, &rpc.GetTransactionOpts{
  //  Commitment: rpc.CommitmentConfirmed,
  //  MaxSupportedTransactionVersion: &version,
  //}) 
  //if err != nil{
  //  fmt.Println(err)
  //  return
  //}
  //
  //spew.Dump(tx)
}

