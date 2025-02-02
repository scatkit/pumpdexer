package solana

import (
	"fmt"
  "math"
	//"encoding/json"
	"crypto"
	"crypto/ed25519"
	crypto_rand "crypto/rand"
	"errors"
  "crypto/sha256"

	"filippo.io/edwards25519"
	"github.com/mr-tron/base58"
)

type PublicKey [PublicKeyLength]byte
type PublicKeySlice []PublicKey

func (slice *PublicKeySlice) UniqueAppend(pubkey PublicKey) bool {
	if !slice.Has(pubkey) {
		slice.Append(pubkey)
		return true
	}
	return false
}

func (slice *PublicKeySlice) Append(pubkeys ...PublicKey) {
	*slice = append(*slice, pubkeys...)
}

func (slice PublicKeySlice) Has(pubkey PublicKey) bool {
	for _, key := range slice {
		if key.Equals(pubkey) {
			return true
		}
	}
	return false
}

func (key PublicKey) Equals(pb PublicKey) bool {
	return key == pb
}

func (key *PublicKey) UnmarshalText(data []byte) error {
	return key.Set(string(data))
}

func (key *PublicKey) Set(s string) (err error) {
	*key, err = PublicKeyFromBase58(s)
	if err != nil {
		return fmt.Errorf("invalid public key %s: %w", s, err)
	}
	return
}

var ZeroPublicKey = PublicKey{}

func (key PublicKey) IsZero() bool {
	return key == ZeroPublicKey
}

func (key PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(base58.Encode(key[:]))
}

func (key *PublicKey) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	*key, err = PublicKeyFromBase58(s)
	if err != nil {
		return fmt.Errorf("invalid public key %q: %w", s, err)
	}
	return
}

func (key PublicKey) String() string {
	return base58.Encode(key[:])
}

func (key *PublicKey) Bytes() []byte {
	return []byte(key[:])
}

// Create a public key from base58 encoded string
func PublicKeyFromBase58(pubkey string) (out PublicKey, err error) {
	res, err := base58.Decode(pubkey)
	if err != nil {
		return out, fmt.Errorf("deocde: %w", err)
	}
	if len(res) != PublicKeyLength {
		return out, fmt.Errorf("invalid length, expected %v got %d", PublicKeyLength, len(res))
	}

	copy(out[:], res) // <-- out is 32 bit array
	return out, nil
}

func MustPubkeyFromBase58(pubkey_string string) PublicKey {
	out, err := PublicKeyFromBase58(pubkey_string)
	if err != nil {
		panic(err)
	}
	return out
}

func PublicKeyFromBytes(b []byte) (pub PublicKey) {
	byteLength := len(b)
	if byteLength != PublicKeyLength {
		panic(fmt.Errorf("invalid public key size, expected %v, got %d", PublicKeyLength, byteLength))
	}
	copy(pub[:], b)
	return pub
}

type PrivateKey []byte

const (
	PublicKeyLength  = 32
	PrivateKeyLength = 64

	// max length of derived pubkey seed
	MaxSeedLength = 32

	// max number of seeds
	MaxSeeds = 16

	// num of bytes in the signature
	SignatureLength = 64
)

func (key PrivateKey) Validate() error {
	_, err := ValidatePrivateKey(key)
	return err
}

func ValidatePrivateKey(priv_bytes []byte) (bool, error) {
	if len(priv_bytes) != ed25519.PrivateKeySize {
		return false, fmt.Errorf("Invalid private key size. Expected: %d, got: %d", ed25519.PrivateKeySize, len(priv_bytes))
	}
	// check if the public key is on the ed25519 curve
	pub := ed25519.PrivateKey(priv_bytes).Public().(ed25519.PublicKey)
	if !IsOnCurve(pub) {
		return false, errors.New("the corresponding public key is NOT on the ed25519 curve")
	}
	return true, nil
}

