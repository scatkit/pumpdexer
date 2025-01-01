package rpc
import (
  "context"
  "github.com/scatkit/pumpdexer/solana"
)

// block is a Slot
// UnixTimeSeconds is int64
func (cl *Client) GetBlockTime(ctx context.Context, block uint64) (out *solana.UnixTimeSeconds, err error){
  params := []interface{}{block,}
  err = cl.rpcClient.CallForInfo(ctx, &out, "getBlockTime", params)
  return out, err
}


