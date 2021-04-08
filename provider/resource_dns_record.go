package provider

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/whizus/go-stackit"
	"log"
)

func resourceDnsRecord() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDnsRecordCreate,
		ReadContext:   resourceDnsRecordRead,
		DeleteContext: resourceDnsRecordDelete,
		UpdateContext: resourceDnsRecordUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDnsRecordImport,
		},
		Schema: map[string]*schema.Schema{
			"zone": {
				Type: schema.TypeString,
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
				Required: true,
			},
			"ttl": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"data": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

type Record struct {
	stackit.Record

	zone string
	id string
}

func printDebugRecord(record Record) {
	log.Printf("[DEBUG] Pinto: Record{Id=%s, Name=%s, Zone=%s, Class=%s, Data=%s, Type=%s}",
		record.id, record.Name, record.zone, record.Class, record.Data, record.Type)
}

func computeRecordId(provider string, environment string, record Record) string {
	idString := record.Data + "-" + string(record.Type) + "." + record.Name + "." + record.zone + environment + "." + provider + "."
	h := sha1.New()
	h.Write([]byte(idString))
	return hex.EncodeToString(h.Sum(nil))
}

func dataToRecord(d *schema.ResourceData) Record {
	var record Record
	record.zone = d.Get("zone").(string)
	record.Name = d.Get("name").(string)
	record.Type = stackit.RecordType(d.Get("type").(string))
	record.Data = d.Get("data").(string)
	record.Class = d.Get("class").(string)
	_, ok := d.GetOk("ttl")
	if ok {
		ttl := int64(d.Get("ttl").(int))
		record.Ttl = &ttl
	}
	return record
}

func createRecord(pinto *PintoProvider, ctx context.Context, record Record) error {
	log.Printf("[DEBUG] Pinto: Creating Record:")
	printDebugRecord(record)
	crr := stackit.NewCreateRecordRequestModel(pinto.provider,record.zone,record.Name, record.Type,record.Data)
	crr.SetEnvironment(pinto.environment)
	crr.SetClass(stackit.RecordClass(record.Class))
	crr.SetTtl(int32(*record.Ttl))
	_, resp, gErr := pinto.client.RecordsApi.ApiDnsRecordsPost(ctx).CreateRecordRequestModel(*crr).Execute()
	if gErr.Error() != "" {
		handleClientError("RECORD CREATE", gErr.Error(), resp)
		return fmt.Errorf(gErr.Error())
	}
	return nil
}

func resourceDnsRecordCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(pctx, stackit.ContextAPIKeys, pinto.apiKey)
	}

	record := dataToRecord(d)
	record.id = computeRecordId(pinto.provider, pinto.environment, record)
	log.Printf("[INFO] Pinto: Creating record %s in environment %s of provider %s", record.id, pinto.environment, pinto.provider)
	if !record.HasTtl() {
		// if no TTL is set, then we use the default value 3600
		ttl64 := int64(3600)
		record.Ttl = &ttl64
	}
	err := createRecord(pinto, pctx, record)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("ttl", *record.Ttl)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(record.id)

	return diags
}

func resourceDnsRecordRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(pctx, stackit.ContextAPIKeys, pinto.apiKey)
	}

	record := dataToRecord(d)
	log.Printf("[INFO] Pinto: Reading information for record with name %s in environment %s of provider %s", record.Name + "." + record.zone,
		pinto.environment, pinto.provider)
	log.Printf("[DEBUG] Pinto: Reading Record:")
	printDebugRecord(record)
	r, resp, gErr := pinto.client.RecordsApi.ApiDnsRecordsGet(pctx).Name(record.Name).RecordType(record.Type).Zone(record.zone).
		Environment(pinto.environment).Provider(pinto.provider).Execute()

	if gErr.Error() != "" {
		handleClientError("RECORD READ", gErr.Error(), resp)
		return diag.Errorf(gErr.Error())
	}
	record.id = computeRecordId(pinto.provider, pinto.environment, record)
	if len(r) == 0 {
		log.Printf("[WARN] Pinto: Could not retrieve information for pinto_dns_record with id %s. Removing it from state", record.id)
		d.SetId("")
	} else {
		err := d.Set("ttl", *r[0].Ttl)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(record.id)
	}

	return diags
}

func deleteRecord(pinto *PintoProvider, ctx context.Context, record Record) error {
	log.Printf("[INFO] Pinto: Deleting record with id %s in environment %s of provider %s", record.id, pinto.environment, pinto.provider)
	log.Printf("[DEBUG] Pinto: Working in env %s of provider %s", pinto.environment, pinto.provider)
	log.Printf("[DEBUG] Pinto: Deleting Record:")
	printDebugRecord(record)
	rBody := make(map[string]string)

	resp, err := pinto.client.RecordsApi.ApiDnsRecordsDelete(ctx).
		Zone(record.zone).
		Provider(pinto.provider).
		Environment(pinto.environment).
		Name(record.Name).
		RecordType(record.Type).
		RequestBody(rBody).
		Execute()
	if err.Error() != "" {
		handleClientError("RECORD DELETE", err.Error(), resp)
		return fmt.Errorf(err.Error())
	}
	return nil
}

func resourceDnsRecordDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(pctx, stackit.ContextAPIKeys, pinto.apiKey)
	}
	record := dataToRecord(d)
	record.id = d.Id()
	err := deleteRecord(pinto, pctx, record)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func buildRecordsFromChange(d *schema.ResourceData) (Record, Record) {
	newRecord := dataToRecord(d)
	oldRecord := dataToRecord(d)
	newRecord.id = d.Id()
	oldRecord.id = d.Id()
	if d.HasChange("name") {
		o, n := d.GetChange("name")
		oldRecord.Name = o.(string)
		newRecord.Name = n.(string)
	}
	if d.HasChange("zone") {
		o, n := d.GetChange("zone")
		oldRecord.zone = o.(string)
		newRecord.zone = n.(string)
	}
	if d.HasChange("type") {
		o, n := d.GetChange("type")
		oldRecord.Type = stackit.RecordType(o.(string))
		newRecord.Type = stackit.RecordType(n.(string))
	}
	if d.HasChange("class") {
		o, n := d.GetChange("class")
		oldRecord.Class = o.(string)
		newRecord.Class = n.(string)
	}
	if d.HasChange("ttl") {
		o, n := d.GetChange("ttl")
		o64 := int64(o.(int))
		n64 := int64(n.(int))
		oldRecord.Ttl = &o64
		newRecord.Ttl = &n64
	}
	if d.HasChange("data") {
		o, n := d.GetChange("data")
		oldRecord.Data = o.(string)
		newRecord.Data = n.(string)
	}

	return oldRecord, newRecord
}

func resourceDnsRecordUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(pctx, stackit.ContextAPIKeys, pinto.apiKey)
	}

	log.Printf("[INFO] Pinto: Updating record with id %s in environment %s of provider %s", d.Id(), pinto.environment, pinto.provider)
	//TODO: pinto api does not support an update of Records at the moment; instead we have to delete and create the Record
	oldRecord, newRecord := buildRecordsFromChange(d)
	err := deleteRecord(pinto, pctx, oldRecord)
	if err != nil {
		return diag.FromErr(err)
	}
	err = createRecord(pinto, pctx, newRecord)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceDnsRecordImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	//pinto := m.(*PintoProvider)
	return []*schema.ResourceData{d},nil
}
