package qingcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/loadbalancer"
)

func resourceQingcloudServerCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudServerCertificateCreate,
		Read:   resourceQingcloudSecuritygroupRead,
		Update: resourceQingcloudServerCertificateUpdate,
		Delete: resourceQingcloudServerCertificateDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "名称",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "介绍",
			},
			"content": &schema.Schema{
				Type:        schema.TypeString,
				Description: "证书内容",
				Required:    true,
			},
			"private_key": &schema.Schema{
				Type:        schema.TypeString,
				Description: "服务器证书私钥",
				Required:    true,
			},
		},
	}
}

func resourceQingcloudServerCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(QingCloudClient).loadbalancer
	params := loadbalancer.CreateServerCertificateRequest{}
	params.CertificateContent.Set(d.Get("content").(string))
	params.PrivateKey.Set(d.Get("private_key").(string))
	params.ServerCertificateName.Set(d.Get("name").(string))
	resp, err := clt.CreateServerCertificate(params)
	if err != nil {
		return err
	}
	d.SetId(resp.ServerCertificateId)

	if d.Get("description") != nil {
		params := loadbalancer.ModifyServerCertificateAttributesRequest{}
		params.Description.Set(d.Get("description").(string))
		params.ServerCertificateName.Set(d.Get("name").(string))
		_, err = clt.ModifyServerCertificateAttributes(params)
	}
	return err
}

func resourceQingcloudServerCertificateRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(QingCloudClient).loadbalancer
	params := loadbalancer.DescribeServerCertificatesRequest{}
	params.ServerCertificates.Set(d.Id())
	resp, err := clt.DescribeServerCertificates(params)
	if err != nil {
		return err
	}
	d.Set("name", resp.ServerCertificateSet[0].ServerCertificateName)
	d.Set("content", resp.ServerCertificateSet[0].CertificateContent)
	d.Set("private_key", resp.ServerCertificateSet[0].PrivateKey)
	return nil

}

func resourceQingcloudServerCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(QingCloudClient).loadbalancer
	params := loadbalancer.ModifyServerCertificateAttributesRequest{}
	params.Description.Set(d.Get("description").(string))
	params.ServerCertificateName.Set(d.Get("name").(string))
	_, err := clt.ModifyServerCertificateAttributes(params)
	return err
}

func resourceQingcloudServerCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(QingCloudClient).loadbalancer
	params := loadbalancer.DeleteServerCertificatesRequest{}
	params.ServerCertificatesN.Add(d.Id())
	_, err := clt.DeleteServerCertificates(params)
	return err
}
