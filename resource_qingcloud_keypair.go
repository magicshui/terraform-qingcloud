package qingcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/keypair"
)

func resourceQingcloudKeypair() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudKeypairCreate,
		Read:   resourceQingcloudKeypairRead,
		Update: resourceQingcloudKeypairUpdate,
		Delete: resourceQingcluodKeypairDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "密钥名称",
			},
			"public_key": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"mode": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: withinArrayString("system", "user"),
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"instance_ids": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceQingcloudKeypairCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).keypair
	params := keypair.CreateKeyPairRequest{}
	params.KeypairName.Set(d.Get("name").(string))
	params.PublicKey.Set(d.Get("public_key").(string))
	result, err := clt.CreateKeyPair(params)
	if err != nil {
		return fmt.Errorf("Error create Keypair: %s", err)
	}
	d.SetId(result.KeypairId)

	// TIP: 如果填写了 description ，那么需要再更新一次密钥的信息
	return modifyKeypairAttributes(d, meta, false)
}

func resourceQingcloudKeypairRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).keypair

	params := keypair.DescribeKeyPairsRequest{}
	params.KeypairsN.Add(d.Id())
	params.Verbose.Set(1)
	params.Limit.Set(1)
	resp, err := clt.DescribeKeyPairs(params)
	if err != nil {
		return fmt.Errorf("Error retrieving Keypair: %s", err)
	}

	// TODO: 如果密钥不存在?

	kp := resp.KeypairSet[0]
	d.Set("name", kp.KeypairName)
	d.Set("description", kp.Description)

	// 获取加载密钥的主机信息？
	// TODO: 这个时候，是否需要创建一个 k/v 的对？
	var instanceIDs = make([]string, 0)
	for _, o := range kp.InstanceIds {
		instanceIDs = append(instanceIDs, o)
	}
	d.Set("instance_ids", instanceIDs)

	return nil
}

// 如果要删除一个密钥，那么需要看一下这个密钥是否在其他的instance上是否有使用
func resourceQingcluodKeypairDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).keypair

	// 从所有的主机上删除密钥
	if err := deleteKeypairFromInstance(meta, d.Id(), d.Get("instance_ids").([]interface{})...); err != nil {
		return fmt.Errorf("Error %s", err)
	}

	// 删除密钥自身
	params := keypair.DeleteKeyPairsRequest{}
	params.KeypairsN.Add(d.Id())
	_, deleteErr := clt.DeleteKeyPairs(params)
	return deleteErr
}

func resourceQingcloudKeypairUpdate(d *schema.ResourceData, meta interface{}) error {
	// 密钥只能更新 Name 和 Description，其他的内容更新不了？
	return modifyKeypairAttributes(d, meta, false)
}
