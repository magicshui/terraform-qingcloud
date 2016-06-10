package qingcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/magicshui/qingcloud-go/eip"
)

func modifyEipAttributes(d *schema.ResourceData, meta interface{}, create bool) error {
	clt := meta.(*QingCloudClient).eip
	modifyAtrributes := eip.ModifyEipAttributesRequest{}
	if create {
		if description := d.Get("description").(string); description == "" {
			return nil
		}
	} else {
		if !d.HasChange("description") && !d.HasChange("name") {
			return nil
		}
	}

	modifyAtrributes.Eip.Set(d.Id())
	modifyAtrributes.Description.Set(d.Get("description").(string))
	modifyAtrributes.EipName.Set(d.Get("name").(string))
	_, err := clt.ModifyEipAttributes(modifyAtrributes)
	if err != nil {
		return fmt.Errorf("Error modify eip description: %s", err)
	}
	return nil
}
