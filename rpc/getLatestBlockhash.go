package rpc
import (
  "context"
  "github.com/scatkit/pumpdexer/solana"
)

func (cl *Client) GetLatestBlockhash(ctx context.Context, commitment CommitmentType,
) (out *GetLatestBlockhashResult, err error){
  params := []interface{}{}
  if commitment != ""{
    params = append(params, map[string]interface{}{"commitment": commitment})
  }
  
  cl.rpcClient.CallForInfo(ctx, &out, "getLatestBlockhash", params)
  return 
}

type GetLatestBlockhashResult struct{
  RPCContext
  Value *LatestBlockhashResult `json:"value"`
}
 
type LatestBlockhashResult struct{
  Blockhash solana.Hash `json:"blockhash"`
  LastValueBlockHeight uint64 `json:"lastValidBlockHeight"`
  
}

