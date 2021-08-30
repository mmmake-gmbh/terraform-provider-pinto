package pinto

import (
	"context"
	"log"

	"github.com/camaoag/project-pinto-sdk-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDnsRecord() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDnsRecordRead,
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
			"zone": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"class": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ttl": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"data": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func recordToRecord(r gopinto.Record, zone string, environment string, provider string) Record {
	var record Record
	record.zone = zone
	record.Name = r.Name
	record.Type = r.Type
	record.Data = r.Data
	record.Class = r.Class
	record.Ttl = r.Ttl
	record.provider = provider
	record.environment = environment
	return record
}

func dataSourceDnsRecordRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(pctx, gopinto.ContextAPIKeys, pinto.apiKey)
	}

	environment := getEnvironment(pinto, d)
	provider, err := getProvider(pinto, d)
	zone := d.Get("zone").(string)
	name := d.Get("name").(string)
	_type := d.Get("type").(string)

	log.Printf("[INFO] Pinto: Read record for name=%s, zone=%s, type=%s, provider=%s, environment=%s",
		name, zone, _type, provider, environment)

	if err != nil {
		return diag.FromErr(err)
	}

	request := pinto.client.RecordsApi.DnsApiRecordsGet(pctx).
		Zone(zone).
		Name(name).
		RecordType(gopinto.RecordType(_type))

	r, resp, gErr := request.Execute()

	if resp.StatusCode >= 400 {
		return diag.Errorf(handleClientError("[DS] RECORD READ", gErr.Error(), resp))
	}
	if len(r) > 1 {
		return diag.Errorf("Cannot uniquely identify a resource with (name=%s, zone=%s, type=%s, provider=%s, environment=%s). "+
			"Wanted 1, got %d", name, zone, _type, pinto.provider, pinto.environment, len(r))
	}

	record := recordToRecord(r[0], zone, environment, provider)
	record.id = computeRecordId(record)

	d.SetId(record.id)
	err = d.Set("name", record.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("zone", record.zone)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("data", record.Data)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("ttl", *record.Ttl)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("class", record.Class)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("type", string(record.Type))
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
