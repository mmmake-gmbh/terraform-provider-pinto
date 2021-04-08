package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/whizus/go-stackit"
	"log"
	"strings"
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func computeZoneId(zoneName string, environment string, provider string) string {
	z := strings.TrimSuffix(zoneName, ".") + "."
	return z + environment + "." + provider + "."
}

func createZone(pinto *PintoProvider, ctx context.Context, zone string) error {
	log.Printf("[INFO] Pinto: Creating zone %s in environment %s of provider %s", zone, pinto.environment, pinto.provider)
	request := pinto.client.ZonesApi.ApiDnsZonesPost(ctx).CreateZoneRequestModel(stackit.CreateZoneRequestModel{
		Provider:    pinto.provider,
		Environment: *stackit.NewNullableString(&pinto.environment),
		Name:        zone,
	})
	_, resp, gErr := request.Execute()
	if gErr.Error() != "" {
		handleClientError("ZONE CREATE", gErr.Error(), resp)
		return fmt.Errorf(gErr.Error())
	}
	return nil
}

func resourceDnsZoneCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(pctx, stackit.ContextAPIKeys, pinto.apiKey)
	}
	zone := d.Get("name").(string)
	err := createZone(pinto, pctx, zone)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(computeZoneId(zone, pinto.environment, pinto.provider))

	return diags
}

func resourceDnsZoneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(pctx, stackit.ContextAPIKeys, pinto.apiKey)
	}

	zone := d.Get("name").(string)
	log.Printf("[INFO] Pinto: Read Zone %s of environment %s for provider %s \n", zone, pinto.provider, pinto.environment)

	z, resp, gErr := pinto.client.ZonesApi.ApiDnsZonesZoneGet(pctx, zone).Provider(pinto.provider).Environment(pinto.environment).Execute()
	if gErr.Error() != "" {
		handleClientError("ZONE READ", gErr.Error(), resp)
		return diag.Errorf(gErr.Error())
	}
	e := d.Set("name", z.Name)
	if e != nil {
		return diag.FromErr(e)
	}

	return diags
}

func deleteZone(pinto *PintoProvider, ctx context.Context, zone string) error {
	log.Printf("[INFO] Pinto: Deleting zone %s in environment %s of provider %s", zone, pinto.environment, pinto.provider)
	resp, err := pinto.client.ZonesApi.ApiDnsZonesZoneDelete(ctx, zone).Provider(pinto.provider).Environment(pinto.environment).Execute()
	if err.Error() != "" {
		handleClientError("ZONE DELETE", err.Error(), resp)
		return fmt.Errorf(err.Error())
	}
	return nil
}

func resourceDnsZoneDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(pctx, stackit.ContextAPIKeys, pinto.apiKey)
	}
	zone := d.Get("name").(string)
	err := deleteZone(pinto, pctx, zone)
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
		pctx = context.WithValue(pctx, stackit.ContextAPIKeys, pinto.apiKey)
	}

	log.Printf("[INFO] Pinto: Updating zone %s in environment %s of provider %s", d.Id(), pinto.environment, pinto.provider)
	//TODO: pinto api does not support an update of zones at the moment; instead we have to delete and create the zone
	oldZone, newZone := d.GetChange("name")
	err := deleteZone(pinto, ctx, oldZone.(string))
	if err != nil {
		return diag.FromErr(err)
	}
	err = createZone(pinto, ctx, newZone.(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(computeZoneId(newZone.(string),pinto.environment,pinto.provider))
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
	zoneName := ""
	for i := 0; i <= lastSplice; i++ {
		zoneName = zoneName + zoneSplices[i] + "."
	}
	log.Printf("[DEBUG] Pinto: ZoneName = %s", zoneName)
	err := d.Set("name", zoneName)
	if err != nil {
		return nil, err
	}
	d.SetId(computeZoneId(zoneName, pinto.environment, pinto.provider))

	return []*schema.ResourceData{d},nil
}