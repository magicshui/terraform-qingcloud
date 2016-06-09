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
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"public_key": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"encrypt": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ssh-rsa",
				Description:  "加密算法，有效值为 ssh-rsa 和 ssh-dss，默认为 ssh-rsa。",
				ValidateFunc: withinArrayString("ssh-rsa", "ssh-dss"),
			},

			// Computed
			"id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "SSH 密钥 ID",
			},
		},
	}
}

func resourceQingcloudKeypairCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).keypair

	// 创建请求参数
	params := keypair.CreateKeyPairRequest{}
	params.KeypairName.Set(d.Get("name").(string))
	params.PublicKey.Set(d.Get("public_key").(string))
	params.Mode.Set("user")
	params.EncryptMethod.Set(d.Get("encrypt").(string))

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
	if len(resp.KeypairSet) == 0 {
		return err
	}

	kp := resp.KeypairSet[0]
	d.Set("name", kp.KeypairName)
	d.Set("description", kp.Description)

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
