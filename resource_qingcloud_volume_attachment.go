package qingcloud

import (
	"log"

	"fmt"
	"strings"

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

// ID 的计算方式
func volumeAttachmentID(d *schema.ResourceData) string {
	return fmt.Sprintf("%s*%s", d.Get("volume").(string), d.Get("instance").(string))
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
	d.SetId(volumeAttachmentID(d))

	// 这里选择使用这个磁盘 ID
	_, err = VolumeTransitionStateRefresh(clt, d.Get("volume").(string))
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

	_, err = VolumeTransitionStateRefresh(clt, d.Get("volume").(string))
	return err
}

// NOTE: 磁盘将只会卸载而不会删除，防止误删除的情况发生
func resourceQingcloudVolumeAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).volume

	params := volume.DetachVolumesRequest{}
	// 这个时候不能这样计算了
	params.Instance.Set(d.Get("instance").(string))
	params.VolumesN.Add(d.Get("volume").(string))

	log.Printf("volume attachment id: %s", d.Id())

	// volumeID := strings.Split(d.Id(), "*")[0]
	// instanceID := strings.Split(d.Id(), "*")[1]
	// params.Instance.Set(instanceID)
	// params.VolumesN.Add(volumeID)

	// 解绑磁盘
	_, err := clt.DetachVolumes(params)
	if err != nil {
		// 这个防止界面上删除的情况发生
		if strings.Contains(err.Error(), "not attached") {
			return nil
		}
		return err
	}

	// 注意 ID
	_, err = VolumeTransitionStateRefresh(clt, d.Get("volume").(string))
	return err
}
