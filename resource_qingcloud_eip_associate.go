package qingcloud

import (
	"log"

	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/eip"
	"github.com/magicshui/qingcloud-go/router"
)

func resourceQingcloudEipAssociate() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudEipAssociateCreate,
		Read:   resourceQingcloudEipAssociateRead,
		Update: resourceQingcloudEipAssociateUpdate,
		Delete: resourceQingcloudEipAssociateDelete,
		Schema: map[string]*schema.Schema{
			"eip": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "公网IP",
			},
			"resource": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "资源 ID，只能设置成 路由 或者 主机",
			},
		},
	}
}

// NOTE: 不设置ID会导致不能删除数据
func resourceQingcloudEipAssociateCreate(d *schema.ResourceData, meta interface{}) error {
	eipID := d.Get("eip").(string)
	d.SetId(eipID)

	resourceID := d.Get("resource").(string)
	switch strings.Split(resourceID, "-")[0] {

	case "rtr":
		clt := meta.(*QingCloudClient).router
		params := router.ModifyRouterAttributesRequest{}
		params.Eip.Set(eipID)
		params.Router.Set(resourceID)
		if _, err := clt.ModifyRouterAttributes(params); err != nil {
			return err
		}
		return applyRouterUpdates(meta, resourceID)

	case "i":
		clt := meta.(*QingCloudClient).eip
		params := eip.AssociateEipRequest{}
		params.Eip.Set(eipID)
		params.Instance.Set(resourceID)
		if _, err := clt.AssociateEip(params); err != nil {
			return err
		}
		_, err := InstanceTransitionStateRefresh(meta.(*QingCloudClient).instance, resourceID)
		return err
	}
	return nil
}

func resourceQingcloudEipAssociateRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).eip
	params := eip.DescribeEipsRequest{}
	params.EipsN.Add(d.Get("eip").(string))
	params.Verbose.Set(1)

	resp, err := clt.DescribeEips(params)
	if err != nil {
		return err
	}

	if resp.TotalCount != 1 {
		return fmt.Errorf("资源不存在")
	}

	d.Set("resource", resp.EipSet[0].Resource.ResourceID)
	return nil
}

// 绑定的资源不能再重新使用，需要删除
func resourceQingcloudEipAssociateUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// 删除绑定
func resourceQingcloudEipAssociateDelete(d *schema.ResourceData, meta interface{}) error {
	resourceID := d.Get("resource").(string)
	eipID := d.Get("eip").(string)

	log.Printf("Resource ID: %s", resourceID)

	switch strings.Split(resourceID, "-")[0] {
	case "rtr":
		clt := meta.(*QingCloudClient).router
		params := router.ModifyRouterAttributesRequest{}
		params.Eip.Set(eipID)
		params.Router.Set(resourceID)
		if _, err := clt.ModifyRouterAttributes(params); err != nil {
			return err
		}
		return applyRouterUpdates(meta, resourceID)

	case "i":
		clt := meta.(*QingCloudClient).eip
		params := eip.DissociateEipsRequest{}
		params.EipsN.Add(eipID)
		if _, err := clt.DissociateEips(params); err != nil {
			return err
		}
		_, err := InstanceTransitionStateRefresh(meta.(*QingCloudClient).instance, resourceID)
		return err
	}
	return nil
}
