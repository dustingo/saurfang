package pkg

import (
	"encoding/json"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v6/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"saurfang/internal/models/aliyun"
)

type CloudClient interface {
	Fetch() ([]*ecs20140526.DescribeInstancesResponseBodyInstancesInstance, error)
}
type AliyunClient struct {
	Client  *ecs20140526.Client
	Region  string
	GroupId string
}

func NewAlyumClient(ak, sk, endpoint, region, groupId string) (AliyunClient, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(ak),
		AccessKeySecret: tea.String(sk),
	}
	config.Endpoint = tea.String(endpoint)
	cli := &ecs20140526.Client{}
	cli, err := ecs20140526.NewClient(config)

	return AliyunClient{Client: cli, Region: region, GroupId: groupId}, err
}
func (a AliyunClient) Fetch() ([]aliyun.ECS, error) {
	client := a.Client
	var nextToken string
	instances := []aliyun.ECS{}

	for {
		request := &ecs20140526.DescribeInstancesRequest{
			RegionId:        tea.String(a.Region),
			ResourceGroupId: tea.String(a.GroupId),
			NextToken:       tea.String(nextToken),
		}
		runtime := &util.RuntimeOptions{}
		res, err := client.DescribeInstancesWithOptions(request, runtime)
		if err != nil {
			return instances, err
		}
		instance := aliyun.ECS{}
		json.Unmarshal([]byte(res.Body.String()), &instance)
		instances = append(instances, instance)
		nextToken = *res.Body.NextToken
		if nextToken == "" {
			break
		}
	}
	return instances, nil
}
