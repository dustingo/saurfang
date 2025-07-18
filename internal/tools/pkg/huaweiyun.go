package pkg

import (
	"encoding/json"
	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v6/client"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	hwecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	hwregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/region"
	"log"
	"saurfang/internal/models/huaweiyun"
	"time"
)

type HwCloudClient interface {
	Fetch() ([]*ecs20140526.DescribeInstancesResponseBodyInstancesInstance, error)
}
type HuaweiClient struct {
	Client  *hwecs.EcsClient
	Region  string
	GroupId string
}

func NewHuaweiClient(ak, sk, region, groupId string) (cli HuaweiClient, err error) {
	rg, err := hwregion.SafeValueOf(region)
	if err != nil {
		return
	}
	//auth := basic.NewCredentialsBuilder().
	//	WithAk(ak).
	//	WithSk(sk).
	//	Build()

	auth, err := basic.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		SafeBuild()
	if err != nil {
		return
	}
	hcClient, err := hwecs.EcsClientBuilder().WithRegion(rg).WithCredential(auth).SafeBuild()
	if err != nil {
		return
	}
	client := hwecs.NewEcsClient(hcClient)

	return HuaweiClient{
		Client:  client,
		Region:  region,
		GroupId: groupId,
	}, nil
}

func (h *HuaweiClient) Fetch() (*[]huaweiyun.ECS, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f", r)
		}
	}()
	client := h.Client
	var servers []huaweiyun.ECS
	var nextToken string
	for {
		time.Sleep(2 * time.Second)
		request := &model.ListCloudServersRequest{}
		if nextToken == "" {
			var listExpectFields = []string{
				"addresses",
				"market_info",
				"metadata",
			}
			request.ExpectFields = &listExpectFields
			request.EnterpriseProjectId = &h.GroupId
			limitRequest := int32(5)
			request.Limit = &limitRequest
			//request.Marker = &marker
		} else {
			var listExpectFields = []string{
				"addresses",
				"market_info",
				"metadata",
			}
			request.ExpectFields = &listExpectFields
			request.EnterpriseProjectId = &h.GroupId
			limitRequest := int32(5)
			request.Limit = &limitRequest
			request.Marker = &nextToken
		}
		response, err := client.ListCloudServers(request)
		if err != nil {
			return &servers, err
		}
		instance := huaweiyun.ECS{}
		data, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &instance); err != nil {
			return &servers, err
		}
		servers = append(servers, instance)
		if len(instance.ServersLinks) == 0 {
			nextToken = ""
		} else {
			nextToken = instance.Servers[len(instance.Servers)-1].Id
		}
		if nextToken == "" {
			break
		}
	}
	return &servers, nil
}
