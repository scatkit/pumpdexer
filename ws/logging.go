package ws

import (
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

var zlog *zap.Logger
var traceEnabled = logging.IsTraceEnabled("pumpdexer", "github.com/scatkit/pumpdexer/ws")

func init() {
	logging.Register("github.com/scatkit/pumpdexer/ws", &zlog)
}
