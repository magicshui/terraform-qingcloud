package qingcloud

import (
	"github.com/hashicorp/terraform/helper/mutexkv"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "青云的 ID ",
			},
			"secret": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "青云的密钥",
			},
			"zone": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "青云的 zone ",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"qingcloud_eip":           resourceQingcloudEip(),
			"qingcloud_eip_associate": resourceQingcloudEipAssociate(),

			"qingcloud_keypair":        resourceQingcloudKeypair(),
			"qingcloud_keypair_attach": resourceQingcloudKeypairAttach(),

			"qingcloud_securitygroup":      resourceQingcloudSecuritygroup(),
			"qingcloud_securitygroup_rule": resourceQingcloudSecuritygroupRule(),

			"qingcloud_vxnet": resourceQingcloudVxnet(),

			"qingcloud_router":              resourceQingcloudRouter(),
			"qingcloud_router_static":       resourceQingcloudRouterStatic(),
			"qingcloud_router_static_entry": resourceQingcloudRouterStaticEntry(),

			"qingcloud_instance": resourceQingcloudInstance(),

			"qingcloud_cache": resourceQingcloudCache(),

			"qingcloud_mongo": resourceQingcloudMongo(),

			"qingcloud_volume":            resourceQingcloudVolume(),
			"qingcloud_volume_attachment": resourceQingcloudVolumeAttachment(),

			"qingcloud_loadbalancer":             resourceQingcloudLoadbalancer(),
			"qingcloud_loadbalancer_listener":    resourceQingcloudLoadbalancerListener(),
			"qingcloud_loadbalancer_backend":     resourceQingcloudLoadbalancerBackend(),
			"qingcloud_loadbalancer_policy":      resourceQingcloudLoadbalancerPloicy(),
			"qingcloud_loadbalancer_policy_rule": resourceQingcloudLoadbalancerPloicyRule(),

			"qingcloud_server_certificate": resourceQingcloudServerCertificate(),
		},
		ConfigureFunc: providerConfigure,
	}
}

var qingcloudMutexKV = mutexkv.NewMutexKV()

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		ID:     d.Get("id").(string),
		Secret: d.Get("secret").(string),
		Zone:   d.Get("zone").(string),
	}
	return config.Client()
}
