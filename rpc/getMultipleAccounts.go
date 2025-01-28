package rpc
import (
  "context"
  "errors"
  "github.com/scatkit/pumpdexer/solana"
)

type GetMultipleAccountsResult struct{
  RPCContext
  Value []*Account
}

type GetMultipleAccountsOpts GetAccountInfoOpts

func (cl *Client) GetMultipleAccounts(ctx context.Context, accounts ...solana.PublicKey,
) (out *GetMultipleAccountsResult, err error){
  return cl.GetMultipleAccountsWithOpts(ctx, accounts, nil)
}

func (cl *Client) GetMultipleAccountsWithOpts(ctx context.Context, accounts []solana.PublicKey, opts *GetMultipleAccountsOpts,
) (out *GetMultipleAccountsResult, err error){
  params := []interface{}{accounts}

	if opts != nil {
		obj := map[string]interface{}{}
		if opts.Encoding != "" {
			obj["encoding"] = opts.Encoding
		}
		if opts.Commitment != "" {
			obj["commitment"] = opts.Commitment
		}
		if opts.DataSlice != nil {
			obj["dataSlice"] = map[string]interface{}{
				"offset": opts.DataSlice.Offset,
				"length": opts.DataSlice.Length,
			}
			if opts.Encoding == solana.EncodingJsonParsed {
				return nil, errors.New("cannot use dataSlice with EncodingJSONParsed")
			}
		}
		if opts.MinContextSlot != nil {
			obj["minContextSlot"] = *opts.MinContextSlot
		}
		if len(obj) > 0 {
			params = append(params, obj)
		}
	}

	err = cl.rpcClient.CallForInfo(ctx, &out, "getMultipleAccounts", params)
	if err != nil {
		return nil, err
	}
	if out == nil || out.Value == nil {
		return nil, errors.New("not found")
	}
	return
}

