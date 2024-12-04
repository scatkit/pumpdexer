package main
import(
  //"time"
  "context" 
  "fmt"
  rpc "go_projects/rpc"
  "go_projects/solana"
  "go_projects/dexes"
  "github.com/davecgh/go-spew/spew"
  "math/big"
)

const rpcURL = "https://api.mainnet-beta.solana.com" 
const pubKEY = "3pvmL7M24uqzudAxUYmvixtkWTC5yaDhTUSyB8cewnJK"


func main() {
  pub_key := solana.MustPubkeyFromBase58(pubKEY)
  client := rpc.New(rpcURL) 
  resp, _ := client.GetAccountInfo(context.Background(), pub_key)
  
  poolInfo := dexes.GetPoolInfo(resp.Value.Data.GetBinary())
  spew.Dump(poolInfo)
  fmt.Printf("Address of pair: %v\n",pub_key.String())
  fmt.Printf("Address of base(meme) token: %v\n",poolInfo.BaseVault.String())
  fmt.Printf("Address of quote(sol) token: %v\n",poolInfo.QuoteVault.String())
  base_token, err := client.GetTokenAccountBalance(context.Background(), poolInfo.BaseVault)
  if err != nil{
    fmt.Errorf("Error: %v\n", err)
  }
  quote_token, err := client.GetBalance(context.Background(), poolInfo.QuoteVault)
  if err != nil{
    fmt.Errorf("Error: %v\n", err)
  }
  lamports := new(big.Float).SetUint64(uint64(quote_token.Value))
  solAmount := new(big.Float).Quo(lamports, new(big.Float).SetUint64(uint64(solana.LAMPORTS_PER_SOL)))
  
  fmt.Printf("Pooled MEME token: %f\n", *base_token.Value.UiAmount)
  fmt.Printf("Pooled SOL: %v\n", solAmount)
  fmt.Println()
  
  sol_price := big.NewFloat(235.93)
  token_price_in_sol := new(big.Float).Quo(solAmount,new(big.Float).SetUint64(uint64(*base_token.Value.UiAmount)))
  token_price_in_usd := new(big.Float).Mul(token_price_in_sol, sol_price)
  fmt.Printf("Token price in SOL: %f\n",token_price_in_sol)
  fmt.Printf("Token price in USD: %f\n",token_price_in_usd)
  token_in_sol := new(big.Float).Mul(token_price_in_sol, big.NewFloat(*base_token.Value.UiAmount))
  totalSol := new(big.Float).Add(token_in_sol, solAmount)
  fmt.Printf("Liqudity: %f\n", new(big.Float).Mul(sol_price, totalSol))
  
  
 }
