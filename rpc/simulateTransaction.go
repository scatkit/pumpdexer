package rpc
import (
  "context"
  "fmt"
  "encoding/base64"
  "github.com/scatkit/pumpdexer/solana"
)


type SimulateTransactionResponse struct{
  Err interface{}       `json:"err,omitmepty"`
  Logs []string         `json:"logs,omitempty"`
  Accounts []*Account   `json:"accounts"`
  UnitsConsumed *uint64 `json:"unitsConsumed,omitempty"`
}

type SimulateTransactionOpts struct{
  // If true the transaction signatures will be verified
	// (default: false, conflicts with ReplaceRecentBlockhash)
  SigVerify bool
  // Commitment level to simulate the transaction at.
	// (default: "finalized").
	Commitment CommitmentType
  // If true the transaction recent blockhash will be replaced with the most recent blockhash.
	// (default: false, conflicts with SigVerify)
	ReplaceRecentBlockhash bool
  Accounts *SimulateTransactionAccountOpts
}

type SimulateTransactionAccountOpts struct{
  Encoding solana.EncodingType
  Addresses []solana.PublicKey 
}

func (cl *Client) SimulateTransaction(
	ctx context.Context,
	transaction *solana.Transaction,
) (out *SimulateTransactionResponse, err error) {
	return cl.SimulateTransactionWithOpts(
		ctx,
		transaction,
		nil,
	)
}

func (cl *Client) SimulateTransactionWithOpts(ctx context.Context, transaction *solana.Transaction, opts *SimulateTransactionOpts,
) (out *SimulateTransactionResponse, err error){
  txData, err := transaction.MarshalBinary()
  if err != nil{
    return nil, fmt.Errorf("SimulateTransactionWithOpts: encode the `transaction`: %w",err)
  }
  return cl.SimulateRawTransactionWithOpts(ctx, txData, opts)
}

func (cl *Client) SimulateRawTransactionWithOpts(ctx context.Context, transactionData []byte, opts *SimulateTransactionOpts,
) (out *SimulateTransactionResponse, err error){
  obj := map[string]interface{}{"encoding":"base64"}
  
  if opts != nil{
    if opts.Commitment != ""{
      obj["commitment"] = opts.Commitment
    }
    if opts.SigVerify{
      obj["sigVerify"] = opts.SigVerify
    }
    if opts.ReplaceRecentBlockhash{
      obj["replaceRecentBlockhash"] = opts.ReplaceRecentBlockhash
    }
    if opts.Accounts != nil{
      obj["accounts"] = map[string]interface{}{
        "encoding": opts.Accounts.Encoding,
        "addresses": opts.Accounts.Addresses,
      }
    }
  }
  
  b64Data := base64.StdEncoding.EncodeToString(transactionData)
  params := []interface{}{b64Data,obj,}
  
  err = cl.rpcClient.CallForInfo(ctx, &out, "simulateTransaction", params)
  return out, err
}
