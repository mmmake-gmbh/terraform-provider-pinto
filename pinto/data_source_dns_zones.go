package pinto

import (
	"context"
	"log"

	gopinto "github.com/camaoag/project-pinto-sdk-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDnsZones() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDnsZonesRead,
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
		pctx = context.WithValue(pctx, gopinto.ContextAPIKeys, pinto.apiKey)
	}

	log.Printf("[INFO] Pinto: Read zones at %s for %s \n", pinto.provider, pinto.environment)

	request := pinto.client.ZonesApi.DnsApiZonesGet(pctx).XApiOptions(pinto.xApiOptions)
	rz, resp, err := request.Execute()
	if resp.StatusCode >= 400 {
		return diag.Errorf(handleClientError("[DS] ZONES READ", err.Error(), resp))
	}

	zones := make([]interface{}, len(rz), len(rz))
	for i, z := range rz {
		temp := Zone{
			name:        z.Name,
			environment: pinto.environment,
			provider:    pinto.provider,
		}
		zone := make(map[string]interface{})
		zone["id"] = computeZoneId(temp)
		zone["name"] = z.Name
		zones[i] = zone
	}

	d.SetId(pinto.environment + "." + pinto.provider + ".")
	e := d.Set("zones", zones)
	if e != nil {
		return diag.FromErr(err)
	}

	return diags
}
