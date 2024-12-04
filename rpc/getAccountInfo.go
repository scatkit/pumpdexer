package rpc
import( 
  "fmt"
  "errors"
  "context"
  "github.com/scatkit/pumpdexer/solana"
)

func (cl *Client) GetAccountInfo(ctx context.Context, account solana.PublicKey) (output *AccountInfoResult, err error){
  return cl.GetAccountInfoWithOpts(ctx, account, &GetAccountInfoOpts{Commitment:"", DataSlice: nil,},)
}

type GetAccountInfoOpts struct{
  Encoding        solana.EncodingType // <- string
  Commitment      CommitmentType
  DataSlice       *DataSlice  
  MinContextSlot *uint64 // <- minimum slot that the request can be evaluated at
}


func (cl *Client) GetAccountInfoWithOpts(ctx context.Context, account solana.PublicKey, opts *GetAccountInfoOpts) (*AccountInfoResult, error){
  out, err := cl.getAccountInfoWithOpts(ctx, account, opts)
  if err != nil{
    return nil, err
  }
  if out == nil || out.Value == nil{
    return nil, fmt.Errorf("Value is just empty")
  }
  return out, nil
}
 

func (cl *Client) getAccountInfoWithOpts(ctx context.Context, account solana.PublicKey, opts *GetAccountInfoOpts) (out *AccountInfoResult, err error){ 
  obj := map[string]interface{}{
    "encoding": solana.EncodingBase64, // base case encoding 
  }
  if opts != nil{
    if opts.Encoding != ""{
      obj["encoding"] = opts.Encoding
    }
    if opts.Commitment != ""{
      obj["commitment"] = opts.Commitment
    }
    if opts.DataSlice != nil{
      obj["dataSlice"] = map[string]interface{}{
        "offset": opts.DataSlice.Offset,
        "length": opts.DataSlice.Length,
      }
      if opts.Encoding == solana.EncodingJsonParsed{
        return nil, errors.New("cannot use DataSlice with jsonPrased encoding")
      }
    }
    if opts.MinContextSlot != nil{
      obj["minContextSlot"] = *opts.MinContextSlot
    }
  }
  
  params := []interface{}{account}
  if len(obj) > 0 {
    params = append(params,obj)
  }
  err = cl.rpcClient.CallForInfo(ctx, &out, "getAccountInfo", params) // <-- out is AccountInfoResult struct
  if err != nil{
    return nil, err
  }
  if out == nil{
    return nil, errors.New("expected a value, got null result")
  }
  
  return out, nil
}
