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
			schemaProvider: {
				Type:     schema.TypeString,
				Optional: true,
			},
			schemaEnvironment: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"id": {
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

	zone, err := createZoneFromData(pinto, d)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Pinto: Read Zone %s at %s for %s \n", zone.name, zone.provider, zone.environment)

	request := pinto.client.ZonesApi.ApiDnsZonesZoneGet(pctx, zone.name).Provider(zone.provider)
	if zone.environment != "" {
		request = request.Environment(zone.environment)
	}
	_, resp, err := request.Execute()
	if err.Error() != "" {
		return diag.Errorf(handleClientError("[DS] ZONE READ", err.Error(), resp))
	}
	d.SetId(computeZoneId(zone))

	return diags
}
