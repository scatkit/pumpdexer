package main
import (
  "fmt"
  "context"
  "math/big"
  //"github.com/davecgh/go-spew/spew"
  "github.com/gagliardetto/solana-go"
  "github.com/gagliardetto/solana-go/rpc"
  "github.com/gagliardetto/solana-go/rpc/ws"
)
  
func convertToSol(lamport uint64) (solAmount *big.Float){
  var lamportAmount = new(big.Float).SetUint64(lamport)
  res :=  new(big.Float).Quo(lamportAmount, new(big.Float).SetUint64(solana.LAMPORTS_PER_SOL))
  return res
}

func main(){
  ctx := context.Background()
  client, err := ws.Connect(context.Background(), rpc.MainNetBeta_WS)
  if err != nil{
    panic(err)
  }
  program := solana.MustPublicKeyFromBase58("8LqocGsMwPJ7h2s1r8k4Vmc9c222Z4fMae25uz58qb3n")
  
  sub, err := client.AccountSubscribe(
    program,
    "",
  )
  if err != nil{
    panic(err)
  }
  defer sub.Unsubscribe()

  for {
    got, err := sub.Recv(ctx)
    if err != nil{
      fmt.Errorf("Error: %v\n",err)
    }
    fmt.Printf("Pooled SOL:",convertToSol(got.Value.Account.Lamports)) 
  }
}
