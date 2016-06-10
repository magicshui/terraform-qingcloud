package qingcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/volume"
)

func resourceQingcloudVolumeAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudVolumeAttachmentCreate,
		Read:   resourceQingcloudVolumeAttachmentRead,
		Update: resourceQingcloudVolumeAttachmentUpdate,
		Delete: resourceQingcloudVolumeAttachmentDelete,
		Schema: map[string]*schema.Schema{
			"instance": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "主机 ID",
			},
			"volume": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "磁盘 ID",
			},

			// 自动计算
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceQingcloudVolumeAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).volume

	params := volume.AttachVolumesRequest{}
	params.Instance.Set(d.Get("instance").(string))
	params.VolumesN.Add(d.Get("volume").(string))
	_, err := clt.AttachVolumes(params)
	if err != nil {
		return err
	}
	d.SetId(d.Get("volume").(string))
	_, err = VolumeTransitionStateRefresh(clt, d.Id())
	return err
}

func resourceQingcloudVolumeAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).volume
	params := volume.DescribeVolumesRequest{}
	params.VolumesN.Add(d.Get("volume").(string))
	params.Verbose.Set(1)
	resp, err := clt.DescribeVolumes(params)
	if err != nil {
		return err
	}

	if len(resp.VolumeSet) == 0 {
		return fmt.Errorf("资源可能已经被删除了")
	}

	k := resp.VolumeSet[0]
	d.Set("instance", k.Instance.InstanceID)
	d.Set("volume", k.VolumeID)
	d.SetId(volumeAttachmentID(d))
	return nil
}

// 硬盘挂载不能更新
func resourceQingcloudVolumeAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceQingcloudVolumeAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).volume

	params := volume.DetachVolumesRequest{}
	params.Instance.Set(d.Get("instance").(string))
	params.VolumesN.Add(d.Get("volume").(string))

	// 解绑磁盘
	_, err := clt.DetachVolumes(params)
	if err != nil {
		return err
	}

	_, err = VolumeTransitionStateRefresh(clt, d.Id())
	return err
}

func volumeAttachmentID(d *schema.ResourceData) string {
	return fmt.Sprintf("%s-%s", d.Get("instance").(string), d.Get("volume").(string))
}
