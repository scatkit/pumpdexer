package associatedtokenaccount
import(
  "errors"
  "fmt"
  
  bin "github.com/gagliardetto/binary"
  "github.com/scatkit/pumpdexer/solana"
)

type Create struct {
	Payer  solana.PublicKey `bin:"-" borsh_skip:"true"`
	Wallet solana.PublicKey `bin:"-" borsh_skip:"true"`
	Mint   solana.PublicKey `bin:"-" borsh_skip:"true"`

  // [0] = [WRITE, SIGNER] Payer: `Funding account`
	// [1] = [WRITE] AssociatedTokenAccount: `Associated token account address to be created`
	// [2] = [] Wallet: `Wallet address for the new associated token account`
	// [3] = [] TokenMint: `The token mint for the new associated token account`
	// [4] = [] SystemProgram: `System program ID`
  // [5] = [] TokenProgram: `SPL token program ID`
	// [6] = [] SysVarRent: `SysVarRentPubkey`
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst Create) Build() *Instruction{
  // find the associated token address. It's created from user's wallet + token's mint address
  associatedTokenAddress, _, _ := solana.FindAssociatedTokenAddress(inst.Wallet, inst.Mint)
  
  keys := []*solana.AccountMeta{
    {
      PublicKey:    inst.Payer,
      IsSigner:     true,
      IsWritable:   true,
    },
    {
      PublicKey:    associatedTokenAddress,
      IsSigner:     false,
      IsWritable:   true,
    },
    {
      PublicKey:    inst.Wallet,
      IsSigner:     false,
      IsWritable:   false,
    },
    { 
      PublicKey:    inst.Mint,
      IsSigner:     false,
      IsWritable:   false,
    },
    { 
      PublicKey:    solana.SystemProgramID,
      IsSigner:     false,
      IsWritable:   false,
    },
    { 
      PublicKey:    solana.TokenProgramID,
      IsSigner:     false,
      IsWritable:   false,
    },
    { 
      PublicKey:    solana.SysVarRentPubkey,
      IsSigner:     false,
      IsWritable:   false,
    },
  }
  
  inst.AccountMetaSlice = keys
  
  return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.NoTypeIDDefaultID,
	}}
}

// ValidateAndBuild validates the instruction accounts.
// If there is a validation error, return the error.
// Otherwise, build and return the instruction.
func (inst Create) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Create) Validate() error {
	if inst.Payer.IsZero() {
		return errors.New("Payer not set")
	}
	if inst.Wallet.IsZero() {
		return errors.New("Wallet not set")
	}
	if inst.Mint.IsZero() {
		return errors.New("Mint not set")
	}
	_, _, err := solana.FindAssociatedTokenAddress(
		inst.Wallet,
		inst.Mint,
	)
	if err != nil {
		return fmt.Errorf("error while FindAssociatedTokenAddress: %w", err)
	}
	return nil
}

func (inst Create) MarshalWithEncoder(encoder *bin.Encoder) error {
	return encoder.WriteBytes([]byte{}, false)
}

func (inst *Create) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	return nil
}


// NewCreateInstructionBuilder creates a new `Create` instruction builder.
func NewCreateInstructionBuilder() *Create {
	nd := &Create{}
	return nd
}

func (inst *Create) SetPayer(payer solana.PublicKey) *Create {
	inst.Payer = payer
	return inst
}

func (inst *Create) SetWallet(wallet solana.PublicKey) *Create {
	inst.Wallet = wallet
	return inst
}

func (inst *Create) SetMint(mint solana.PublicKey) *Create {
	inst.Mint = mint
	return inst
}

func NewCreateInstruction(
  payer solana.PublicKey,
  walletAddress solana.PublicKey,
  splTokenMintAddress solana.PublicKey,
) *Create {
  return NewCreateInstructionBuilder().
  SetPayer(payer).
  SetWallet(walletAddress).
  SetMint(splTokenMintAddress)
}


