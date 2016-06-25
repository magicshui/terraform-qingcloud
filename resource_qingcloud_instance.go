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
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "描述信息",
			},
			"image": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "映像ID，此映像将作为主机的模板。可传青云提供的映像ID，或自己创建的映像ID	",
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

			// TODO: 加入 terminate 属性？
			"init_keypair": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "初始化的密钥",
			},
		},
	}
}

func resourceQingcloudInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).instance

	// TODO: 判断当前的主机是否是已经关闭的状态？

	params := instance.RunInstancesRequest{}
	params.InstanceName.Set(d.Get("name").(string))
	params.ImageID.Set(d.Get("image").(string))
	params.InstanceType.Set(d.Get("type").(string))
	params.InstanceClass.Set(d.Get("class").(string))
	params.LoginKeypair.Set(d.Get("init_keypair").(string))
	params.LoginMode.Set("keypair")

	resp, err := clt.RunInstances(params)
	if err != nil {
		return err
	}
	if len(resp.Instances) != 1 {
		return fmt.Errorf("[QC] instance error,not 1, real is : %d", len(resp.Instances))
	}

	d.SetId(resp.Instances[0])

	// 等机器完成配置
	_, err = InstanceTransitionStateRefresh(clt, d.Id())
	if err != nil {
		return err
	}

	// 更新description
	if d.Get("description").(string) != "" {
		params := instance.ModifyInstanceAttributesRequest{}
		params.Instance.Set(d.Id())
		params.InstanceName.Set(d.Get("name").(string))
		params.Description.Set(d.Get("description").(string))
		_, err := clt.ModifyInstanceAttributes(params)
		if err != nil {
			return err
		}
		// 等机器完成配置
		_, err = InstanceTransitionStateRefresh(clt, d.Id())
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceQingcloudInstanceRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).instance

	params := instance.DescribeInstanceRequest{}
	params.InstancesN.Add(d.Id())
	params.Verbose.Set(1)
	resp, err := clt.DescribeInstances(params)
	if err != nil {
		return err
	}

	if len(resp.InstanceSet) == 0 {
		return fmt.Errorf("[ERROR] Instance: %s not found", d.Id())
	}

	k := resp.InstanceSet[0]

	d.Set("type", k.InstanceType)
	d.Set("class", k.InstanceClass)
	d.Set("name", k.InstanceName)
	d.Set("image", k.Image.ImageID)
	d.Set("status", k.Status)
	d.SetId(k.InstanceID)

	return nil
}

// TODO: 如果机器资源更新了？
func resourceQingcloudInstanceUpdate(d *schema.ResourceData, meta interface{}) error {

	clt := meta.(*QingCloudClient).instance
	if d.HasChange("name") || d.HasChange("description") {
		params := instance.ModifyInstanceAttributesRequest{}
		params.Instance.Set(d.Id())
		params.InstanceName.Set(d.Get("name").(string))
		params.Description.Set(d.Get("description").(string))
		_, err := clt.ModifyInstanceAttributes(params)
		return err
	}

	// TODO: 主机运行状态？

	// 改变类型
	if d.HasChange("type") {
		// 如果主机在运行中，首先对主机进行关机操作
		if d.Get("status") == "running" {
			params := instance.StopInstancesRequest{}
			params.InstancesN.Add(d.Id())
			_, err := clt.StopInstances(params)
			if err != nil {
				return err
			}
		}

		// 等待操作完成
		_, err := InstanceTransitionStateRefresh(clt, d.Id())
		if err != nil {
			return err
		}

		params2 := instance.ResizeInstancesRequest{}
		params2.InstancesN.Add(d.Id())
		params2.InstanceType.Set(d.Get("type").(string))
		_, err = clt.ResizeInstances(params2)
		if err != nil {
			return err
		}

		// 开启主机
		params3 := instance.StartInstancesRequest{}
		params3.InstancesN.Add(d.Id())
		_, err = clt.StartInstances(params3)
		if err != nil {
			return err
		}

		// 等待操作完成
		_, err = InstanceTransitionStateRefresh(clt, d.Id())
		if err != nil {
			return err
		}

	}

	if d.HasChange("class") || d.HasChange("image") {
		return fmt.Errorf("Can't change class or type or image")
	}

	// 其他改变不能生效
	return nil
}

func resourceQingcloudInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).instance

	// 删除依赖的资源
	params := instance.StopInstancesRequest{}
	params.InstancesN.Add(d.Id())
	params.Force.Set(1)

	_, err := clt.StopInstances(params)
	if err != nil {
		return err
	}
	_, err = InstanceTransitionStateRefresh(clt, d.Id())
	return err
}
