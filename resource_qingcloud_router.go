package qingcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/router"
)

func resourceQingcloudRouter() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudRouterCreate,
		Read:   resourceQingcloudRouterRead,
		Update: resourceQingcloudRouterUpdate,
		Delete: nil,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "路由器名称",
			},
			"type": &schema.Schema{
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: withinArrayInt(0, 1, 2),
				Description: "路由器类型: 0 - 中型，1 - 小型，2 - 大型，默认为 1	",
			},
			"vpc_network": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: withinArrayString("192.168.0.0/16", "172.16.0.0/16"),
				Description:  "VPC 网络地址范围，目前支持 192.168.0.0/16 或 172.16.0.0/16 。 注：此参数只在北京3区需要且是必填参数。",
			},
			"securitygroup": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "需要加载到路由器上的防火墙ID",
			},

			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "介绍",
			},

			// 计算的数据
			"private_ip": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// resourceQingcloudRouterCreate
func resourceQingcloudRouterCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).router
	params := router.CreateRoutersRequest{}
	params.RouterName.Set(d.Get("name").(string))
	params.RouterType.Set(d.Get("type").(int))
	params.VpcNetwork.Set(d.Get("vpc_network").(string))
	params.SecurityGroup.Set(d.Get("securitygroup").(string))
	resp, err := clt.CreateRouters(params)
	if err != nil {
		return err
	}
	d.SetId(resp.Routers[0])

	if err := modifyRouterAttributes(d, meta, false); err != nil {
		return err
	}

	return resourceQingcloudRouterRead(d, meta)
}

func resourceQingcloudRouterRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).router
	_, err := RouterTransitionStateRefresh(clt, d.Id())
	if err != nil {
		return err
	}

	// 设置请求参数
	params := router.DescribeRoutersRequest{}
	params.RoutersN.Add(d.Id())
	params.Verbose.Set(1)
	resp, err := clt.DescribeRouters(params)
	if err != nil {
		return err
	}
	if len(resp.RouterSet) == 0 {
		return fmt.Errorf("资源可能我删除了")
	}

	v := resp.RouterSet[0]
	d.SetId(v.RouterID)
	d.Set("name", v.RouterName)
	d.Set("type", v.RouterType)
	d.Set("vpc_network", v.Vxnets)
	d.Set("securitygroup", v.SecurityGroupID)
	d.Set("description", v.Description)
	d.Set("private_ip", v.PrivateIP)
	return nil

}

func resourceQingcloudRouterDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).router
	params := router.PowerOffRoutersRequest{}
	params.RoutersN.Add(d.Id())
	_, err := clt.PowerOffRouters(params)
	if err != nil {
		return err
	}

	_, err = RouterTransitionStateRefresh(clt, d.Id())
	return err
}

func resourceQingcloudRouterUpdate(d *schema.ResourceData, meta interface{}) error {
	if !d.HasChange("description") && !d.HasChange("name") {
		return nil
	}

	err := modifyRouterAttributes(d, meta, false)
	if err != nil {
		return err
	}
	return applyRouterUpdates(meta, d.Id())

}
