package solana

import (
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

//var traceEnabled = logging.IsTraceEnabled("solana-go", "github.com/gagliardetto/solana-go")
var zlog, tracer = logging.PackageLogger("solana", "github.com/scatkit/pumpdexer/solana")

func init() {
  var err error 
  zlog, err = zap.NewDevelopment() 
  if err != nil{
    panic(err)
  }
	//logging.Register("github.com/gagliardetto/solana-go", &zlog)
}
