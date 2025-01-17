package rpc
import (
  "context"
  //"github.com/scatkit/pumpdexer/solana"
)

func (cl *Client) GetMinimumBalanceForRentExemption(ctx context.Context, length uint64, commitment CommitmentType,
) (lamports uint64, err error){ 
  params := []interface{}{length}
  if commitment != ""{
    params = append(params, map[string]interface{}{"commitment": commitment})
  }
  err = cl.rpcClient.CallForInfo(ctx, &lamports, "getMinimumBalanceForRentExemption", params)

  return lamports, err 
}
