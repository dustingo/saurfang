package pkg

import (
	"context"
	"fmt"
	"github.com/apenella/go-ansible/v2/pkg/adhoc"
	"golang.org/x/sync/semaphore"
	"gorm.io/gorm"
	"log"
	"saurfang/internal/models/gamehost"
	"time"
)

const (
	Active  uint = 1
	Offline uint = 0
)

func checkActive(public_ip, private_ip string, port int) (bool, error) {
	var host string
	if private_ip != "" {
		host = private_ip
	} else {
		host = public_ip
	}
	ansibleAdhocOptions := &adhoc.AnsibleAdhocOptions{
		Inventory:  fmt.Sprintf("%s,", host),
		ModuleName: "ping",
		ExtraVars:  map[string]interface{}{"ansible_ssh_port": port},
	}
	err := adhoc.NewAnsibleAdhocExecute("all").WithAdhocOptions(ansibleAdhocOptions).Execute(context.TODO())
	if err != nil {
		return false, err
	}
	return true, nil
}
func CheckActiveInterval(db *gorm.DB) {
	// 待优化
	// is_active 状态需要增加缓存来避免数据库重复写入s
	ctx := context.Background()
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			const maxConcurrency = 2
			sem := semaphore.NewWeighted(int64(maxConcurrency))
			// 获取主机列表
			var hosts []gamehost.SaurfangHosts
			if err := db.Select("id", "hostname", "public_ip", "private_ip", "port").Find(&hosts).Error; err != nil {
				log.Println("db error: ", err)
				continue
			}
			for _, host := range hosts {
				if err := sem.Acquire(ctx, 1); err != nil {
					log.Println("acquire semaphore error: ", err)
					continue
				}
				go func(host gamehost.SaurfangHosts) {
					defer sem.Release(1)
					active, err := checkActive(host.PublicIP, host.PrivateIP, host.Port)
					if err != nil {
						log.Println("checkActive error:", err)
						db.Model(&gamehost.SaurfangHosts{}).Where("id = ?", host.ID).Update("is_active", Offline)
						return
					}
					if !active {
						log.Println("Host not alive:", host.Hostname, host.PublicIP, host.PrivateIP)
						db.Model(&gamehost.SaurfangHosts{}).Where("id = ?", host.ID).Update("is_active", Offline)
					} else {
						db.Model(&gamehost.SaurfangHosts{}).Where("id = ?", host.ID).Update("is_active", Active)
					}
				}(host)
			}

		}
	}

}
