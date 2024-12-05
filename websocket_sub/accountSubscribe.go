package ws
import (
  "errors"
  "context"
  "github.com/scatkit/pumpdexer/solana"
  "github.com/scatkit/pumpdexer/rpc"
)
 
type AccountResult struct{
  Context struct{
    Slot uint64
  } `json:"context"`
  Value struct{
    rpc.Account
  } `json:"value"`
}

func (cl *Client) AccountSubscribeWithOpts(account solana.PublicKey, commitment CommitmentType, encoding solana.EncodingType, 
) (*AccountSubscription, error) {
  params := []interface{}{account.String()}
  conf := map[string]interface{}{"encoding": "base64"}
  if encodign != ""{
    conf["encoding"] = encoding
  }
  if commitment != ""{
    conf"commitment"] = commitment
  }
  
  genSub, err := cl.subscribe(
    params,
    conf,
    "accountSubscribe",
    "accountUnsubscribe",
    func(msg []byte) (error, interface{}){
      var acc_res AccountResult
      err := decodeResponseFromMessage(msg, &acc_res)
      return &acc_res, err
    }
  )
  
  if err != nil{
    return nil, err
  }
  
  return &AccountSubscription{
    sub: genSub,
  }, nil
}

type AccountSubscription struct{
  sub *Subscription
}

func (sw *AccountSubcription) Recv(ctx context.Context) (*AccountResult, error){
  select{
    case <- ctx.Done():
      return nil, ctx.Err()
    case d,ok := <- sw.sub.stream:
      if !ok{
        return nil, errors.New("sub is no longer active")
      } 
      // type assertion tells the complier: trust me, d contains a value of AccountResult
      return d.(*AccountResult), nil // d (interface contating a result) is evaluated as AccountResult and returned
  }
}
 
func (sw *AccountSubscription) Err() <-chan error{
  return sw.sub.err 
}
 
func (sw *AccountSubscription) Response() <-chan *AccountResult{
  typedChan := make(chan *AccountResult, 1)
  go func(ch chan *AccountResult){
    d, ok := <-sw.sub.stream
    if !ok {
      return
    }
    ch <- d.(*AccountResult)
  }
  )(typedChan)
  return typedChan
}
 
func (sw *AccountSubscription) Unsubscribe(){
  sw.sub.Unsubscribe()
}
