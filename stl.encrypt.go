package stl

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
)

type XPEncryptImpl struct {

}

/**
* MD5
* @param str string 需要加密的字符串
* @param to_upper bool 返回类型 true大写 false小写
 */
func Md5String(str string, toUpper bool) string {
	if toUpper {
		return fmt.Sprintf("%X", md5.Sum([]byte(str)))
	} else {
		return fmt.Sprintf("%x", md5.Sum([]byte(str)))
	}
}

/*********************** Padding ********************/
//PKCS7填充，
func PKCS7Padding(cipher []byte, blockSize int) []byte {
	padding := blockSize - len(cipher) % blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipher, padText...)
}

//PKCS7反填充，
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}

//PKCS5填充，
func PKCS5Padding(cipher []byte, blockSize int) []byte {
	padding := blockSize - len(cipher)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipher, padText...)
}


//PKCS5反填充，
func PKCS5UnPadding(cipher []byte) []byte {
	length := len(cipher)
	// 去掉最后一个字节 unPadding 次
	unPadding := int(cipher[length-1])
	return cipher[:(length - unPadding)]
}

//0填充
func ZeroPadding(cipher []byte, blockSize int) []byte {
	padding := blockSize - len(cipher)%blockSize
	padText := bytes.Repeat([]byte{0}, padding)//用0去填充
	return append(cipher, padText...)
}

//反0填充
func ZeroUnPadding(origData []byte) []byte {
	return bytes.TrimFunc(origData,
		func(r rune) bool {
			return r == rune(0)
		})
}
/*********************** Padding ********************/

/*********************** RSA ********************/
const ENCRYPT_KEYSIZE, DECRYPT_KEYSIZE = 117, 128 //1024

var hashPrefixes = map[crypto.Hash][]byte{
	crypto.MD5:       {0x30, 0x20, 0x30, 0x0c, 0x06, 0x08, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x02, 0x05, 0x05, 0x00, 0x04, 0x10},
	crypto.SHA1:      {0x30, 0x21, 0x30, 0x09, 0x06, 0x05, 0x2b, 0x0e, 0x03, 0x02, 0x1a, 0x05, 0x00, 0x04, 0x14},
	crypto.SHA224:    {0x30, 0x2d, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x04, 0x05, 0x00, 0x04, 0x1c},
	crypto.SHA256:    {0x30, 0x31, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x01, 0x05, 0x00, 0x04, 0x20},
	crypto.SHA384:    {0x30, 0x41, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x02, 0x05, 0x00, 0x04, 0x30},
	crypto.SHA512:    {0x30, 0x51, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x03, 0x05, 0x00, 0x04, 0x40},
	crypto.MD5SHA1:   {}, // A special TLS case which doesn't use an ASN1 prefix.
	crypto.RIPEMD160: {0x30, 0x20, 0x30, 0x08, 0x06, 0x06, 0x28, 0xcf, 0x06, 0x03, 0x00, 0x31, 0x04, 0x14},
}

//截取byte
func subString(array []byte, begin int, length int) []byte {
	lth := len(array)
	// 简单的越界判断
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length
	if end > lth {
		end = lth
	}

	// 返回子串
	return array[begin:end]
}

//多个[]byte数组合并成一个[]byte
func bytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

/**
 * 获取公钥
 * @param file_path string 公钥路径 ex: ./conf/keys/public_key.pem
 */
func GetPublicKey(filePath string) (*rsa.PublicKey, error){
	var publicKey *rsa.PublicKey
	pubData, err := ioutil.ReadFile(filePath)
	//解析证书
	block, _ := pem.Decode([]byte(pubData))
	cert, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Println(err.Error())
		return publicKey, err
	}
	publicKey = cert.(*rsa.PublicKey)
	return publicKey, nil
}

/**
 * 获取私钥
 * @param file_path string 私钥路径 ex: ./conf/keys/private_key.pem
 */
func GetPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	pfxData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	//解析证书
	block, _ := pem.Decode(pfxData)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println("解析证书错误" + err.Error())
		return nil, err
	}
	return privateKey, nil
}

/**
 * 公钥加密
 */
func PublicEncrypt(pub *rsa.PublicKey, data []byte) ([]byte, error) {
	signData, err := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err != nil {
		return nil, err
	}
	return []byte(base64.StdEncoding.EncodeToString(signData)), nil
}

/**
 * 公钥分段加密
 */
func PublicPartEncrypt(pub *rsa.PublicKey, data []byte) ([]byte, error) {
	length := len(data)
	i := 0
	var finalData []byte
	for i < length {
		signData, err := rsa.EncryptPKCS1v15(rand.Reader, pub, subString(data, i, ENCRYPT_KEYSIZE))
		if err != nil {
			return nil, err
		}
		finalData = bytesCombine(finalData, signData)
		i += ENCRYPT_KEYSIZE
	}
	return []byte(base64.StdEncoding.EncodeToString(finalData)), nil
}

/**
 * 公钥验签
 */
func PublicVerifySign(pub *rsa.PublicKey, src []byte, sign []byte) error {
	h := sha1.New()
	h.Write(src)
	hashed := h.Sum(nil)

	return rsa.VerifyPKCS1v15(pub, crypto.SHA1, hashed, sign)
}

