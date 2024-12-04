package rpc

import(
  stdjson "encoding/json"
  "io"
  "net/http"
  "net/http/httptest"
  "testing"
     
  "github.com/stretchr/testify/require"
)
 
type mockJSONRPCServer struct{
  *httptest.Server
  body []byte
}
 
func mockJSONRPC(t *testing.T, response interface{}) (mock *mockJSONRPCServer, close func()){
  mock = &mockJSONRPCServer{
    Server: httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request){
      var err error
      mock.body, err = io.ReadAll(req.Body)
      require.NoError(t, err)
      
      var responseBody []byte
      if v, ok := response.(stdjson.RawMessage); ok{ 
        responseBody = v // if response is already a json RawMessage
      } else{
        responseBody, err = stdjson.Marshal(response)
        require.NoError(t, err)
      }

      rw.Write(responseBody) // <-- writes the the responseBody (either rawJson or jsonMarhal) to the body
    })),
  }
  
  return mock, func(){mock.Close()}
}

func (s *mockJSONRPCServer) RequestBodyAsJSON(t *testing.T) (out string){
  return string(s.body)  
}
 
func (s *mockJSONRPCServer) RequestBody(t *testing.T) (out map[string]interface{}){
  err := stdjson.Unmarshal(s.body, &out)
  require.NoError(t, err)
  
  return out
}
