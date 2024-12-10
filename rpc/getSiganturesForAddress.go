package rpc
import (
  "context"
  "github.com/scatkit/pumpdexer/solana"
)

type GetSignaturesForAddressOpts struct{
  Commitment CommitmentType `json:"commitment,omitempty"`
  MinContextSlot *uint64 
  Limit *int `json:"limit,omitempty"`
  Before solana.Signature `json:"before,omitempty"`
  Until solana.Signature `json:"until,omitempty"`
}

func (cl *Client) GetSignaturesForAddress(ctx context.Context, account solana.PublicKey,
) ([]*TransactionSignature, error){
  return cl.GetSignaturesForAddressWithOpts(ctx, account, nil)
}

func (cl *Client) GetSignaturesForAddressWithOpts(ctx context.Context, account solana.PublicKey, opts *GetSignaturesForAddressOpts,
) (out []*TransactionSignature, err error){
  params := []interface{}{account}
  if opts != nil{
    obj := map[string]interface{}{}
    if opts.Commitment != "" {
      obj["commitment"] = opts.Commitment
    }
    if opts.MinContextSlot != nil {
      obj["minContextSlot"] = *opts.MinContextSlot
    }
		if opts.Limit != nil {
			obj["limit"] = opts.Limit
		}
		if !opts.Before.IsZero() {
			obj["before"] = opts.Before
		}
		if !opts.Until.IsZero() {
			obj["until"] = opts.Until
		}
		if len(obj) > 0 {
			params = append(params, obj)
		}
  }
  
  err = cl.rpcClient.CallForInfo(ctx, &out, "getSignaturesForAddress", params)
  return out, err
}
