package qingcloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/vxnet"
)

func resourceQingcloudVxnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudVxnetCreate,
		Read:   resourceQingcloudVxnetRead,
		Update: resourceQingcloudVxnetUpdate,
		Delete: resourceQingcloudVxnetDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "VXNet 的名称",
			},
			"type": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Description: "私有网络类型，1 - 受管私有网络，0 - 自管私有网络。	",
				Default:      1,
				ValidateFunc: withinArrayInt(0, 1),
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceQingcloudVxnetCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).vxnet
	params := vxnet.CreateVxnetsRequest{}
	params.VxnetName.Set(d.Get("name").(string))
	params.VxnetType.Set(d.Get("type").(int))
	resp, err := clt.CreateVxnets(params)
	if err != nil {
		return err
	}

	d.SetId(resp.Vxnets[0])

	if err := modifyVxnetAttributes(d, meta, false); err != nil {
		return err
	}

	return resourceQingcloudVxnetRead(d, meta)
}

func resourceQingcloudVxnetRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).vxnet
	params := vxnet.DescribeVxnetsRequest{}
	params.VxnetsN.Add(d.Id())
	params.Verbose.Set(1)
	resp, err := clt.DescribeVxnets(params)
	if err != nil {
		return fmt.Errorf("Error retrieving vxnets: %s", err)
	}

	if len(resp.VxnetSet) != 1 {
		return fmt.Errorf("资源可能被删除了")
	}

	sg := resp.VxnetSet[0]
	d.Set("name", sg.VxnetName)
	d.Set("description", sg.Description)

	// TODO: 青云目前不支持
	// d.Set("router", sg.Router.RouterID)
	d.Set("ip_network", sg.Router.IPNetwork)
	return nil
}

func resourceQingcloudVxnetDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).vxnet
	params := vxnet.DeleteVxnetsRequest{}
	params.VxnetsN.Add(d.Id())
	_, err := clt.DeleteVxnets(params)
	return err
}

func resourceQingcloudVxnetUpdate(d *schema.ResourceData, meta interface{}) error {
	return modifyVxnetAttributes(d, meta, false)
}
