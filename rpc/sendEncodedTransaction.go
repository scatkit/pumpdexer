package rpc
import(
  "github.com/scatkit/pumpdexer/solana"
  "context"
)
 
func (cl *Client) SendEncodedTransactionWithOpts(ctx context.Context, encodedTransaction string, opts TransactionOpts,
) (signature solana.Signature, err error){
  obj := map[string]interface{}{}
  
  if opts.Encoding == ""{
    // deafult to base64 encoding
    obj["encoding"] = "base64"
  } else{
    obj["encoding"] = opts.Encoding
  }
  
  obj["skipPreflight"] = opts.SkipPreflight

  if opts.PreflightCommitment != ""{
    obj["preflightCommitment"] = opts.PreflightCommitment
  }
  if opts.MaxRetries != nil{
    obj["maxRetries"] = opts.MaxRetries
  }
  if opts.MinContextSlot != nil{
    obj["minContextSlot"] = opts.MinContextSlot
  }
  params := []interface{}{encodedTransaction, obj}
  err = cl.rpcClient.CallForInfo(ctx, &signature, "sendTransaction", params)
  
  return signature, err 
}
