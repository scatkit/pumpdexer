package ws

type Subscription struct{
  req *request 
  subID uint64
  stream chan result // channel that accepts the result (interface)
  err    chan error
  closeFunc func(err error)
  closed    bool
  unsubscribeMethod string
  decoderFunc  decoderFunc
}

type decoderFunc func([]byte) (interface{}, error)

func newSubscription(req *request, closeFunc func(err error), unsubMethod string, decoderFunc decoderFunc,
) *Subscription{
  
  return &Subscription{
    req:      req,
    subID:    0,
    stream:   make(chan result, 200_000),
    err:      make(chan error, 100_000),
    closeFunc:  closeFunc,
    unsubscribeMethod: unsubMethod,
    decoderFunc: decoderFunc,
  }
}


func (s *Subscription) Unsubscribe(){
  s.unsubscribe(nil)
}

func (s *Subscription) unsubscribe(err error){
  s.closeFunc(err) 
  s.closed = true 
  close(s.stream)
  close(s.err)
}
