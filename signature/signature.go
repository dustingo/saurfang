package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"
)

var (
	accessKey     string
	secretKey     string
	requestPath   string
	requestMethod string
	//requestTime   string
)

func main() {
	flag.StringVar(&accessKey, "key", "", "access key")
	flag.StringVar(&secretKey, "secret", "", "secret key")
	flag.StringVar(&requestPath, "request-path", "", "request path")
	flag.StringVar(&requestMethod, "request-method", "", "request method(capital) ")
	//flag.StringVar(&requestTime, "request-time", "", "request time")
	flag.Parse()
	timestamp := strconv.FormatInt(time.Now().Local().Unix(), 10)
	sig, err := generateSignature(accessKey, secretKey, requestMethod, requestPath, timestamp)
	if err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Println("X-Access-Key:", accessKey)
	fmt.Println("X-Timestamp:", timestamp)
	fmt.Println("X-Signature:", sig)

}

// generateSignature 根据请求的信息重新生成签名
func generateSignature(accessKey, secretKey, method, path string, timestamp string) (string, error) {
	canonicalRequest := fmt.Sprintf(
		"%s-%s-%s-%s",
		accessKey,
		method,
		path,
		timestamp,
	)
	fmt.Println("canonicalRequest:", canonicalRequest)
	// 生成签名
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(canonicalRequest))
	signature := hex.EncodeToString(h.Sum(nil))
	return signature, nil

}
