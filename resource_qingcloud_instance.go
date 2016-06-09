package qingcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/instance"
)

func resourceQingcloudInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudInstanceCreate,
		Read:   resourceQingcloudInstanceRead,
		Update: resourceQingcloudInstanceUpdate,
		Delete: resourceQingcloudInstanceDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "主机名称",
			},
			"image": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ImageID:  "镜像ID",
			},
			"type": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "主机类型",
			},
			"class": &schema.Schema{
				Type:     schema.TypeString,
				Default:  "0",
				Optional: true,
				ForceNew: true,
				Description: "主机性能类型: 性能型:0 ,超高性能型:1	",
				ValidateFunc: withinArrayString("0", "1"),
			},
			"vxnet": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"security_group": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"keypairs": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},

			// 如下是计算处理的结果，不需要手工设置
			"private_ip": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"eip_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"eip_addr": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceQingcloudInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).instance

	params := instance.RunInstancesRequest{}
	params.InstanceName.Set(d.Get("name").(string))
	params.ImageId.Set(d.Get("image").(string))
	params.InstanceType.Set(d.Get("type").(string))
	params.VxnetsN.Add(d.Get("vxnet").(string))
	params.SecurityGroup.Set(d.Get("security_group").(string))
	params.InstanceClass.Set(d.Get("class").(string))

	// 设置登陆的密钥
	// 这个地方需要确认一下，就是如果以后这个值变化了，那么是否需要保留？
	params.LoginMode.Set("keypair")
	for _, kp := range d.Get("keypairs").(*schema.Set).List() {
		params.LoginKeypair.Set(kp.(string))
	}

	resp, err := clt.RunInstances(params)
	if err != nil {
		return fmt.Errorf("Error run instance :%s", err)
	}
	d.SetId(resp.Instances[0])

	// 等机器完成配置
	if _, err := InstanceTransitionStateRefresh(clt, d.Id()); err != nil {
		return err
	}
	return resourceQingcloudInstanceRead(d, meta)
}

func resourceQingcloudInstanceRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).instance

	params := instance.DescribeInstanceRequest{}
	params.InstancesN.Add(d.Id())
	params.Verbose.Set(1)
	resp, err := clt.DescribeInstances(params)
	if err != nil {
		return fmt.Errorf("[ERROR] Descirbe Instance :%s", err)
	}

	if len(resp.InstanceSet) == 0 {
		return fmt.Errorf("[ERROR] Instance: %s not found", d.Id())
	}
	if len(resp.InstanceSet) == 0 {
		return fmt.Errorf("[ERROR] Instance: %s Vxnet: %s not found", d.Id(), d.Get("vxnet").(string))
	}

	k := resp.InstanceSet[0]

	// TODO: not setting the default value
	d.Set("type", k.InstanceType)
	d.Set("class", k.InstanceClass)
	d.Set("name", k.InstanceName)
	d.Set("image", k.Image.ImageID)

	d.Set("vxnet", k.Vxnets[0].VxnetID)
	d.Set("vxnet_name", k.Vxnets[0].VxnetName)
	d.Set("private_ip", k.Vxnets[0].PrivateIP)

	// 可能有
	d.Set("eip_id", k.Eip.EipID)
	d.Set("eip_addr", k.Eip.EipAddr)
	var keypairs = []schema.NewSet(f, items)
	for i := 0; i < len(k.KeypairIds); i++ {
		d.Set("keypairs", value)
	}
	return nil
}

func resourceQingcloudInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceQingcloudInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).instance

	params := instance.StopInstancesRequest{}
	params.InstancesN.Add(d.Id())
	params.Force.Set(1)

	_, err := clt.StopInstances(params)
	if err != nil {
		return fmt.Errorf("Error run instance :%s", err)
	}
	return nil
}
