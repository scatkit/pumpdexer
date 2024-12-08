package ws
import (
  "fmt"
  "io"
  "time"
  "sync"
  "context"
  "strconv"
  "net/http"
  "github.com/scatkit/pumpdexer/json2"
  "github.com/gorilla/websocket"
  "github.com/buger/jsonparser"
  "go.uber.org/zap"
  //"github.com/davecgh/go-spew/spew"
)

type result interface{}

const (
  // max time allowed to write a message to the peer
  writeWait = 10 * time.Second
  // max time allowed to read the next pong message from the peer
  pongWait = 60 * time.Second
  // Send pings to peer with the 90% of the pong interval
  pingPeriod = (pongWait * 9) / 10
)

type Client struct{
  rpcURL    string 
  conn      *websocket.Conn
  connCtx       context.Context
  connCtxCancel context.CancelFunc
  lock          sync.RWMutex
  subscriptionByRequestID map[uint64]*Subscription
  subscriptionByWSSubID   map[uint64]*Subscription
  reconnectOnErr          bool
  shortID                 bool
}
 
func (cl *Client) Close(){
  cl.lock.Lock()
  defer cl.lock.Unlock()
  cl.connCtxCancel()
  cl.conn.Close()
}

func (cl *Client) closeSubscription(reqID uint64, err error){
  cl.lock.Lock()
  defer cl.lock.Unlock()
  
  sub, found := cl.subscriptionByRequestID[reqID]
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
  cl.lock.Lock()
  defer cl.lock.Unlock()
  
  req := newRequest(params, subscriptionMethod, conf, cl.shortID) // returns a request struct
  data, err := req.encode() // serializing the request into bytes
  if err != nil{
    return nil, fmt.Errorf("subscribe: unable to encode the subscription request: %w", err)
  }
  
  //fmt.Println("Created a new request:") 
  //spew.Dump(req)
  //fmt.Println()
  
  sub := newSubscription(
    req,
    func(err error){
      cl.closeSubscription(req.ID, err)
    },
    unsubscribeMethod,
    decoderFunc,
  )  
  //fmt.Println("New sub created:")
  //spew.Dump(sub)
  //fmt.Println()
   
  cl.subscriptionByRequestID[req.ID] = sub
  zlog.Info("added new subscription to websocket client", zap.Int("count", len(cl.subscriptionByRequestID))) 
  zlog.Debug("writing data to conn", zap.String("data", string(data)))
  
  // sets a deadline to the server (as client) for completing the write operations
  cl.conn.SetWriteDeadline(time.Now().Add(writeWait))
  
  // websocket.TextMessage - sends message astext payload UTF-8 encoded
  err = cl.conn.WriteMessage(websocket.TextMessage, data)
  if err != nil{
    delete(cl.subscriptionByRequestID, req.ID)
    return nil, fmt.Errorf("unable to write request: %w", err)
  }
  
  return sub, nil
}

func (cl *Client) unsubscribe(subID uint64, unsubMethod string) error{
  req := newRequest([]interface{}{subID}, unsubMethod, nil, cl.shortID) // make a request for unsub method
  data, err := req.encode() // serializes the request into bytes
  if err != nil{
    return fmt.Errorf("unable to decode unsubscription msg for subID: %d and method %s", subID, unsubMethod)
  }
  cl.conn.SetWriteDeadline(time.Now().Add(writeWait))
  err = cl.conn.WriteMessage(websocket.TextMessage, data) // send ping to the server (sending a message, asking to unsub)
  if err != nil{
    return fmt.Errorf("unable to send unsubscription msg for subID: %d and method %s", subID, unsubMethod)
  }
  return nil
}
 
func decodeResponseFromMessage(msg []byte, reply interface{}) (err error){
  var resp *response
  if err := json.Unmarshal(msg, &resp); err != nil{
    return err
  }
  
  if resp.Error != nil{ // error on the response body side
    jsonErr := &json2.Error{}
    if err := json.Unmarshal(*resp.Error, jsonErr); err != nil{ 
      return &json2.Error{
        Code: json2.E_SERVER,
        Message: string(*resp.Error),
      }
    }
    return jsonErr
  }
  
  if resp.Params == nil{
    return json2.ErrNullResult // if response is good, but has no params (no result)
  }
   
  return json.Unmarshal(*resp.Params.Result, &reply)
}
 
func ConnectWithOptions(ctx context.Context, rpcEndpoint string, opts *Options) (client *Client, err error){
  client = &Client{
    rpcURL: rpcEndpoint,
    subscriptionByRequestID: map[uint64]*Subscription{},
    subscriptionByWSSubID:   map[uint64]*Subscription{},
  }

  // Customize how the client connects to the server
  dialer := &websocket.Dialer{ 
    Proxy:            http.ProxyFromEnvironment, // <-- func(*http.Request) (*url.URL, error)
    HandshakeTimeout: DefaultHandshakeTimeout, // <-- 45 secs wait max to prevent handing during connection
    EnableCompression: true,
 } 
  
 if opts != nil && opts.ShortID{
   client.shortID = opts.ShortID
 }

 if opts != nil && opts.HandshakeTimeout > 0{
   dialer.HandshakeTimeout = opts.HandshakeTimeout
 }

 var httpHeader http.Header = nil
 if opts != nil && opts.HttpHeader != nil && len(opts.HttpHeader) > 0{
   httpHeader = opts.HttpHeader
 }

 var resp *http.Response // to hold the http resp from DialContext
// Makes a connection to a websocket. httpHeader (optinal) sent along with handshake request
// Context to limit the dialing duration
 client.conn, resp, err = dialer.DialContext(ctx, rpcEndpoint, httpHeader)

 if err != nil{
   if resp != nil{
     body, _ := io.ReadAll(resp.Body)
     err = fmt.Errorf("new ws client: dial: %w, status: %s, body: %q", err, resp.Status, string(body))
   } else{
     err = fmt.Errorf("new ws client: dial: %w", err)
   }
  return nil, err
 }

// connCtxCancel function to cancel connCtx
 client.connCtx, client.connCtxCancel = context.WithCancel(context.Background())
 go func(){
   client.conn.SetReadDeadline(time.Now().Add(pongWait)) // <-- the client expects to receive a next Pong from server within that time, else timeout
  // sent by a server in response to the Ping message sent by the client
  // the function is envoked whenever a Pong message is received
   client.conn.SetPongHandler(func(string) error {client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil})
   ticker := time.NewTicker(pingPeriod)
   for{
     select{
      case <-client.connCtx.Done():
        return
      case <-ticker.C:
        client.sendPing()
     }
   }
  }()
  
  go client.receiveMessages()
  return client, nil
}
 
