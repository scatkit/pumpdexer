package rpc

import (
  "context"
  //"fmt"
  "testing"
  stdjson "encoding/json"
  "math/big"
  "go_projects/solana"
  //"github.com/davecgh/go-spew/spew"
  "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustJSONToInterface(rawData []byte) interface{}{
  var out interface{}
  err := stdjson.Unmarshal(rawData, &out)
  if err != nil{
    panic(err)
  }
  return out
}

func mustAnyToJSON(raw interface{}) []byte{
  out, err := stdjson.Marshal(raw)
  if err != nil{
    panic(err)
  }
  return out
}

func wrapIntoRPC(body string) string{
  return `{"jsonrpc": "2.0","result":` + body + `,"id":0}`
}
func TestClient_GetAccountInfo(t *testing.T) {
	responseBody := `{"context":{"slot":83986105},"value":{"data":["dGVzdA==","base64"],"executable":true,"lamports":999999,"owner":"11111111111111111111111111111111","rentEpoch":18446744073709551615}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPubkeyFromBase58(pubkeyString)
	out, err := client.GetAccountInfo(context.Background(), pubKey)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]interface{}{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getAccountInfo",
			"params": []interface{}{
				pubkeyString,
				map[string]interface{}{
					"encoding": "base64",
				},
			},
		},
		reqBody,
	)

	rentEpoch, _ := new(big.Int).SetString("18446744073709551615", 10)
	assert.Equal(t,
		&AccountInfoResult{
			RPCContext: RPCContext{
				Context{Slot: 83986105},
			},
			Value: &Account{
				Lamports: 999999,
				Owner:    solana.MustPubkeyFromBase58("11111111111111111111111111111111"),
				Data: &DataBytesOrJSON{
					rawDataEncoding: solana.EncodingBase64,
					asDecodedBinary: solana.Data{
						Content:  []byte{0x74, 0x65, 0x73, 0x74},
						Encoding: solana.EncodingBase64,
					},
				},
				Executable: true,
				RentEpoch:  rentEpoch,
			},
		}, out)
}

func TestClient_GetAccountInfoWithOpts(t *testing.T) {
	responseBody := `{"context":{"slot":83986105},"value":{"data":["dGVzdA==","base64"],"executable":true,"lamports":999999,"owner":"11111111111111111111111111111111","rentEpoch":207}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	offset := uint64(22)
	length := uint64(33)
	minContextSlot := uint64(123456)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPubkeyFromBase58(pubkeyString)

	opts := &GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: CommitmentFinalized,
		DataSlice: &DataSlice{
			Offset: &offset,
			Length: &length,
		},
		MinContextSlot: &minContextSlot,
	}
	_, err := client.GetAccountInfoWithOpts(
		context.Background(),
		pubKey,
		opts,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]interface{}{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getAccountInfo",
			"params": []interface{}{
				pubkeyString,
				map[string]interface{}{
					"encoding":   string(solana.EncodingBase64),
					"commitment": string(CommitmentFinalized),
					"dataSlice": map[string]interface{}{
						"offset": float64(offset),
						"length": float64(length),
					},
					"minContextSlot": float64(minContextSlot),
				},
			},
		},
		reqBody,
	)
}

func TestClient_GetTokenSupply(t *testing.T){
  responseBody := `{"context":{"slot": 1114},"value":{"amount":"100000","decimals":2,"uiAmount":1000,"uiAmountString":"1000"}}`
  server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
  defer closer()
 
  client := New(server.URL)
  pubKey := solana.MustPubkeyFromBase58("D27DgiipBR5dRdij2L6NQ27xwyiLK5Q2DsEM5ML5EuLK")
  
  out, err := client.GetTokenSupply(context.Background(), pubKey, CommitmentFinalized)
  require.NoError(t, err)
  
  reqBody := server.RequestBody(t)
  assert.NotNil(t, reqBody["id"])
  reqBody["id"] = any(nil)

  assert.Equal(t,
    map[string]interface{}{
      "jsonrpc": "2.0", "id": any(nil),
      "method": "getTokenSupply",
      "params": []interface{}{
        pubKey.String(),
        map[string]interface{}{
          "commitment": string(CommitmentFinalized),
        },
      },

    }, reqBody,
  )
  
  expected := mustJSONToInterface([]byte(responseBody))
  got := mustJSONToInterface(mustAnyToJSON(out))
  assert.Equal(t, expected, got, "both deserialized value must be equal")
}

func TestClient_GetBalance(t *testing.T) {
	responseBody := `{"context":{"slot":83987501},"value":19039980000}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()

	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPubkeyFromBase58(pubkeyString)
	out, err := client.GetBalance(
		context.Background(),
		pubKey,
		CommitmentFinalized,
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]interface{}{
      "jsonrpc": "2.0",
			"id":      any(nil),
			"method":  "getBalance",
			"params": []interface{}{
				pubkeyString,
				map[string]interface{}{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	assert.Equal(t,
		&GetBalanceResult{
			RPCContext: RPCContext{
				Context{Slot: 83987501},
			},
			Value: 19039980000,
		}, out)
}

func TestClient_GetTokenAccountBalance(t *testing.T) {
	responseBody := `{"context":{"slot":1114},"value":{"amount":"9864","decimals":2,"uiAmount":98.64,"uiAmountString":"98.64"}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPubkeyFromBase58(pubkeyString)

	out, err := client.GetTokenAccountBalance(
		context.Background(),
		pubKey,
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]interface{}{
      "jsonrpc": "2.0",
			"id":      any(nil),
			"method":  "getTokenAccountBalance",
			"params": []interface{}{
				pubkeyString,
				map[string]interface{}{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetBlockTime(t *testing.T) {
	responseBody := `1625230849`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	block := 55
	out, err := client.GetBlockTime(
		context.Background(),
		uint64(block),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]interface{}{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getBlockTime",
			"params": []interface{}{
				float64(block),
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))
	got := mustJSONToInterface(mustAnyToJSON(out))
	assert.Equal(t, expected, got, "both deserialized values must be equal")
}
