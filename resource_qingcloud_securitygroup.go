package qingcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/securitygroup"
)

func resourceQingcloudSecuritygroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudSecuritygroupCreate,
		Read:   resourceQingcloudSecuritygroupRead,
		Update: resourceQingcloudSecuritygroupUpdate,
		Delete: resourceQingcloudSecuritygroupDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "防火墙名称",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "防火墙介绍",
			},
		},
	}
}

func resourceQingcloudSecuritygroupCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).securitygroup

	name := d.Get("name").(string)
	params := securitygroup.CreateSecurityGroupRequest{}
	params.SecurityGroupName.Set(name)
	resp, err := clt.CreateSecurityGroup(params)
	if err != nil {
		return err
	}

	d.SetId(resp.SecurityGroupId)

	if d.Get("description") != nil {
		params := securitygroup.ModifySecurityGroupAttributesRequest{}
		params.SecurityGroup.Set(resp.SecurityGroupId)
		params.Description.Set(d.Get("description").(string))
		_, err = clt.ModifySecurityGroupAttributes(params)
	}
	return err
}

func resourceQingcloudSecuritygroupRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).securitygroup
	params := securitygroup.DescribeSecurityGroupsRequest{}
	params.SecurityGroupsN.Add(d.Id())
	params.Verbose.Set(1)
	resp, err := clt.DescribeSecurityGroups(params)
	if err != nil {
		return err
	}
	if len(resp.SecurityGroupSet) != 1 {
		return fmt.Errorf("资源可能被删除了")
	}
	sg := resp.SecurityGroupSet[0]
	d.Set("name", sg.SecurityGroupName)
	d.Set("description", sg.Description)
	return nil
}

func resourceQingcloudSecuritygroupDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).securitygroup
	params := securitygroup.DeleteSecurityGroupsRequest{}
	params.SecurityGroupsN.Add(d.Id())
	_, err := clt.DeleteSecurityGroups(params)
	if err != nil {
		return err
	}
	return nil
}

func resourceQingcloudSecuritygroupUpdate(d *schema.ResourceData, meta interface{}) error {
	if !d.HasChange("name") && !d.HasChange("description") {
		return nil
	}

	clt := meta.(*QingCloudClient).securitygroup
	params := securitygroup.ModifySecurityGroupAttributesRequest{}
	if d.HasChange("description") {
		params.Description.Set(d.Get("description").(string))
	}
	if d.HasChange("name") {
		params.SecurityGroupName.Set(d.Get("name").(string))
	}

	params.SecurityGroup.Set(d.Id())

	_, err := clt.ModifySecurityGroupAttributes(params)
	if err != nil {
		return err
	}
	return nil
}
