package pinto

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/whizus/gopinto"
)

func resourceDnsZone() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDnsZoneCreate,
		ReadContext:   resourceDnsZoneRead,
		DeleteContext: resourceDnsZoneDelete,
		UpdateContext: resourceDnsZoneUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDnsZoneImport,
		},
		Schema: map[string]*schema.Schema{
			schemaProvider: {
				Type:     schema.TypeString,
				Optional: true,
			},
			schemaEnvironment: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

type Zone struct {
	name        string
	environment string
	provider    string
}

func createZoneFromData(p *PintoProvider, d *schema.ResourceData) (Zone, error) {
	var zone Zone
	environment := getEnvironment(p, d)
	provider, err := getProvider(p, d)
	if err != nil {
		return zone, err
	}
	zone.provider = provider
	zone.environment = environment

	zone.name = d.Get("name").(string)

	return zone, nil
}

func computeZoneId(zone Zone) string {
	z := strings.TrimSuffix(zone.name, ".") + "."
	return z + zone.environment + "." + zone.provider + "."
}

func createZone(client *gopinto.APIClient, ctx context.Context, zone Zone) error {
	log.Printf("[INFO] Pinto: Creating zone %s in environment %s of provider %s", zone.name, zone.environment, zone.provider)
	request := client.ZonesApi.ApiDnsZonesPost(ctx).CreateZoneRequestModel(gopinto.CreateZoneRequestModel{
		Provider:    zone.provider,
		Environment: *gopinto.NewNullableString(&zone.environment),
		Name:        zone.name,
	})
	_, resp, gErr := request.Execute()
	if gErr.Error() != "" {
		return fmt.Errorf(handleClientError("ZONE CREATE", gErr.Error(), resp))
	}
	return nil
}

func resourceDnsZoneCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	zone, err := createZoneFromData(pinto, d)
	if err != nil {
		return diag.FromErr(err)
	}
	err = createZone(pinto.client, ctx, zone)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(computeZoneId(zone))

	return diags
}

func resourceDnsZoneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(ctx, gopinto.ContextAPIKeys, pinto.apiKey)
	}

	zone := d.Get("name").(string)
	environment := getEnvironment(pinto, d)
	provider, err := getProvider(pinto, d)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Pinto: Read Zone %s of environment %s for provider %s \n", zone, provider, environment)

	request := pinto.client.ZonesApi.ApiDnsZonesZoneGet(pctx, zone).Provider(provider)
	if environment != "" {
		request = request.Environment(environment)
	}
	z, resp, gErr := request.Execute()
	if resp.StatusCode >= 400 {
		return diag.Errorf(handleClientError("ZONE READ", gErr.Error(), resp))
	}
	e := d.Set("name", z.Name)
	if e != nil {
		return diag.FromErr(e)
	}

	return diags
}

func deleteZone(client *gopinto.APIClient, ctx context.Context, zone Zone) error {
	log.Printf("[INFO] Pinto: Deleting zone %s in environment %s of provider %s", zone.name, zone.environment, zone.provider)
	request := client.ZonesApi.ApiDnsZonesZoneDelete(ctx, zone.name).Provider(zone.provider)
	if zone.environment != "" {
		request = request.Environment(zone.environment)
	}
	resp, err := request.Execute()
	if  resp.StatusCode >= 400 {
		return fmt.Errorf(handleClientError("ZONE DELETE", err.Error(), resp))
	}
	return nil
}

func resourceDnsZoneDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(ctx, gopinto.ContextAPIKeys, pinto.apiKey)
	}

	zone, err := createZoneFromData(pinto, d)
	if err != nil {
		return diag.FromErr(err)
	}
	err = deleteZone(pinto.client, pctx, zone)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceDnsZoneUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(ctx, gopinto.ContextAPIKeys, pinto.apiKey)
	}

	log.Printf("[INFO] Pinto: Updating zone %s in environment %s of provider %s", d.Id(), pinto.environment, pinto.provider)
	//TODO: pinto api does not support an update of zones at the moment; instead we have to delete and create the zone
	oldZone, err := createZoneFromData(pinto, d)
	if err != nil {
		return diag.FromErr(err)
	}
	newZone, _ := createZoneFromData(pinto, d)
	oldZoneS, newZoneS := d.GetChange("name")
	oldZone.name = oldZoneS.(string)
	newZone.name = newZoneS.(string)
	err = deleteZone(pinto.client, pctx, oldZone)
	if err != nil {
		return diag.FromErr(err)
	}
	err = createZone(pinto.client, pctx, newZone)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceDnsZoneImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	pinto := m.(*PintoProvider)
	zoneId := d.Id()
	log.Printf("[INFO] Pinto: Importing zone with id %s", zoneId)

	if !strings.Contains(zoneId, pinto.environment) || !strings.Contains(zoneId, pinto.provider) {
		return nil, fmt.Errorf("invalid Import. ID has to be of format \"{zoneName}.{environment}.{provider}.\"")
	}
	zoneSplices := strings.Split(zoneId, ".")
	// -1 because the array is of index [0,..,length-1]
	// -3 because the last three splices contain "environment" "provider" and "" [after last . is nothing]
	lastSplice := len(zoneSplices) - 4
	provider := zoneSplices[len(zoneSplices)-2]
	environment := zoneSplices[len(zoneSplices)-3]
	zoneName := ""
	for i := 0; i <= lastSplice; i++ {
		zoneName = zoneName + zoneSplices[i] + "."
	}
	zone := Zone{
		name:        zoneName,
		environment: environment,
		provider:    provider,
	}
	log.Printf("[DEBUG] Pinto: ZoneName = %s", zoneName)
	err := d.Set("name", zoneName)
	if err != nil {
		return nil, err
	}
	err = d.Set(schemaProvider, provider)
	if err != nil {
		return nil, err
	}
	err = d.Set(schemaEnvironment, environment)
	if err != nil {
		return nil, err
	}
	d.SetId(computeZoneId(zone))

	return []*schema.ResourceData{d}, nil
}
