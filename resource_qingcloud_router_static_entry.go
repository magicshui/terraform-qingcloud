package qingcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/router"
)

func resourceQingcloudRouterStaticEntry() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudRouterStaticEntryCreate,
		Read:   resourceQingcloudRouterStaticEntryRead,
		Update: resourceQingcloudRouterStaticEntryUpdate,
		Delete: resourceQingcloudRouterStaticEntryDelete,
		Schema: map[string]*schema.Schema{

			"router_static": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: "需要增加条目的路由器规则ID	",
			},
			"val1": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: `PPTP 账户信息：val1 表示账户名
					三层 GRE 隧道：val1 表示目标网络
					三层 IPsec 隧道：val1 表示本地网络 (val2 可为空)`,
			},
			"val2": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: `PPTP 账户信息：val2 表示密码
					三层 IPsec 隧道：val2 表示目标网络 (val1 可为空)`,
			},

			"router": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "路由器 ID",
			},
		},
	}
}

func resourceQingcloudRouterStaticEntryCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).router

	params := router.AddRouterStaticEntriesRequest{}
	params.RouterStatic.Set(d.Get("router_static").(string))
	params.EntriesNVal1.Add(d.Get("val1").(string))
	params.EntriesNVal2.Add(d.Get("val2").(string))
	resp, err := clt.AddRouterStaticEntries(params)
	if err != nil {
		return err
	}

	if len(resp.RouterStaticEntries) != 1 {
		return fmt.Errorf("资源不存在")
	}

	d.SetId(resp.RouterStaticEntries[0])
	return applyRouterUpdates(meta, d.Get("router").(string))
}

func resourceQingcloudRouterStaticEntryRead(d *schema.ResourceData, meta interface{}) error {
	// NOTICE: 青云的这个请求有 bug，使用了 EntryID 但是返回了所有的数据

	clt := meta.(*QingCloudClient).router

	params := router.DescribeRouterStaticEntriesRequest{}
	params.RouterStaticEntryID.Set(d.Id())
	resp, err := clt.DescribeRouterStaticEntries(params)
	if err != nil {
		return err
	}

	if len(resp.RouterStaticEntrySet) == 0 {
		return fmt.Errorf("资源不存在")
	}

	for _, v := range resp.RouterStaticEntrySet {
		if v.RouterID == d.Id() {
			d.Set("val1", v.Val1)
		}
	}

	return nil
}

func resourceQingcloudRouterStaticEntryDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).router
	params := router.DeleteRouterStaticEntriesReqeust{}
	params.RouterStaticEntriesN.Add(d.Id())
	_, err := clt.DeleteRouterStaticEntries(params)
	if err != nil {
		return err
	}
	return applyRouterUpdates(meta, d.Get("router").(string))
}

func resourceQingcloudRouterStaticEntryUpdate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).router
	params := router.ModifyRouterStaticEntryAttributesReqeust{}
	params.RouterStaticEntry.Set(d.Id())
	params.Val1.Set(d.Get("val1").(string))
	params.Val2.Set(d.Get("val2").(string))

	_, err := clt.ModifyRouterStaticEntryAttributes(params)
	if err != nil {
		return err
	}

	return applyRouterUpdates(meta, d.Get("router").(string))
}
