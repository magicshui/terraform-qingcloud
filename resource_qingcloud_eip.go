package qingcloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/eip"
)

func resourceQingcloudEip() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudEipCreate,
		Read:   resourceQingcloudEipRead,
		Update: resourceQingcloudEipUpdate,
		Delete: resourceQingcloudEipDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "公网 IP 的名称",
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"bandwidth": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "公网IP带宽上限，单位为Mbps",
			},
			"billing_mode": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "traffic",
				Description:  "公网IP计费模式：bandwidth 按带宽计费，traffic 按流量计费，默认是 bandwidth",
				ValidateFunc: withinArrayString("traffic", "bandwidth"),
			},
			"need_icp": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				Description:  "是否需要备案，1为需要，0为不需要，默认是0",
				ValidateFunc: withinArrayInt(0, 1),
			},

			// 自动计算资源
			"addr": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceQingcloudEipCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).eip
	params := eip.AllocateEipsRequest{}
	params.Bandwidth.Set(d.Get("bandwidth").(int))
	params.BillingMode.Set(d.Get("billing_mode").(string))
	params.EipName.Set(d.Get("name").(string))
	params.NeedIcp.Set(d.Get("need_icp").(int))
	resp, err := clt.AllocateEips(params)
	if err != nil {
		return err
	}
	d.SetId(resp.Eips[0])

	if err := modifyEipAttributes(d, meta, true); err != nil {
		return err
	}

	// 配置一下
	return resourceQingcloudEipRead(d, meta)
}

func resourceQingcloudEipRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).eip
	_, err := EipTransitionStateRefresh(clt, d.Id())
	if err != nil {
		return err
	}

	params := eip.DescribeEipsRequest{}
	params.EipsN.Add(d.Id())
	params.Verbose.Set(1)

	resp, err := clt.DescribeEips(params)
	if err != nil {
		return err
	}

	if len(resp.EipSet) == 0 {
		return fmt.Errorf("资源可能已经被删除")
	}

	e := resp.EipSet[0]

	d.Set("name", e.EipName)
	d.Set("description", e.Description)
	d.Set("billing_mode", e.BillingMode)
	d.Set("bandwidth", e.Bandwidth)

	// 自动计算
	d.SetId(e.EipID)
	d.Set("addr", e.EipAddr)

	return nil
}

func resourceQingcloudEipDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).eip

	params := eip.DissociateEipsRequest{}
	params.EipsN.Add(d.Id())

	_, err := clt.DissociateEips(params)
	if err != nil {
		return nil
	}

	// 等待 EIP 的状态稳定下来
	_, err = EipTransitionStateRefresh(clt, d.Id())
	return err
}

func resourceQingcloudEipUpdate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).eip

	if !d.HasChange("name") && !d.HasChange("description") && !d.HasChange("bandwidth") && !d.HasChange("billing_mode") {
		return nil
	}

	// TODO: 是否需要添加

	if d.HasChange("bandwidth") {
		params := eip.ChangeEipsBandwidthRequest{}
		params.EipsN.Add(d.Id())
		params.Bandwidth.Set(d.Get("bandwidth").(int))
		_, err := clt.ChangeEipsBandwidth(params)
		if err != nil {
			return err
		}
	}

	if d.HasChange("billing_mode") {
		params := eip.ChangeEipsBillingModeRequest{}
		params.EipsN.Add(d.Id())
		params.BillingMode.Set(d.Get("billing_mode").(string))
		_, err := clt.ChangeEipsBillingMode(params)
		if err != nil {
			return err
		}
	}

	if err := modifyEipAttributes(d, meta, false); err != nil {
		return err
	}

	_, err := EipTransitionStateRefresh(clt, d.Id())
	return err
}
