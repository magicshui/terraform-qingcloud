package qingcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	// "github.com/magicshui/qingcloud-go/tag"
)

func tagsSchema() *schema.Schema {
	return &schema.Schema{
		// TypeSet 是什么？
		Type:     schema.TypeSet,
		Optional: true,
	}
}
