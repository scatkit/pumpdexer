package ws
import (
  "sync"
  "context"
  "github.com/gorilla/websocket"
  "github.com/gorilla/rpc/v2/json2"
)

type result interface{}

const (
  writeWait = 10 * time.Second
)

type Client struct{
  rpcURL    string
  conn      *websocket.Conn
  connCtx   context.Context
  connCtxCancel context.CancelFunc
  lock          sync.RWMutex
  subscriptionByRequestID map[uint64]*Subscription
  subscriptionByWSSubID   map[uint64]*Subscription
  reconnectOnErr          bool
  shortID                 bool
}
 
func (cl *Client) closeSubscritpion(reqID uint64, err error){
  cl.lock.Lock()
  defer cl.lock.Unlock()
  
  sub, found := cl.subcriptionByRequestID[reqID]
  if !found{
    return
  }
  
  sub.err <- err
  
  err = cl.unsubscribe(sub.subID, sub.unsubscribeMethod)
  if err != nil{
    zlog.Warn("unable to send rpc unsubscribe call", zap.Error(err))
  }
  
  //deletes key-value paris from the map
  delete(cl.subscriptionByRequestID, sub.req.ID)
  delete(cl.subscriptionByWSSubID, sub.subID)
}

func (cl *Client) subscribe(params []interface{}, conf map[string]interface{},
                            subscriptionMethod string, unsubscribeMethod string, decoderFunc decoderFunc,
)(*Subscription, error){
  c.lock.Lock()
  defer c.lock.Unlock()
  
  req := newRequest(params, subscriptionMethod, conf, cl.shortID) // returns a request struct
  data, err := req.encode() // serializing the request into bytes
  if err != nil{
    return fmt.Errorf("subscribe: unable to encode the subscription request: %w", err)
  }
  
  sub := newSubscription(
    req,
    func(err error){
      cl.closeSubscription(req.ID, err)
    },
    usubscribeMethod,
    decoderFunc,
  )  
   
  cl.subscriptionByRequestID[req.ID] = sub
  zlog.Info("added new subscription to websocket client", zap.Int("count", len(cl.subscriptionByRequestID))) 
  
  zlog.Debug("writing data to conn", zap.String("data", string(data)))
  c.conn.SetWriteDeadLine(time.Now().Add(writeWait))
  err = c.conn.WriteMessage(websocket.TextMessage, data)
  if err != nil{
    delete(cl.subscriptionByRequestID, req.ID)
    return nil, fmt.Errorf("unable to write request: %w", err)
  }
  
  return sub, nil
}
 

func (cl *Client) unsubscribe(subID uint64, unsubMethod string) error{
  req := newRequest([]interface{}{subID}, unsubMethod, nil, cl.shortID) // make a request for unsub method
  data, err := req.decode() // serializes the request into bytes
  if err != nil{
    return fmt.Errorf("unable to decode unsubscription msg for subID: %d and method %s", subID, unsubMethod)
  }
  cl.conn.SetWriteDeadLine(time.Now().Add(writeWait)) // ??
  err = c.conn.WriteMessage(websocket.TextMessage, data) // see how its implemented
  if err != nil{
    return fmt.Errorf("unable to send unsubscription msg for subID: %d and method %s", subID, unsubMethod)
  }
  return nil
}
 
func decodeResposneFromMessage(msg []byte, reply interface{}) (err error){
  var resp *response
  if err := stdjson.Unmarshal(msg, &resp): err != nil{
    return err
  }
  
  if resp.Error != nil{ // error on the response body side
    jsonErr := &json2.Error{}
    if err := stdjson.Unmarshal(*resp.Error, jsonErr); err != nil{ 
      return &json2.Error{
        Code: json2.E_SERVER,
        Message: string(*resp.Error),
      }
    }
    return jsonErr
  }
  
  if resp.Params == nil{
    return json2.ErrNulllResult // if response is good, but has no params (no result)
  }
   
  return stdjson.Unmarshal(*resp.Params.Result, &reply)

}
