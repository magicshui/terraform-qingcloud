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
		Update: nil,
		Delete: resourceQingcloudVolumeAttachmentDelete,
		Schema: map[string]*schema.Schema{
			"instance": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"volume": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
	if err != nil {
		return err
	}
	return nil
}

func resourceQingcloudVolumeAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).volume
	params := volume.DescribeVolumesRequest{}
	params.VolumesN.Add(d.Get("volume_id").(string))
	params.Verbose.Set(1)
	resp, err := clt.DescribeVolumes(params)
	if err != nil {
		return fmt.Errorf("Error read volume %s", err)
	}
	if len(resp.VolumeSet) == 0 {
		return fmt.Errorf("Error no volume: %s", d.Id())
	}
	k := resp.VolumeSet[0]
	d.Set("instance", k.Instance.InstanceID)
	d.Set("volume", k.VolumeID)
	d.SetId(k.VolumeID)
	return nil
}

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
	if err != nil {
		return fmt.Errorf("Error waiting for volume (%s) to update: %s", d.Id(), err)
	}
	return nil
}

func volumeAttachmentID(d *schema.ResourceData) string {
	return fmt.Sprintf("%s-%s", d.Get("instance_id").(string), d.Get("volume_id").(string))
}
