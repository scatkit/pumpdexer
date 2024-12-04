package rpc
import (
  "context"
  "github.com/scatkit/pumpdexer/solana"
)
 
func (cl *Client) GetTokenSupply(ctx context.Context, account solana.PublicKey, commitment CommitmentType,
) (out *GetTokenSupplyResult, err error){
  params := []interface{}{account}
  if commitment != ""{
    params = append(params, map[string]interface{}{"commitment":commitment})
  }
  err = cl.rpcClient.CallForInfo(ctx, &out, "getTokenSupply", params)
  
  return out, err
}
 
type GetTokenSupplyResult struct{
  RPCContext
  Value *UiTokenAmount `json:"value"`
}
