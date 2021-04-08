package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/whizus/go-stackit"
	"log"
)

func dataSourceDnsZones() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDnsZonesRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDnsZonesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(pctx, stackit.ContextAPIKeys, pinto.apiKey)
	}

	log.Printf("[INFO] Pinto: Read zones at %s for %s \n", pinto.provider, pinto.environment)

	rz, resp, err := pinto.client.ZonesApi.ApiDnsZonesGet(pctx).Provider(pinto.provider).Environment(pinto.environment).Execute()
	if err.Error() != "" {
		handleClientError("[DS] ZONES READ", err.Error(), resp)
		return diag.Errorf(err.Error())
	}

	zones :=  make([]interface{}, len(rz), len(rz))
	for i,z := range rz {
		zone := make(map[string]interface{})
		zone["id"] = computeZoneId(z.Name,pinto.environment,pinto.provider)
		zone["name"] = z.Name
		zones[i] = zone
	}

	d.SetId(pinto.environment + "." +  pinto.provider + ".")
	e := d.Set("zones", zones)
	if e != nil {
		return diag.FromErr(err)
	}

	return diags
}
