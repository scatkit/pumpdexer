package rpc
import(
  "context"
  "github.com/scatkit/pumpdexer/solana"
)

func (cl *Client) GetTokenAccountBalance(ctx context.Context, account solana.PublicKey, commitment CommitmentType,
) (out *GetTokenAccountBalanceResult, err error){
  params := []interface{}{account}
  if commitment != ""{
    params = append(params, map[string]interface{}{"commitment":commitment})
  }
  err = cl.rpcClient.CallForInfo(ctx, &out, "getTokenAccountBalance", params) 
  return 
}

type GetTokenAccountBalanceResult struct{
  RPCContext
  Value *UiTokenAmount `json:"value"`
}


