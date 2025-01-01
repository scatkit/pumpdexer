package rpc

import (
	"context"
  "errors"
	"github.com/scatkit/pumpdexer/solana"
)

type GetSignatureStatusesResult struct {
	RPCContext
	Value []*SignatureStatusesResult
}

type SignatureStatusesResult struct {
	// The slot the transaction was porccessed.
	Slot               uint64                 `json:"slot"`
	Confirmantions     *uint64                `json:"confirmations"`
	Err                interface{}            `json:"err"`
	ConfirmationStatus ConfirmationStatusType `json:"confirmationStatus"`
}

func (cl *Client) GetSignatureStatuses(ctx context.Context, searchTransactionHistory bool, transactionSignatures ...solana.Signature,
) (out *GetSignatureStatusesResult, err error) {
	params := []interface{}{transactionSignatures}
  
  err = cl.rpcClient.CallForInfo(ctx, &out, "getSignatureStatuses", params)
  if err != nil{
    return nil, err
  }
  
  if out == nil || out.Value == nil{
    return nil, errors.New("Not found")
  }
  
  return out, nil
}
