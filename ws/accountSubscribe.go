package ws
import (
  //"fmt"
  "context"
  "errors"
  "github.com/scatkit/pumpdexer/rpc"
  "github.com/scatkit/pumpdexer/solana"
  //"github.com/davecgh/go-spew/spew"
)
 
// The params holds result
type AccountResult struct{
  Context struct{
    Slot uint64
  } `json:"context"`
  Value struct{
    rpc.Account
  } `json:"value"`
}

func (cl *Client) AccountSubscribeWithOpts(account solana.PublicKey, commitment rpc.CommitmentType, encoding solana.EncodingType, 
) (*AccountSubscription, error) {
  params := []interface{}{account.String()}
  conf := map[string]interface{}{"encoding": "base64"}
  if encoding != ""{
    conf["encoding"] = encoding
  }
  if commitment != ""{
    conf["commitment"] = commitment
  }
  
  genSub, err := cl.subscribe( // returns a new subscription
    params,
    conf,
    "accountSubscribe",
    "accountUnsubscribe",
    func(msg []byte) (interface{}, error){
      var acc_res AccountResult
      err := decodeResponseFromMessage(msg, &acc_res)
      return &acc_res, err
    },
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

func (ac *AccountSubscription) Recv(ctx context.Context) (*AccountResult, error){
  select{
    case <- ctx.Done():
      return nil, ctx.Err()
    case d, ok := <- ac.sub.stream:
      if !ok{
        return nil, errors.New("sub is no longer active")
      } 
      // type assertion tells the complier: trust me, d contains a value of AccountResult
      return d.(*AccountResult), nil // d (interface contating a result) is evaluated as AccountResult and returned
    case err := <-ac.sub.err:
      return nil, err
  }
}
 
func (ac *AccountSubscription) Err() <-chan error{
  return ac.sub.err 
}
 
func (ac *AccountSubscription) Response() <-chan *AccountResult{
  typedChan := make(chan *AccountResult, 1)
  go func(ch chan *AccountResult){
    d, ok := <-ac.sub.stream
    if !ok {
      return
    }
    ch <- d.(*AccountResult)
  }(typedChan)
  return typedChan 
}
 
func (ac *AccountSubscription) Unsubscribe(){
  ac.sub.Unsubscribe()
}
