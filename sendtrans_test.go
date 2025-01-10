package main

import (
	"testing"

  "context"
  "github.com/scatkit/pumpdexer/rpc"
  "github.com/scatkit/pumpdexer/solana"
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
	blockhash, err := client.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	require.NoError(t, err)
  
  //spew.Dump(blockhash)

	trx, err := solana.NewTransaction(instructions, blockhash.Value.Blockhash)
  
  spew.Dump(trx)
	require.NoError(t, err)

	assert.Equal(t, trx.Message.Header, solana.MessageHeader{
		NumRequiredSignatures:       3,
		NumReadonlySignedAccounts:   1,
		NumReadonlyUnsignedAccounts: 3,
	})

	assert.Equal(t, trx.Message.RecentBlockhash, blockhash.Value.Blockhash)

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
