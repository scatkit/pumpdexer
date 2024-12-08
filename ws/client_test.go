package ws
import(
  "fmt"
  "context"
  "go.uber.org/zap"
  "github.com/stretchr/testify/require"
  "testing"
  "github.com/scatkit/pumpdexer/solana"
)

func Test_AccountSubscribe(t *testing.T){
  //t.Skip("some text")
  
  zlog,_ = zap.NewDevelopment()
  co, err := ConnectWithOptions(context.Background(), "ws://api.mainnet-beta.solana.com:80", nil)
  defer co.Close()
  require.NoError(t, err)
  
  poolID := solana.MustPubkeyFromBase58("ABLmVkXfVNwuBu6J2mYyxDXRTzqnGj22c6vBTGgELLn2")
  sub, err := co.AccountSubscribeWithOpts(poolID,"","")
  require.NoError(t, err)
  
  data, err := sub.Recv(context.Background())
  if err != nil{
    fmt.Errorf("received error: %w", err)
    return
  }
  fmt.Println(data) 
  return
}
