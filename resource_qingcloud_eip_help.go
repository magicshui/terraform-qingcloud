package qingcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/eip"
)

func modifyEipAttributes(d *schema.ResourceData, meta interface{}, create bool) error {
	if create && d.Get("description").(string) == "" {
		return nil
	}

	if !d.HasChange("description") && !d.HasChange("name") {
		return nil
	}

	clt := meta.(*QingCloudClient).eip
	modifyAtrributes := eip.ModifyEipAttributesRequest{}

	modifyAtrributes.Eip.Set(d.Id())
	modifyAtrributes.Description.Set(d.Get("description").(string))
	modifyAtrributes.EipName.Set(d.Get("name").(string))

	_, err := clt.ModifyEipAttributes(modifyAtrributes)
	return err
}
