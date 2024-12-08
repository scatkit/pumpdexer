package main
import(
  "context"
  "fmt"
  "github.com/scatkit/pumpdexer/ws" 
  "github.com/scatkit/pumpdexer/solana" 
  //"github.com/scatkit/pumpdexer/dexes" 
  //"github.com/davecgh/go-spew/spew"
)
 
const  URL_ENDPOINT = "wss://api.mainnet-beta.solana.com"

func main(){
  client, err := ws.ConnectWithOptions(context.Background(), URL_ENDPOINT,nil)
  if err != nil{
    panic(err)
  }
  defer client.Close()
  poolID := solana.MustPubkeyFromBase58("ABLmVkXfVNwuBu6J2mYyxDXRTzqnGj22c6vBTGgELLn2")
  
  sub, err := client.AccountSubscribeWithOpts(poolID, "", solana.EncodingBase64)
  //spew.Dump(sub)
  if err != nil{
    panic(err)
  }
  defer sub.Unsubscribe()
  
  for {
    // fmt.Println("in the loop")
    got, err := sub.Recv(context.Background())
    if err != nil{
      panic(err)
    }  
    fmt.Printf("%T\n",got)
    //spew.Dump(dexes.GetPoolInfo(got.Value.Data.GetBinary()))
    }
}
