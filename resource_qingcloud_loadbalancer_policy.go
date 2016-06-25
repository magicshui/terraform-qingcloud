package qingcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/loadbalancer"
)

func resourceQingcloudLoadbalancerPloicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudLoadbalancerPloicyCreate,
		Read:   resourceQingcloudLoadbalancerPloicyRead,
		Update: resourceQingcloudLoadbalancerPloicyUpdate,
		Delete: resourceQingcloudLoadbalancerPloicyDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "名称",
				Required:    true,
			},
			"operator": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "转发策略规则间的逻辑关系：”and” 是『与』，”or” 是『或』，默认是 “or”",
				ValidateFunc: withinArrayString("and", "or"),
				Required:     true,
			},
		},
	}
}

func resourceQingcloudLoadbalancerPloicyCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).loadbalancer
	params := loadbalancer.CreateLoadBalancerPolicyRequest{}
	params.Operator.Set(d.Get("operator").(string))
	resp, err := clt.CreateLoadBalancerPolicy(params)
	if err != nil {
		return err
	}

	d.SetId(resp.LoadbalancerPolicyId)

	params2 := loadbalancer.ModifyLoadBalancerPolicyAttributesRequest{}
	params2.LoadbalancerPolicy.Set(d.Id())
	params2.LoadbalancerPolicyName.Set(d.Get("name").(string))
	_, err = clt.ModifyLoadBalancerPolicyAttributes(params2)

	return err
}

func resourceQingcloudLoadbalancerPloicyRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).loadbalancer
	params := loadbalancer.DescribeLoadBalancerPoliciesRequest{}
	params.Verbose.Set(1)
	params.LoadbalancerPoliciesN.Add(d.Id())
	resp, err := clt.DescribeLoadBalancerPolicies(params)
	d.Set("name", resp.LoadbalancerPolicySet[0].LoadbalancerPolicyName)
	return err
}

func resourceQingcloudLoadbalancerPloicyUpdate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).loadbalancer
	params := loadbalancer.ModifyLoadBalancerPolicyAttributesRequest{}
	params.LoadbalancerPolicy.Set(d.Id())
	params.LoadbalancerPolicyName.Set(d.Get("name").(string))
	_, err := clt.ModifyLoadBalancerPolicyAttributes(params)
	return err
}

func resourceQingcloudLoadbalancerPloicyDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).loadbalancer
	params := loadbalancer.DeleteLoadBalancerPoliciesRequest{}
	params.LoadbalancerPoliciesN.Add(d.Id())
	_, err := clt.DeleteLoadBalancerPolicies(params)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}
