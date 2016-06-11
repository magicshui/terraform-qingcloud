package qingcloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/volume"
)

func resourceQingcloudVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudVolumeCreate,
		Read:   resourceQingcloudVolumeRead,
		Update: resourceQingcloudVolumeUpdate,
		Delete: resourceQingcloudVolumeDelete,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"size": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				Description: "硬盘容量，目前可创建最小 10G，最大 500G 的硬盘， 在此范围内的容量值必须是 10 的倍数	",
			},
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
			"type": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				ForceNew: true,
				Description: `性能型是 0
					超高性能型是 3 (只能与超高性能主机挂载，目前只支持北京2区)，
					容量型因技术升级过程中，在各区的 type 值略有不同:
					  北京1区，亚太1区：容量型是 1
					  北京2区，广东1区：容量型是 2`,
			},
		},
	}
}

func changeQingcloudVolumeAttributes(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).volume
	modifyParams := volume.ModifyVolumeAttributesRequest{}
	modifyParams.Volume.Set(d.Id())
	modifyParams.Description.Set(d.Get("description").(string))
	modifyParams.VolumeName.Set(d.Get("name").(string))
	_, err := clt.ModifyVolumeAttributes(modifyParams)
	if err != nil {
		return err
	}
	_, err = VolumeTransitionStateRefresh(meta.(*QingCloudClient).volume, d.Id())
	return err
}

func resourceQingcloudVolumeCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).volume
	params := volume.CreateVolumesRequest{}
	params.Size.Set(d.Get("size").(int))
	params.VolumeName.Set(d.Get("name").(string))
	params.VolumeType.Set(d.Get("type").(int))
	resp, err := clt.CreateVolumes(params)
	if err != nil {
		return err
	}
	if len(resp.Volumes) != 1 {
		return fmt.Errorf("volumes response is not 1")
	}
	d.SetId(resp.Volumes[0])

	if err := changeQingcloudVolumeAttributes(d, meta); err != nil {
		return err
	}

	_, err = VolumeTransitionStateRefresh(clt, d.Id())
	return err
}

func resourceQingcloudVolumeRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).volume
	params := volume.DescribeVolumesRequest{}
	params.VolumesN.Add(d.Id())
	params.Verbose.Set(1)
	resp, err := clt.DescribeVolumes(params)
	if err != nil {
		return err
	}

	if len(resp.VolumeSet) == 0 {
		return fmt.Errorf("硬盘已经被删除")
	}

	d.Set("name", resp.VolumeSet[0].VolumeName)
	d.Set("description", resp.VolumeSet[0].Description)
	d.Set("size", resp.VolumeSet[0].Size)

	return nil
}

// TODO: 当更改磁盘大小的时候，如果同时取消了关联，那么将会导致首先取消关联，然后大小改变以后，会再关联上，而不会取消
// 所以取消关联的操作，需要分步进行？
func resourceQingcloudVolumeUpdate(d *schema.ResourceData, meta interface{}) error {
	if !d.HasChange("size") && !d.HasChange("description") && d.HasChange("name") {
		return nil
	}

	clt := meta.(*QingCloudClient).volume
	if d.HasChange("size") {
		// 如果磁盘的大小改变了

		// TODO: 如果其他的状态怎么整？
		// 判断有没有加载到主机？
		params := volume.DescribeVolumesRequest{}
		params.VolumesN.Add(d.Id())
		params.Verbose.Set(1)
		resp, err := clt.DescribeVolumes(params)
		if err != nil {
			return err
		}
		if len(resp.VolumeSet) == 0 {
			return fmt.Errorf("资源已经被删除")
		}
		v := resp.VolumeSet[0]

		// 如果正在使用中，首先需要从主机上卸载磁盘
		if v.Status == "in-use" {
			// 卸载磁盘
			params := volume.DetachVolumesRequest{}
			params.VolumesN.Add(d.Id())
			params.Instance.Set(v.Instance.InstanceID)
			_, err := clt.DetachVolumes(params)
			if err != nil {
				return err
			}
			// 等待磁盘状态稳定下来
			_, err = VolumeTransitionStateRefresh(clt, d.Id())
			if err != nil {
				return err
			}
		}

		// 之前 > 现在，那么就不执行操作，错误提示：磁盘大小只能变大，不能缩小
		oldSize, newSize := d.GetChange("size")
		if oldSize.(int) > newSize.(int) {
			d.Set("size", oldSize.(int))
			return fmt.Errorf("Error you can only increase the size", errors.New("INCREASE SIZE ONLY"))
		}

		params2 := volume.ResizeVolumesRequest{}
		params2.VolumesN.Add(d.Id())
		params2.Size.Set(d.Get("size").(int))
		_, err = clt.ResizeVolumes(params2)
		if err != nil {
			return err
		}

		// 等待磁盘状态稳定下来
		_, err = VolumeTransitionStateRefresh(clt, d.Id())
		if err != nil {
			return err
		}

		// 如果磁盘在使用中
		if v.Status == "in-use" {
			params := volume.AttachVolumesRequest{}
			params.VolumesN.Add(d.Id())
			params.Instance.Set(v.Instance.InstanceID)
			_, err := clt.AttachVolumes(params)
			if err != nil {
				return err
			}
			// 等待磁盘稳定下来
			_, err = VolumeTransitionStateRefresh(clt, d.Id())
			if err != nil {
				return err
			}
		}
	}

	// 其他信息变化
	if d.HasChange("description") || d.HasChange("name") {
		if err := changeQingcloudVolumeAttributes(d, meta); err != nil {
			return err
		}
	}
	_, err := VolumeTransitionStateRefresh(clt, d.Id())
	return err
}

func resourceQingcloudVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).volume
	params := volume.DescribeVolumesRequest{}
	params.VolumesN.Add(d.Id())
	params.Verbose.Set(1)
	resp, err := clt.DescribeVolumes(params)
	if err != nil {
		return err
	}
	if len(resp.VolumeSet) == 0 {
		return fmt.Errorf("资源已经被删除")
	}
	v := resp.VolumeSet[0]

	// 如果在使用中
	if v.Status == "in-use" {
		params := volume.DetachVolumesRequest{}
		params.VolumesN.Add(d.Id())
		params.Instance.Set(v.Instance.InstanceID)
		_, err := clt.DetachVolumes(params)
		if err != nil {
			return err
		}
		// 等待
		_, err = VolumeTransitionStateRefresh(clt, d.Id())
		return err
	}

	// TODO: 以后再删除资源
	// params := volume.DeleteVolumesRequest{}
	// params.VolumesN.Add(d.Id())
	// _, err := clt.DeleteVolumes(params)
	// if err != nil {
	// 	return fmt.Errorf(
	// 		"Error deleting volume: %s", err)
	// }

	_, err = VolumeTransitionStateRefresh(clt, d.Id())
	return err
}
