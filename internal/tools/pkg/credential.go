package pkg

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"saurfang/internal/config"
	"saurfang/internal/models/credential"
	"saurfang/internal/models/user"
	"strconv"
	"time"
)

// generateSecureRandomString 生成随机字符串
func generateSecureRandomString(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be greater than 0")
	}
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	for i := range bytes {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}
	return string(bytes), nil
}

// GenerateSecureRandomStringForTest 仅用于测试的公开函数
// 生成随机字符串
func GenerateSecureRandomStringForTest(length int) (string, error) {
	return generateSecureRandomString(length)
}

// GenerateSignatureForTest 仅用于测试的公开函数
// 生成签名
func GenerateSignatureForTest(accessKey, secretKey, method, path, timestamp string) (string, error) {
	return generateSignature(accessKey, secretKey, method, path, timestamp)
}
func GenerateAKSKPair(userid uint) (*credential.UserCredential, error) {
	accessKey, err := generateSecureRandomString(20)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access key: %v", err)
	}
	secretKey, err := generateSecureRandomString(40)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret key: %v", err)
	}
	return &credential.UserCredential{
		AccessKey: accessKey,
		SecretKey: secretKey,
		UserID:    userid,
		Status:    "active",
	}, nil
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
	// 生成签名
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(canonicalRequest))
	signature := hex.EncodeToString(h.Sum(nil))
	return signature, nil
}

// VerifySignature 校验签名user_id,code,error
func VerifySignature(accessKey, signature, method, path string, timestamp string) (uint, int, bool) {
	userid, sk, ok := validAk(accessKey)
	if !ok {
		return 0, 401, false
	}
	// 验证时间戳（防止重放攻击）
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return 0, 403, false
	}

	// 时间戳有效期5分钟
	if time.Now().Local().Unix()-ts > 300 {
		return 0, 403, false
	}
	sig, err := generateSignature(accessKey, sk, method, path, timestamp)
	if err != nil {
		return 0, 403, false
	}
	if !hmac.Equal([]byte(signature), []byte(sig)) {
		return 0, 403, false
	}
	return userid, 200, true
}

// validAk 校验ak是否有效,返回 user_id,sk,bool
func validAk(ak string) (uint, string, bool) {
	var userCredential credential.UserCredential
	if err := config.DB.Table("user_credentials").Where("access_key = ?", ak).Find(&userCredential).Error; err != nil {
		return 0, "", false
	}
	if userCredential.Status == "active" {
		return userCredential.UserID, userCredential.SecretKey, true
	} else {
		return 0, "", false
	}
}

// GetRoleOfUser 获取用户的角色ID

func GetRoleOfUser(user_id uint) (uint, error) {
	var ur user.UserRole
	if err := config.DB.Where("user_id = ?", user_id).First(&ur).Error; err != nil {
		return 0, err
	}
	return ur.RoleID, nil
}