func (key PrivateKey) Sign(payload []byte) (Signature, error) {
	if err := key.Validate(); err != nil {
		return Signature{}, err
	}
	p := ed25519.PrivateKey(key)
	signData, err := p.Sign(crypto_rand.Reader, payload, crypto.Hash(0))
	if err != nil {
		return Signature{}, err
	}

	var sig Signature
	copy(sig[:], signData)

	return sig, nil
}

func (key PrivateKey) String() string {
	return base58.Encode(key)
}

func IsOnCurve(b []byte) bool {
	if len(b) != ed25519.PublicKeySize {
		return false
	}
	_, err := new(edwards25519.Point).SetBytes(b)
	isOnCurve := err == nil
	return isOnCurve
}

func NewPrivateKey() (PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(crypto_rand.Reader)
	if err != nil {
		return nil, err
	}
	var publicKey PublicKey
	copy(publicKey[:], pub)
	return PrivateKey(priv), nil
}

func (key PrivateKey) PublicKey() PublicKey {
	if err := key.Validate(); err != nil {
		panic(err)
	}

	priv := ed25519.PrivateKey(key)
	pub := priv.Public().(ed25519.PublicKey)

	var publicKey PublicKey
	copy(publicKey[:], pub)

	return publicKey
}

//func NewRandomPrivateKey() (PrivateKey, error) {
//	pub, priv, err := ed25519.GenerateKey(crypto_rand.Reader)
//	if err != nil {
//		return nil, err
//	}
//	var publicKey PublicKey
//	copy(publicKey[:], pub)
//	return PrivateKey(priv), nil
//}

func PrivateKeyFromBase58(in string) (PrivateKey, error) {
	out, err := base58.Decode(in)
	if err != nil {
		return nil, err
	}
	if _, err := ValidatePrivateKey(out); err != nil {
		return nil, err
	}
	return out, nil
}

func MustPrivkeyFromBase58(in string) PrivateKey {
	out, err := PrivateKeyFromBase58(in)
	if err != nil {
		panic(err)
	}
	return out
}

const PDA_MARKER = "ProgramDerivedAddress"

// creates a program address
func CreateProgramAddress(seeds [][]byte, programID PublicKey) (PublicKey, error) {
	if len(seeds) > MaxSeeds { // max 16
		return PublicKey{}, errors.New("max length exceeded")
	}

	for _, seed := range seeds {
		if len(seed) > MaxSeedLength { // max seed is 32 bytes
			return PublicKey{}, errors.New("max length exceeded") 
		}
	}

	buf := []byte{}
	for _, seed := range seeds {
		buf = append(buf, seed...)
	}

	buf = append(buf, programID[:]...)
	buf = append(buf, []byte(PDA_MARKER)...)
	hash := sha256.Sum256(buf)

	if IsOnCurve(hash[:]) {
		return PublicKey{}, errors.New("invalid seeds; address must fall off the curve")
	}

	return PublicKeyFromBytes(hash[:]), nil
}

func FindProgramAddress(seed [][]byte, programID PublicKey) (PublicKey, uint8, error){
  var address PublicKey
  var err error
  // start with the Max seed
  bumpSeed := uint8(math.MaxUint8) 
  
  for bumpSeed != 0{
    address, err = CreateProgramAddress(append(seed, []byte{byte(bumpSeed)}), programID)
    if err == nil{
      return address, bumpSeed, nil
    }
    bumpSeed--
  }

  return PublicKey{}, bumpSeed, errors.New("unable to find valid program address")
}

// ATA is created from user's wallet + token mint's address
func FindAssociatedTokenAddress(wallet PublicKey, mint PublicKey,
) (PublicKey, uint8, error){
  return findAssociatedTokenAddressAndBumpSeed(wallet, mint, SPLAssociatedTokenAccountProgramID)
}

func findAssociatedTokenAddressAndBumpSeed(walletAddress PublicKey, splTokenMintAddress PublicKey,programID PublicKey,
) (PublicKey, uint8, error){
	return FindProgramAddress([][]byte{
		walletAddress[:],
		TokenProgramID[:],
		splTokenMintAddress[:], // <-- ATA program 
	},
		programID, // <-- this program defines a common implementation for Fungible and Non Fungible tokens.

	)
}
