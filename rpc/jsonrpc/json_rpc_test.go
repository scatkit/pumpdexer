package jsonrpc
import(
  "context"
  "fmt"
  "io"
  "net/http"
  "net/http/httptest"
  "os"
  "testing"
  //"github.com/davecgh/go-spew/spew"
  . "github.com/onsi/gomega"
)

//func TestExample(t *testing.T){
//  RegisterTestingT(t)
//  
//  Expect(10/2).To(Equal(5))
//}

var requestChan = make(chan *RequestData, 1)

type RequestData struct {
	request *http.Request
	body    string
}

var responseBody = ""
var httpServer *httptest.Server

// start the testhttp server and stop it when tests are finished
func TestMain(m *testing.M) {
	httpServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		// put request and body to channel for the client to investigate them
		requestChan <- &RequestData{r, string(data)}

		fmt.Fprintf(w, responseBody)
	}))
	defer httpServer.Close()

	os.Exit(m.Run())
}
func TestClientHeader(t *testing.T) {
  RegisterTestingT(t)
  
	rpcClient := newClient(httpServer.URL)
  rpcClient.Call(context.Background(),"random_method",1,2,3,4)
  req := (<-requestChan).request

  Expect(req.Method).To(Equal("POST"))
  Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))
  Expect(req.Header.Get("Accept")).To(Equal("application/json"))
}
 
func TestClientCall(t *testing.T){
  RegisterTestingT(t)
  rpcClient := newClient(httpServer.URL)
  
  
  rpcClient.Call(context.Background(), "emptyMethod")
  Expect((<-requestChan).body).To(Equal(`{"jsonrpc":"2.0","id":1,"method":"emptyMethod"}`))
  
  rpcClient.Call(context.Background(), "nullParam",nil)
  Expect((<-requestChan).body).To(Equal(`{"jsonrpc":"2.0","id":1,"method":"nullParam","params":[null]}`))
  
  rpcClient.Call(context.Background(), "doubleNullParam",nil,nil)
  Expect((<-requestChan).body).To(Equal(`{"jsonrpc":"2.0","id":1,"method":"doubleNullParam","params":[null,null]}`))
  
  rpcClient.Call(context.Background(), "doubleNullParam",nil,nil)
  Expect((<-requestChan).body).To(Equal(`{"jsonrpc":"2.0","id":1,"method":"doubleNullParam","params":[null,null]}`))
  
  rpcClient.Call(context.Background(), "interfaceSlice",[]interface{}{"hey",2})
  Expect((<-requestChan).body).To(Equal(`{"jsonrpc":"2.0","id":1,"method":"interfaceSlice","params":["hey",2]}`))
  
  rpcClient.Call(context.Background(), "solanaMethod","getAccountInfo", map[string]interface{}{"encoding":"base64"})
  Expect((<-requestChan).body).To(Equal(`{"jsonrpc":"2.0","id":1,"method":"solanaMethod","params":["getAccountInfo",{"encoding":"base64"}]}`))
}