func (cl *Client) sendPing(){
  cl.lock.Lock()
  defer cl.lock.Unlock()
  
  //fmt.Println("Ping sent")
  cl.conn.SetWriteDeadline(time.Now().Add(writeWait))
  if err := cl.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil{
    return
  }
}
 
func (cl *Client) receiveMessages(){
  for{
    select{
    case <-cl.connCtx.Done():
      return
    default:
      _, msgInBytes, err := cl.conn.ReadMessage()
      if err != nil{
        cl.closeAllSubscription(err)
        return
      }
      
    cl.handleMessages(msgInBytes)
    }
  }
}

func (cl *Client) closeAllSubscription(err error){
  cl.lock.Lock()
  defer cl.lock.Unlock()
  
  for _, sub := range cl.subscriptionByRequestID{
    sub.err <- err
  }
  
  cl.subscriptionByRequestID = map[uint64]*Subscription{}
  cl.subscriptionByWSSubID = map[uint64]*Subscription{}
}

 
func getUint64WithOk(data []byte, path ...string) (uint64, bool){
  res, err := getUint64(data, path...)
  if err == nil{
    //fmt.Println("got through")
    return res, true
  }
  //fmt.Println("didnt get through")
  return 0, false
}
 
func getUint64(data []byte, keys ...string) (val uint64, err error){
  v, val_type, _, err := jsonparser.Get(data, keys...) 
  //fmt.Printf("Value(%T): %v, ValueType(%T): %v Error: %v\n",v,v, val_type,val_type, err)
  if err != nil{
    return 0, err
  }
  if val_type != jsonparser.Number{
    return 0, fmt.Errorf("The value isn't a number: %s", string(v))
  }
  //x,err := strconv.ParseUint(string(v), 10, 64)
  //fmt.Println(x,err)
  return strconv.ParseUint(string(v), 10, 64)  // <-- converts string to uint, base 10 (decimal), 64 bit
}

/*
  When receiving a mesage with `id`, the `result` will be a subscription number e.g:
  { "jsonrpc": "2.0", "result": 23784, "id": 1 } <-- subID is `result` and reqId is `id`
  This number `23784` will be assosiated with all futures messages that go to this request
*/
func (cl *Client) handleMessages(message []byte){
  requestID, ok := getUint64WithOk(message, "id")
  // New subscription case
  if ok{
    subID, _ := getUint64WithOk(message, "result")
    //fmt.Println(subID)
    cl.handleNewSubscriptionMessage(requestID, subID)
  }
  
  subID, _ := getUint64WithOk(message, "params", "subscription") // jsonparser will get the subscription number nested inside the params
  cl.handleSubscriptionMessage(subID, message)
}
 
func (cl *Client) handleSubscriptionMessage(subID uint64, message []byte){
  if traceEnabled{
    zlog.Debug("received a subscrption message", zap.Uint64("subscription_id", subID))
  }
  
  cl.lock.RLock()
  sub, found := cl.subscriptionByWSSubID[subID]
  cl.lock.RUnlock()
  if !found{
    zlog.Warn("unable to find subsscription for ws message", zap.Uint64("subscription_id", subID))
    return
  }
  
  //Decode the subscription using the decoderFunc
  result, err := sub.decoderFunc(message)
  if err != nil{
    cl.closeSubscription(sub.req.ID, fmt.Errorf("Unable to decode client's response"))
    return
  }
  
  // Check if the channel is full
  if len(sub.stream) >= cap(sub.stream){
    zlog.Warn("closing ws client subscription... not consuming fast enought", zap.Uint64("request_id", sub.req.ID))
    cl.closeSubscription(sub.req.ID, fmt.Errorf("reached channel's max cap %d", len(sub.stream)))
    return 
  }
  
  if !sub.closed{
    sub.stream <- result
  }
  
  return
}

func (cl *Client) handleNewSubscriptionMessage(requestID, subID uint64){
  cl.lock.Lock()
  defer cl.lock.Unlock()
  
  if traceEnabled {
		zlog.Debug("received new subscription message",
			zap.Uint64("message_id", requestID),
			zap.Uint64("subscription_id", subID),
		)
  }
    
  callBack, found := cl.subscriptionByRequestID[requestID] // returns *Subscription
  //fmt.Println(found)
  if !found{
    zlog.Error("cannot find websocket message handler for a new stream.... this should not happen",
    zap.Uint64("request_id", requestID), zap.Uint64("subscription_id", subID))
    return
  }

  callBack.subID = subID
  cl.subscriptionByWSSubID[subID] = callBack
 
  zlog.Debug("registered ws subscription",
		zap.Uint64("subscription_id", subID),
		zap.Uint64("request_id", requestID),
		zap.Int("subscription_count", len(cl.subscriptionByWSSubID)),
	)  
  return
}

