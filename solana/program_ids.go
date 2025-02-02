package solana

var(
	// Create new accounts, allocate account data, assign accounts to owning programs,
	// transfer lamports from System Program owned accounts and pay transacation fees.
	SystemProgramID = MustPubkeyFromBase58("11111111111111111111111111111111")
)

var(
  // The Mint for native SOL Token accounts
	SolMint    = MustPubkeyFromBase58("So11111111111111111111111111111111111111112")
	WrappedSol = SolMint
)

var(
  // A Token program on the Solana blockchain.
  // This program defines a common implementation for Fungible and Non Fungible tokens.
  TokenProgramID = MustPubkeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
  // This program defines the convention and provides the mechanism for mapping
	// the user's wallet address to the associated token accounts they hold.
	SPLAssociatedTokenAccountProgramID = MustPubkeyFromBase58("ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL")

)


