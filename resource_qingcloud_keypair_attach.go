package qingcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/keypair"
	"strings"

	"log"
)

func resourceQingcloudKeypairAttach() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudKeypairAttachCreate,
		Read:   resourceQingcloudKeypairAttachRead,
		Update: resourceQingcloudKeypairAttachUpdate,
		Delete: resourceQingcluodKeypairAttachDelete,
		Schema: map[string]*schema.Schema{
			"keypair": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "密钥 ID",
				ForceNew:    true,
			},
			"instance": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "主机 ID",
				ForceNew:    true,
			},

			// 自动计算
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceQingcloudKeypairAttachCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).keypair
	params := keypair.AttachKeyPairsRequest{}
	params.KeypairsN.Add(d.Get("keypair").(string))
	params.InstancesN.Add(d.Get("instance").(string))
	_, err := clt.AttachKeyPairs(params)
	if err != nil {
		return err
	}

	// 密钥的ID: keypair-instance
	d.SetId(fmt.Sprintf("%s*%s", d.Get("keypair").(string), d.Get("instance").(string)))
	_, err = InstanceTransitionStateRefresh(meta.(*QingCloudClient).instance, d.Get("instance").(string))
	return err
}

func resourceQingcloudKeypairAttachRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).keypair
	keypairID := d.Get("keypair").(string)
	instanceID := d.Get("instance").(string)

	params := keypair.DescribeKeyPairsRequest{}
	params.KeypairsN.Add(keypairID)
	params.InstanceID.Set(instanceID)
	params.Verbose.Set(1)
	resp, err := clt.DescribeKeyPairs(params)

	log.Printf("[QC] Keypair x Instance: %v", resp)

	if err != nil {
		return err
	}

	if resp.TotalCount == 0 {
		d.SetId("")
		return err
	}

	// 如果主机列表为0，代表没有在使用，那么需要重新加载进去
	if len(resp.KeypairSet[0].InstanceIds) == 0 {
		d.SetId("")
		log.Printf("[QC] INSTANCE REMOVED: %s", instanceID)
		return nil
	}

	// 判断当前的主机是否在列表中？
	for _, v := range resp.KeypairSet[0].InstanceIds {
		if v == instanceID {
			log.Printf("[QC]  found instance: %s", v)
			return nil
		}
		d.SetId("")
	}
	return nil
}

func resourceQingcloudKeypairAttachUpdate(d *schema.ResourceData, meta interface{}) error {
	// if _, n := d.GetChange("need_recreate"); n.(bool) {
	// 	log.Printf("[DEBUG]resourceQingcloudKeypairAttachUpdate recreate ")
	// 	return resourceQingcloudKeypairCreate(d, meta)
	// }
	return nil
}

func resourceQingcluodKeypairAttachDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).keypair
	keypairID := strings.Split(d.Id(), "*")[0]
	instanceID := strings.Split(d.Id(), "*")[1]
	params := keypair.DetachKeyPairsRequest{}
	params.InstancesN.Add(instanceID)
	params.KeypairsN.Add(keypairID)
	_, err := clt.DetachKeyPairs(params)
	if err != nil {
		return err
	}
	// 等待主机的状态改变
	_, err = InstanceTransitionStateRefresh(meta.(*QingCloudClient).instance, instanceID)
	return err
}
