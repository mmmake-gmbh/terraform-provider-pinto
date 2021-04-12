package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/whizus/go-stackit"
	"log"
	"strings"
)

func dataSourceDnsRecords() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDnsRecordsRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Required: true,
			},
			"record_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"records": {
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
						"type": {
							Type:     schema.TypeString,
							Computed: true,
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
				},
			},
		},
	}
}

func dataSourceDnsRecordsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(pctx, stackit.ContextAPIKeys, pinto.apiKey)
	}

	zone := d.Get("zone").(string)
	log.Printf("[INFO] Pinto: Read records from zone %s at %s for %s \n", zone, pinto.provider, pinto.environment)

	request := pinto.client.RecordsApi.ApiDnsRecordsGet(pctx).Environment(pinto.environment).Provider(pinto.provider).Zone(zone)
	val, ok := d.GetOk("record_type")
	if ok {
		request.RecordType(stackit.RecordType(val.(string)))
	}
	val, ok = d.GetOk("name")
	if ok {
		request.Name(val.(string))
	}

	rrecords, resp, err := request.Execute()
	if err.Error() != "" {
		handleClientError("[DS] RECORD READ", err.Error(), resp)
		return diag.Errorf(err.Error())
	}

	records := make([]interface{}, len(rrecords), len(rrecords))
	for i, r := range rrecords {
		idRecord := recordToRecord(r, zone)
		idRecord.id = computeRecordId(pinto.provider, pinto.environment, idRecord)
		record := make(map[string]interface{})
		record["name"] = r.Name
		record["type"] = r.Type
		record["class"] = r.Class
		record["ttl"] = r.Ttl
		record["data"] = r.Data
		record["id"] = idRecord.id
		records[i] = record
	}

	zoneId := strings.TrimSuffix(zone, ".") + "."
	d.SetId(zoneId + pinto.environment + "." + pinto.provider + ".")
	e := d.Set("records", records)
	if e != nil {
		return diag.FromErr(err)
	}

	return diags
}
