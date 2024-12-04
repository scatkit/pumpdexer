package solana

import (
  "fmt"
  "encoding/json"
  //"errors"
  //"crypto/ed25519"
  "github.com/mr-tron/base58"
  //"filippo.io/edwards25519" 
)
 
// check if the provided `b` is on the ed25519 curve
//func IsOnCurve(b []byte) bool{
//  if len(b) != ed25519.PublicKeySize{
//    return false
//  }
//  _, err := new(edwards25519.Point).SetBytes(b)
//  isOnCurve := err == nil
//  return isOnCurve
//}
 
//func ValidatePublicKey(b []byte) (bool, error){
//  if len(b) != ed25519.PublicKeySize{
//    return false, fmt.Errorf("Invalid private key size, expected %v, got %d", ed25519.PrivateKeySize, len(b))
//  }
//  pub := ed25519.PrivateKey(b).Public().(ed25519.PublicKey)
//  if !IsOnCurve(pub){
//    return false, errors.New("the corresponding public key is NOT on the ed25519 curve")
//  }
//  return true, nil
//}
type PublicKey [PublicKeyLength]byte

func (key *PublicKey) UnmarshalText(data []byte) error{
  return key.Set(string(data))
}

func (key *PublicKey) Set(s string) (err error){
  *key, err = PublicKeyFromBase58(s)
  if err != nil{
    return fmt.Errorf("invalid public key %s: %w", s, err)
  }
  return
}

func (key PublicKey) MarshalJSON() ([]byte, error){
  return json.Marshal(base58.Encode(key[:]))
}

func (key *PublicKey) UnmarshalJSON(data []byte) (err error){
  var s string
  if err := json.Unmarshal(data, &s); err != nil{
    return err
  }
  
  *key, err = PublicKeyFromBase58(s)
  if err != nil{
    return fmt.Errorf("invalid public key %q: %w", s, err)
  }
  return 
}

func (key *PublicKey) String() string{
  return base58.Encode(key[:])
}

func (key *PublicKey) Bytes() []byte{
  return []byte(key[:])
}

func PublicKeyFromBase58(pubkey string) (out PublicKey, err error){
  res, err := base58.Decode(pubkey)
  if err != nil{
    return out, fmt.Errorf("deocde: %w",err)
  }
  if len(res) != PublicKeyLength{
    return out, fmt.Errorf("invalid length, expected %v got %d", PublicKeyLength, len(res))
  }
  
  copy(out[:],res) // <-- out is 32bit array
  return out, nil
}

func MustPubkeyFromBase58(pubkey_string string) PublicKey{
  out, err := PublicKeyFromBase58(pubkey_string)
  if err != nil{
    panic(err)
  }
  return out
}

type PrivateKey []byte
const (
  PublicKeyLength = 32 
  PrivateKeyLength = 64
  
  // max length of derived pubkey seed
  MaxSeedLength = 32
  
  // max number of seeds
  MaxSeeds = 16

  // num of bytes in the signature
  SignatureLength = 64
)

