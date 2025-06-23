package pkg

import (
	"bytes"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"saurfang/internal/models/serverconfig"
	"time"
)

var originPath, tempPath string

func init() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalln(err)
	}
	execDir := filepath.Dir(execPath)
	log.Println("运行目录: ", execDir)
	workDirs := []string{"origin", "temp"}
	for _, dir := range workDirs {
		fullPath := filepath.Join(execDir, dir)
		stat, err := os.Stat(fullPath)
		if err == nil {
			if stat.IsDir() {
				continue
			}
		}
		if os.IsNotExist(err) {
			if err := os.MkdirAll(fullPath, 0777); err != nil {
				log.Fatalln("创建依赖目录失败,", err.Error())
			}
		}
	}
	originPath = filepath.Join("origin")
	tempPath = filepath.Join("temp")

}

// ssh 客户端
type SSHConfig struct {
	User           string `json:"user"`
	IP             string `json:"ip"`
	Port           int    `json:"port"` // default 22
	PrivateKeyPath string `json:"private_key_path"`
}

type SFTPClient struct {
	Client *sftp.Client
	Conn   *ssh.Client
}

func NewSSHClient(sshConfig SSHConfig) (*ssh.Client, error) {
	key, err := os.ReadFile(sshConfig.PrivateKeyPath)
	if err != nil {
		return nil, err
	} // Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshConfig.IP, sshConfig.Port), &ssh.ClientConfig{
		User: sshConfig.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         5 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境请使用证书验证
	})
	if err != nil {
		return nil, err
	}
	return sshClient, nil
}

// 初始化SFTP客户端
func NewSFTPClient(sshConfig SSHConfig) (*SFTPClient, error) {
	key, err := os.ReadFile(sshConfig.PrivateKeyPath)
	if err != nil {
		return nil, err
	} // Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshConfig.IP, sshConfig.Port), &ssh.ClientConfig{
		User: sshConfig.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         5 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境请使用证书验证
	})
	if err != nil {
		return nil, err
	}

	// 创建 SFTP 客户端
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, err
	}
	return &SFTPClient{
		Client: sftpClient,
		Conn:   sshClient,
	}, nil
}

// 下载文件
func (s *SFTPClient) DownloadFile(remotePath string) ([]byte, error) {
	file, err := s.Client.Open(remotePath)
	if err != nil {

		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

// 上传文件
func (s *SFTPClient) UploadFile(localPath string, remotePath string) error {
	// 创建远程文件
	s.Client.Rename(remotePath, fmt.Sprintf("%s.%s.bak", remotePath, time.Now().Format("20060102150405")))
	remoteFile, err := s.Client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	// 读取本地文件内容
	localContent, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	// 写入远程文件
	_, err = remoteFile.Write(localContent)
	return err
}

func (s *SFTPClient) LockFile(remotePath string) error {
	lockPath := remotePath + ".lock"
	_, err := s.Client.Create(lockPath)
	if err != nil {
		return err
	}
	return nil
}

func (s *SFTPClient) UnlockFile(remotePath string) error {
	lockPath := remotePath + ".lock"
	err := s.Client.Remove(lockPath)
	if err != nil {
		return fmt.Errorf("failed to unlock file: %s", remotePath)
	}
	return nil
}

func ProcessServer(config serverconfig.Configs) error {
	sftpClient, err := NewSFTPClient(SSHConfig{
		User:           config.User,
		IP:             config.IP,
		Port:           config.Port,
		PrivateKeyPath: os.Getenv("PRIVATE_KEY_PATH"),
	})
	if err != nil {
		return fmt.Errorf("connect failed to %s: %w", config.IP, err)
	}
	defer sftpClient.Conn.Close()

	lockErr := sftpClient.LockFile(config.ConfigFile)
	if lockErr != nil {
		return fmt.Errorf("acquire lock failed on %s: %w", config.ConfigFile, lockErr)
	}
	defer func() {
		unlockErr := sftpClient.UnlockFile(config.ConfigFile)
		if unlockErr != nil {
			log.Println("failed to unlock file: ", unlockErr)
		}
	}()
	// 从远程服务器下载配置文件
	originalContent, err := sftpClient.DownloadFile(config.ConfigFile)
	if err != nil {
		return fmt.Errorf("download failed from %s: %w", config.SvcName, err)
	}
	// 将远程配置文件写入到本地
	originFile, err := os.Create(fmt.Sprintf("%s/%s-%s", originPath, config.ServerId, config.IP))
	//originFile, err := os.Create(fmt.Sprintf("/work/origin/%s-%s", config.ServerId, config.IP))
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	if _, err := originFile.Write(originalContent); err != nil {
		return fmt.Errorf("write origin file failed: %w", err)
	}
	// 以原始配置文件为模板, 使用变量替换,并写入临时文件
	var buf bytes.Buffer
	tmpl, err := template.ParseFiles(originFile.Name())
	if err != nil {
		return fmt.Errorf("template parse file failed: %w", err)
	}
	if err := tmpl.Execute(&buf, config.Vars); err != nil {

	}
	// 尽量不使用must 避免panic
	//if err := template.Must(template.ParseFiles(originFile.Name())).Execute(&buf, config.Vars); err != nil {
	//	return fmt.Errorf("parse template failed: %w", err)
	//}
	//tempFile, err := os.CreateTemp("/work/temp", config.ServerId+"-"+config.IP+"-")
	tempFile, err := os.CreateTemp(tempPath, fmt.Sprintf("%s-%s-%s-", filepath.Base(config.ConfigFile), config.ServerId, config.IP))
	if err != nil {
		return fmt.Errorf("create temp file failed: %w", err)
	}
	os.Chmod(tempFile.Name(), 0644)
	defer tempFile.Close()
	_, err = io.WriteString(tempFile, buf.String())
	if err != nil {
		return fmt.Errorf("write temp file failed: %w", err)
	}
	fmt.Println("tem file name = ", tempFile.Name())
	// 将配置好的配置文件传发送到远程服务器
	if err := sftpClient.UploadFile(tempFile.Name(), config.ConfigFile); err != nil {
		return fmt.Errorf("upload temp file failed: %w", err)
	}
	return nil
}
