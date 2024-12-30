package rpc
import(
  "context"
  "github.com/scatkit/pumpdexer/solana"
)

type result string

func (cl *Client) ReqAirdrop(ctx context.Context, walletAddress solana.PublicKey, lamports uint64, commitment CommitmentType,
) (out result, err error){
  params := []interface{}{
    walletAddress, 
    lamports,    
    map[string]interface{}{
      "commitment": commitment,
    },
  }
  
  err = cl.rpcClient.CallForInfo(ctx, &out, "requestAirdrop", params)
  return 
}
