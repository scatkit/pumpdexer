package main

import (
	"testing"
  "context"
  "fmt"
  "log"
  "time"

  "github.com/scatkit/pumpdexer/rpc"
  "github.com/scatkit/pumpdexer/solana"
  "github.com/scatkit/pumpdexer/programs/system" 
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
  "github.com/davecgh/go-spew/spew"
)

type testTransactionInstructions struct {
	accounts  []*solana.AccountMeta
	data      []byte
	programID solana.PublicKey
}

func (t *testTransactionInstructions) Accounts() []*solana.AccountMeta {
	return t.accounts
}

func (t *testTransactionInstructions) ProgramID() solana.PublicKey {
	return t.programID
}

func (t *testTransactionInstructions) Data() ([]byte, error) {
	return t.data, nil
}

func TestNewTrans(t *testing.T) {
  client := rpc.New("https://api.mainnet-beta.solana.com")
	//DebugNewTransaction = true

	instructions := []solana.Instruction{
		&testTransactionInstructions{
			accounts: []*solana.AccountMeta{
				{PublicKey: solana.MustPubkeyFromBase58("A9QnpgfhCkmiBSjgBuWk76Wo3HxzxvDopUq9x6UUMmjn"), IsSigner: true, IsWritable: false},
				{PublicKey: solana.MustPubkeyFromBase58("9hFtYBYmBJCVguRYs9pBTWKYAFoKfjYR7zBPpEkVsmD"), IsSigner: true, IsWritable: true},
			},
			data:      []byte{0xaa, 0xbb},
			programID: solana.MustPubkeyFromBase58("11111111111111111111111111111111"),
		},
		&testTransactionInstructions{
			accounts: []*solana.AccountMeta{
				{PublicKey: solana.MustPubkeyFromBase58("SysvarC1ock11111111111111111111111111111111"), IsSigner: false, IsWritable: false},
				{PublicKey: solana.MustPubkeyFromBase58("SysvarS1otHashes111111111111111111111111111"), IsSigner: false, IsWritable: true},
				{PublicKey: solana.MustPubkeyFromBase58("9hFtYBYmBJCVguRYs9pBTWKYAFoKfjYR7zBPpEkVsmD"), IsSigner: false, IsWritable: true},
				{PublicKey: solana.MustPubkeyFromBase58("6FzXPEhCJoBx7Zw3SN9qhekHemd6E2b8kVguitmVAngW"), IsSigner: true, IsWritable: false},
			},
			data:      []byte{0xcc, 0xdd},
			programID: solana.MustPubkeyFromBase58("Vote111111111111111111111111111111111111111"),
		},
	}

  // Max age 
	block, err := client.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	require.NoError(t, err)
  
  //spew.Dump(block)

	trx, err := solana.NewTransaction(instructions, block.Value.Blockhash)
  
  spew.Dump(trx)
	require.NoError(t, err)

	assert.Equal(t, trx.Message.Header, solana.MessageHeader{
		NumRequiredSignatures:       3,
		NumReadonlySignedAccounts:   1,
		NumReadonlyUnsignedAccounts: 3,
	})

	assert.Equal(t, trx.Message.RecentBlockhash, block.Value.Blockhash)

	assert.Equal(t, trx.Message.Instructions, []solana.CompiledInstruction{
		{
			ProgramIDIndex: 5,
			Accounts:       []uint16{0, 0o1},
			Data:           []byte{0xaa, 0xbb},
		},
		{
			ProgramIDIndex: 6,
			Accounts:       []uint16{4, 3, 1, 2},
			Data:           []byte{0xcc, 0xdd},
		},
	})
}

func Test_SendTx (t *testing.T){
  client := rpc.New("https://api.devnet.solana.com")
  
  from := solana.PrivateKey{3,52,143,118,105,193,55,74,135,92,111,174,117,94,195,188,191,94,40,79,110,128,9,93,179,214,132,72,186,242,151,135,24,244,66,96,203,170,226,128,236,244,158,231,9,75,5,195,204,250,27,248,27,146,90,33,157,163,116,188,75,78,72,195}
  to := solana.MustPubkeyFromBase58("CJMTJWF97jd3dspsN5qhPp4EpKBHMTnkRvDkpSHUWSGJ")
  
  ctx, cancel := context.WithTimeout(context.Background(), time.Second*8)
  defer cancel()
  
  block, err := client.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
  if err != nil{
    log.Fatal(err)
  }
  
  tx, err := solana.NewTransaction(
    []solana.Instruction{
      system.NewTransferInstruction(
        14000,
        from.PublicKey(),
        to,
      ).Build(),
    },
    block.Value.Blockhash,
    solana.TransactionPayer(from.PublicKey()),
  )
  
  if err != nil{
    log.Fatal(err)
  }
 
  _,err = tx.Sign(
    func(pb solana.PublicKey) *solana.PrivateKey{
      if from.PublicKey().Equals(pb){
        return &from
      }
      return nil
    },
  )
  
  if err != nil{
    log.Fatal(err)
  }

  sig, err := client.SendTransaction(ctx, tx)
  
  if err != nil{
    log.Fatal(err)
  }
  fmt.Println(sig)
}
  
