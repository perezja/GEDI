package crypto

import(

  "fmt"
  "log"
  "errors"
  "reflect"
  "io/ioutil"
  "encoding/hex"
  "crypto/rsa"
  "crypto/x509"
  "golang.org/x/crypto/ssh"
  sha "crypto/sha1"
  mh "github.com/multiformats/go-multihash"
  // will eventually move to more secure
 // "encoding/pem"

)

const(
  sha1 = "sha1"

  // MinRSAKeyBits = 2048
  // The number of bytes in the public key 
  //  hash digest used to derive the node ID
)

type HashType uint64

func(ht HashType) IsValid() error {
  switch ht {
    case mh.SHA1:
      return nil
  }
  err := fmt.Sprintf("HashType.IsValid: %d is an unsupported hash type", ht)
  return errors.New(err)
}

func (ht HashType) Sum(bytes []byte) ([]byte, error) {
  switch ht {
    case mh.SHA1:
      digest := sha.Sum(bytes)
      return digest[:], nil
  }

  err := errors.New(fmt.Sprintf("HashType.Sum: unrecognized hash type %d", ht))
  return nil, err
}

type HashFunction interface {
  Sum() []byte
}

func unmarshallRSAPubKey(rsaPublicKeyPath string) (ssh.PublicKey, error){

  fmt.Println("Reading SSH Public key: ", rsaPublicKeyPath)
	authPublicKey, err := ioutil.ReadFile(rsaPublicKeyPath)
  //fmt.Println(authPublicKey)

  pk, _, _, _, _ := ssh.ParseAuthorizedKey(authPublicKey)
  if err != nil {
    log.Fatal(err)
	}

  pk, ok := pk.(ssh.PublicKey)
  if !ok { log.Fatal("Parsed key (public) is not RSA")}

	return pk.(ssh.PublicKey), nil

}

func unmarshallRSAPrivateKey(rsaPrivateKeyPath string) (*rsa.PrivateKey, error){

	pemPrivateKey, err := ioutil.ReadFile(rsaPrivateKeyPath)
  if err != nil {
    log.Fatal(err)
  }

  pk, err := ssh.ParseRawPrivateKey(pemPrivateKey)
    if err != nil {
					log.Fatal("Unable to parse RSA private key")
	}

  pk, ok := pk.(*rsa.PrivateKey)
  if !ok { log.Fatal("Parsed key is not RSA")}

	return pk.(*rsa.PrivateKey), nil
}

// This needs to be implemented into a data structure similar to what libp2p 
// uses to store both keys and perform signing for authentication
func marshalKeys(ht HashType, rsaPrivateKeyPath string) (string){
  pk, err := unmarshallRSAPrivateKey(rsaPrivateKeyPath)

  // remarshal into PKCS1,ASN.1 DER form
  pkBytes:= x509.MarshalPKCS1PrivateKey(pk)

  pkDigest, _ := ht.Sum(pkBytes)

  // slice the byte arary retured by sha1.Sum() 
  mhDigest, _ := mh.EncodeName(pkDigest[:], string(ht))
  pkHexString := hex.EncodeToString(mhDigest)

  if err != nil {
    log.Fatal(err)
  }

  return(pkHexString)
}

// hashType is a multihash encoding constant that maps to an identifiable
func IDFromRSAPubKey(ht HashType, rsaPubKeyPath string) (mh.Multihash, error){
  pk, err := unmarshallRSAPubKey(rsaPubKeyPath)

  if err != nil {
    return nil, err
  }

  // remarshal into PKCS1,ASN.1 DER form
  pkBytes := pk.Marshal()

  // take hash of RSA key
  if err := ht.IsValid(); err != nil {
    return nil, err
  }

  digest, _ := ht.Sum(pkBytes)

  // slice the byte arary retured by sha1.Sum() 
  // Use of reflect to get custom type into uint64 for encoding function call
  ref := reflect.ValueOf(ht)
  mhash, err := mh.Encode(digest[:], ref.Uint())
  if err != nil {
    return nil, err
  }
  //pkHexString := hex.EncodeToString(mhDigest)

 return mhash, nil
}
