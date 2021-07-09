package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"
)

var (
	privateKey *rsa.PrivateKey //私钥
	publicKey  *rsa.PublicKey  //公钥(放在web端)
)

const (
	salt = "9IYds**^dfd>>??0sY9)?sd@df&*Ty)H23-7%#Pa;]k(" //用户密码加盐
)

//生成密钥对到文件中
func GenerateRsaKey(keySize int) {
	privateKey, _ = rsa.GenerateKey(rand.Reader, keySize)
	derText := x509.MarshalPKCS1PrivateKey(privateKey)
	block := pem.Block{
		Type:  "rsa private key", //作用仅仅是提示给看私匙的人，这是rsa私匙
		Bytes: derText,
	}
	fo, err := os.Create("privateKey.pem")
	if err != nil {
		panic(err)
	}
	defer fo.Close()
	pem.Encode(fo, &block)
	publicKey = &privateKey.PublicKey
	derStream, _ := x509.MarshalPKIXPublicKey(publicKey)
	block = pem.Block{
		Type:  "rsa public key",
		Bytes: derStream,
	}
	fo, _ = os.Create("publicKey.pem")
	defer fo.Close()
	pem.Encode(fo, &block)
}

//初始化rsa公钥,私钥, 和用户密码盐
func initRsaKeys(rsaCfg H) error {
	if privateKeyStr, ok := rsaCfg["privateKey"].(string); !ok {
		return errors.New("读取私钥失败")
	} else {
		block, _ := pem.Decode([]byte(privateKeyStr))
		var err error
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return err
		}
	}
	//if publicKeyStr, ok := rsaCfg["publicKey"].(string); !ok {
	//	return errors.New("读取钥失败")
	//} else {
	//	block, _ := pem.Decode([]byte(publicKeyStr))
	//	if pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes); err != nil {
	//		return err
	//	} else {
	//		publicKey, _ = pubInterface.(*rsa.PublicKey)
	//	}
	//}
	return nil
}

//公钥加密
func RSAEncrypt(text []byte) []byte {
	cipherText, _ := rsa.EncryptPKCS1v15(rand.Reader, publicKey, text)
	return cipherText
}

//私钥解密,解密失败返回nil
func RSADecrypt(cipherText string) []byte {
	b, _ := base64.StdEncoding.DecodeString(cipherText)
	if plainText, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, b); err == nil {
		return plainText
	}
	return nil
}

//用户密码进行MD5加盐哈希
func GetMD5Password(pwd string) string {
	return GetMD5OfStr(pwd + salt)
}

//用户密码用私钥解密,再用MD5加盐哈希的密码,解密失败返回空串
func PassWordHandle(cipherPwd string) string {
	pwd := RSADecrypt(cipherPwd)
	if pwd == nil {
		return ""
	}
	return GetMD5Password(string(pwd))
}