func pkcs1v15HashInfo(hash crypto.Hash, inLen int) (hashLen int, prefix []byte, err error) {
	hashLen = hash.Size()
	if inLen != hashLen {
		return 0, nil, errors.New("mrsa: input must be hashed message")
	}
	prefix, ok := hashPrefixes[hash]
	if !ok {
		return 0, nil, errors.New("mrsa: unsupported hash function")
	}
	return
}

func publicDecrypt(pub *rsa.PublicKey, data []byte) (out []byte, err error) {
	hashLen, prefix, err := pkcs1v15HashInfo(crypto.Hash(0), 0)
	if err != nil {
		return nil, err
	}

	tLen := len(prefix) + hashLen
	k := (pub.N.BitLen() + 7) / 8
	if k < tLen + 11 {
		return nil, fmt.Errorf("length illegal")
	}
	c := new(big.Int).SetBytes(data)
	m := encrypt(new(big.Int), pub, c)
	em := leftPad(m.Bytes(), k)

	out = unLeftPad(em)
	err = nil
	return
}

// copy from crypt/rsa/pkcs1v5.go
func leftPad(input []byte, size int) (out []byte) {
	n := len(input)
	if n > size {
		n = size
	}
	out = make([]byte, size)
	copy(out[len(out) - n:], input)
	return
}

// copy from crypt/rsa/pkcs1v5.go
func unLeftPad(input []byte) (out []byte) {
	n := len(input)
	t := 2
	for i := 2; i < n; i++ {
		if input[i] == 0xff {
			t = t + 1
			if input[i + 1] == 0 {
				t += 1
			}
		} else {
			break
		}
	}
	out = make([]byte, n - t)
	copy(out, input[t:])
	return
}

// copy from crypt/rsa/pkcs1v5.go
func encrypt(c *big.Int, pub *rsa.PublicKey, m *big.Int) *big.Int {
	e := big.NewInt(int64(pub.E))
	c.Exp(m, e, pub.N)
	return c
}

/**
 * 私钥加密
 */
func PrivateEncrypt(privt *rsa.PrivateKey, data []byte) ([]byte, error) {
	signData, err := rsa.SignPKCS1v15(nil, privt, crypto.Hash(0), data)
	if err != nil {
		return nil, err
	}
	return signData, nil
}

/**
 * 私钥分段加密
 */
func PrivatePartEncrypt(privt *rsa.PrivateKey, data []byte) ([]byte, error) {
	length := len(data)
	i := 0
	var finalData []byte
	for i < length {
		signData, err := rsa.SignPKCS1v15(nil, privt, crypto.Hash(0), subString(data, i, ENCRYPT_KEYSIZE))
		if err != nil {
			return nil, err
		}
		finalData = bytesCombine(finalData, signData)
		i += ENCRYPT_KEYSIZE
	}
	return []byte(base64.StdEncoding.EncodeToString(finalData)), nil
}

/**
 * 私钥加签(SHA1withRSA)
 */
func PrivateSign(privt *rsa.PrivateKey, data []byte) ([]byte, error) {
	h := sha1.New()
	h.Write(data)
	hashed := h.Sum(nil)
	signData, err := rsa.SignPKCS1v15(rand.Reader, privt, crypto.SHA1, hashed)
	if err != nil {
		return nil, err
	}
	return signData, nil
}

/**
 * 私钥加签(MD5withRSA)
 */
func PrivateEncryptMD5withRSA(privt *rsa.PrivateKey, data []byte) ([]byte, error) {
	h := md5.New()
	h.Write(data)
	hashed := h.Sum(nil)
	signData, err := rsa.SignPKCS1v15(nil, privt, crypto.MD5, hashed)
	if err != nil {
		return nil, err
	}
	return signData, nil
}

/**
 * 私钥解密
 */
func PrivateDecrypt(privt *rsa.PrivateKey, data []byte) ([]byte, error) {
	length := len(data)
	i := 0
	var finalData []byte
	for i < length {
		signData, err := rsa.DecryptPKCS1v15(rand.Reader, privt, subString(data, i, DECRYPT_KEYSIZE))
		if err != nil {
			return nil, err
		}
		finalData = bytesCombine(finalData, signData)
		i += DECRYPT_KEYSIZE
	}
	return finalData, nil
}
/*********************** RSA ********************/

/*********************** AES ********************/
func (instance *XPEncryptImpl) AESEncrypt(origData, key, iv []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, iv)
	crypt     := make([]byte, len(origData))
	blockMode.CryptBlocks(crypt, origData)
	return base64.StdEncoding.EncodeToString(crypt), nil
}

func (instance *XPEncryptImpl) AESDecrypt(crypt string, key, iv []byte) (string, error) {
	decodeData, err := base64.StdEncoding.DecodeString(crypt)
	if err != nil {
		return "",err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(decodeData))
	blockMode.CryptBlocks(origData, decodeData)
	origData = PKCS5UnPadding(origData)

	return string(origData), nil
}
/*********************** AES ********************/
