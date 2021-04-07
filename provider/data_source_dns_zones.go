package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/whizus/go-stackit"
	"log"
)

func dataSourceDnsZone() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDnsZoneRead,
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
						},
			"name": {
				Type:     schema.TypeString,
				Required: true,
						},
		},
	}
}

func dataSourceDnsZoneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(pctx, stackit.ContextAPIKeys, pinto.apiKey)
	}

	zone := d.Get("name").(string)
	log.Printf("[DEBUG] Pinto: Read Zone %s at %s for %s \n", zone, pinto.provider, pinto.environment)

	z, _, err := pinto.client.ZonesApi.ApiDnsZonesZoneGet(pctx, zone).Provider(pinto.provider).Environment(pinto.environment).Execute()
	if err.Error() != "" {
		return diag.Errorf(err.Error())
	}
	d.SetId(z.Name)

	return diags
}
