package ws

import (
	"context"
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/scatkit/pumpdexer/rpc"
	"github.com/scatkit/pumpdexer/solana"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	//"github.com/davecgh/go-spew/spew"
)

func Test_AccountSubscribe(t *testing.T) {
	t.Skip("Skipped accSub")
	zlog, _ = zap.NewDevelopment()
	co, err := ConnectWithOptions(context.Background(), "ws://api.mainnet-beta.solana.com:80", nil)
	defer co.Close()
	require.NoError(t, err)

	poolID := solana.MustPubkeyFromBase58("GJVvpsKp5snWU4VVZa71yeNnLCRckecooqEhiyRSszMC")
	sub, err := co.AccountSubscribeWithOpts(poolID, "", "")
	require.NoError(t, err)

	defer sub.Unsubscribe()
	for {
		data, err := sub.Recv(context.Background())
		if err != nil {
			fmt.Println("received error", err)
			return
		}
		spew.Dump(data)
	}
}

func Test_LogSubscribe(t *testing.T) {
	//zlog, _ := zap.NewDevelopment()
	liqPool := solana.MustPubkeyFromBase58("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8")

	//tokenAddr := solana.MustPubkeyFromBase58("HwdwHMkaTkYREmWwQDGNgbtAou43nE9ACgdxkW2Wpump")
	co, err := Connect(context.Background(), "ws://api.mainnet-beta.solana.com:80")
	defer co.Close()
	require.NoError(t, err)

	sub, err := co.LogSubscribeToAddress(liqPool, rpc.CommitmentFinalized)
	require.NoError(t, err)

	defer sub.Unsubscribe()
	for {
		got, e := sub.Recv(context.Background())
		if e != nil {
			fmt.Println("Error:", e)
			return
		}
		if got.Value.Err != nil {
			fmt.Println("Unsuccessful transaction")
		} else {
			spew.Dump(got)
		}
	}
	//for {
	//  data, err := sub.Recv(context.Background())
	//  if err != nil{
	//    fmt.Println("Received an error:", err)
	//    return
	//  }
	//  fmt.Println("Data received:", data)
	//}
}
