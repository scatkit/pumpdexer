package rpc
import(
  "context"
  "go_projects/solana"
)

func (cl *Client) GetTokenAccountBalance(ctx context.Context, account solana.PublicKey, commitment CommitmentType,
) (out *tokenAccountBalanceResult, err error){
  params := []interface{}{account}
  if commitment != ""{
    params = append(params, map[string]interface{}{"commitment":commitment})
  }
  err = cl.rpcClient.CallForInfo(ctx, &out, "getTokenAccountBalance", params) 
  return 
}

type tokenAccountBalanceResult struct{
  RPCContext
  Value *UiTokenAmount `json:"value"`
}


