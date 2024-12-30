package rpc
import(
  "context"
  "encoding/base64"
  "fmt" 
  "github.com/scatkit/pumpdexer/solana"
)


func (cl *Client) SendTransaction(ctx context.Context, transaction *solana.Transaction,
) (signature solana.Signature, err error){
  opts := TransactionOpts{
    SkipPreflight: false,
    PreflightCommitment: "",
  }
  
  return cl.SendTransactionWithOpts(ctx, transaction, opts) 
}

func (cl *Client) SendTransactionWithOpts(ctx context.Context, transaction *solana.Transaction, opts TransactionOpts,
) (signature solana.Signature, err error){
  txData, err := transaction.MarshalBinary()
  if err != nil{
    return solana.Signature{}, fmt.Errorf("send transaction: encode transaction: %w", err)
  }
  
  return cl.SendEncodedTransactionWithOpts(ctx, base64.StdEncoding.EncodeToString(txData), opts)
}

