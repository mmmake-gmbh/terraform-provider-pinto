package pinto

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/whizus/gopinto"
	"log"
	"strings"
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
			schemaProvider: {
				Type:     schema.TypeString,
				Optional: true,
			},
			schemaEnvironment: {
				Type:     schema.TypeString,
				Optional: true,
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
	gopinto.Record

	zone        string
	id          string
	environment string
	provider    string
}

func printDebugRecord(record Record) {
	log.Printf("[DEBUG] Pinto: Record{Id=%s, Name=%s, Zone=%s, Class=%s, Data=%s, Type=%s}",
		record.id, record.Name, record.zone, record.Class, record.Data, record.Type)
}

func computeRecordId(record Record) string {
	idString := record.Data + "-" + string(record.Type) + "." + record.Name + "." + record.zone + record.environment + "." + record.provider + "."
	h := sha1.New()
	h.Write([]byte(idString))
	return hex.EncodeToString(h.Sum(nil))
}

func dataToRecord(d *schema.ResourceData, provider *PintoProvider) (Record, error) {
	var record Record
	s, err := getProvider(provider, d)
	if err != nil {
		return record, err
	} else {
		record.provider = s
	}
	record.environment = getEnvironment(provider, d)
	record.zone = d.Get("zone").(string)
	record.Name = d.Get("name").(string)
	record.Type = gopinto.RecordType(d.Get("type").(string))
	record.Data = d.Get("data").(string)
	record.Class = d.Get("class").(string)
	_, ok := d.GetOk("ttl")
	if ok {
		ttl := int64(d.Get("ttl").(int))
		record.Ttl = &ttl
	}
	return record, nil
}

func createRecord(client *gopinto.APIClient, ctx context.Context, record Record) error {
	log.Printf("[DEBUG] Pinto: Creating Record:")
	printDebugRecord(record)
	crr := gopinto.NewCreateRecordRequestModel(record.provider, record.zone, record.Name, record.Type, record.Data)
	if record.environment != "" {
		crr.SetEnvironment(record.environment)
	}
	crr.SetClass(gopinto.RecordClass(record.Class))
	crr.SetTtl(int32(*record.Ttl))
	_, resp, gErr := client.RecordsApi.ApiDnsRecordsPost(ctx).CreateRecordRequestModel(*crr).Execute()
	if gErr.Error() != "" {
		return fmt.Errorf(handleClientError("RECORD CREATE", gErr.Error(), resp))
	}
	return nil
}

func resourceDnsRecordCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(ctx, gopinto.ContextAPIKeys, pinto.apiKey)
	}

	record, err := dataToRecord(d, pinto)
	if err != nil {
		return diag.FromErr(err)
	}
	record.id = computeRecordId(record)
	log.Printf("[INFO] Pinto: Creating record %s in environment %s of provider %s", record.id, pinto.environment, pinto.provider)
	if !record.HasTtl() {
		// if no TTL is set, then we use the default value 3600
		ttl64 := int64(3600)
		record.Ttl = &ttl64
	}
	err = createRecord(pinto.client, pctx, record)
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
		pctx = context.WithValue(pctx, gopinto.ContextAPIKeys, pinto.apiKey)
	}

	record, err := dataToRecord(d, pinto)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Pinto: Reading information for record with name %s in environment %s of provider %s", record.Name+"."+record.zone,
		pinto.environment, pinto.provider)
	log.Printf("[DEBUG] Pinto: Reading Record:")
	printDebugRecord(record)
	request := pinto.client.RecordsApi.ApiDnsRecordsGet(pctx).Name(record.Name).RecordType(record.Type).Zone(record.zone).Provider(record.provider)
	if record.environment != "" {
		request = request.Environment(record.environment)
	}
	r, resp, gErr := request.Execute()

	if resp.StatusCode >= 400 {
		return diag.Errorf(handleClientError("RECORD READ", gErr.Error(), resp))
	}
	record.id = computeRecordId(record)
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

func deleteRecord(client *gopinto.APIClient, ctx context.Context, record Record) error {
	log.Printf("[INFO] Pinto: Deleting record with id %s in environment %s of provider %s", record.id, record.environment, record.provider)
	log.Printf("[DEBUG] Pinto: Working in env %s of pinto %s", record.environment, record.provider)
	log.Printf("[DEBUG] Pinto: Deleting Record:")
	printDebugRecord(record)
	rBody := make(map[string]string)

	request := client.RecordsApi.ApiDnsRecordsDelete(ctx).
		Zone(record.zone).
		Provider(record.provider).
		Name(record.Name).
		RecordType(record.Type).
		RequestBody(rBody)
	if record.environment != "" {
		request = request.Environment(record.environment)
	}
	resp, err := request.Execute()
	if err.Error() != "" {
		return fmt.Errorf(handleClientError("RECORD DELETE", err.Error(), resp))
	}
	return nil
}

func resourceDnsRecordDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(pctx, gopinto.ContextAPIKeys, pinto.apiKey)
	}

	record, err := dataToRecord(d, pinto)
	if err != nil {
		return diag.FromErr(err)
	}
	record.id = d.Id()
	err = deleteRecord(pinto.client, pctx, record)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func buildRecordsFromChange(p *PintoProvider, d *schema.ResourceData) (Record, Record, error) {
	r, err := dataToRecord(d, p)
	if err != nil {
		return r, r, err
	}
	newRecord := r
	r, err = dataToRecord(d, p)
	if err != nil {
		return r, r, err
	}
	oldRecord := r
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
		oldRecord.Type = gopinto.RecordType(o.(string))
		newRecord.Type = gopinto.RecordType(n.(string))
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

	return oldRecord, newRecord, nil
}

func resourceDnsRecordUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pinto := m.(*PintoProvider)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(pctx, gopinto.ContextAPIKeys, pinto.apiKey)
	}

	log.Printf("[INFO] Pinto: Updating record with id %s in environment %s of provider %s", d.Id(), pinto.environment, pinto.provider)
	//TODO: pinto api does not support an update of Records at the moment; instead we have to delete and create the Record
	oldRecord, newRecord, err := buildRecordsFromChange(pinto, d)
	if err != nil {
		return diag.FromErr(err)
	}
	err = deleteRecord(pinto.client, pctx, oldRecord)
	if err != nil {
		return diag.FromErr(err)
	}
	err = createRecord(pinto.client, pctx, newRecord)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceDnsRecordImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	pinto := m.(*PintoProvider)

	pctx := ctx
	if pinto.apiKey != "" {
		pctx = context.WithValue(pctx, gopinto.ContextAPIKeys, pinto.apiKey)
	}

	in := strings.Split(d.Id(), "/")
	if len(in) != 5 {
		return nil, fmt.Errorf("invalid Import. ID has to be of format \"{type}/{name}/{zone}/{environment}/{provider}\"")
	}

	// setting all information in a record var to perform the id calculation below
	var record Record
	record.Type = gopinto.RecordType(in[0])
	record.Name = in[1]
	record.zone = in[2]
	record.environment = in[3]
	record.provider = in[4]

	log.Printf("[DEBUG] retrieving information for %s", d.Id())
	request := pinto.client.RecordsApi.ApiDnsRecordsGet(pctx).Name(record.Name).RecordType(record.Type).Zone(record.zone).
		Provider(record.provider)
	if record.environment != "" {
		request = request.Environment(record.environment)
	}
	r, resp, gErr := request.Execute()
	if gErr.Error() != "" {
		return nil, fmt.Errorf(handleClientError("IMPORT RECORD", gErr.Error(), resp))
	}
	if len(r) > 1 {
		return nil, fmt.Errorf("invalid Import. More than one record matched ID %s/%s/%s", record.Type, record.Name, record.zone)
	}
	record.Data = r[0].Data
	record.Class = r[0].Class
	record.Ttl = r[0].Ttl
	record.id = computeRecordId(record)

	// add gathered info to ResourceData
	d.SetId(record.id)
	err := d.Set(schemaProvider, record.provider)
	if err != nil {
		return nil, err
	}
	err = d.Set(schemaEnvironment, record.environment)
	if err != nil {
		return nil, err
	}
	err = d.Set("name", record.Name)
	if err != nil {
		return nil, err
	}
	err = d.Set("zone", record.zone)
	if err != nil {
		return nil, err
	}
	err = d.Set("data", record.Data)
	if err != nil {
		return nil, err
	}
	err = d.Set("ttl", *record.Ttl)
	if err != nil {
		return nil, err
	}
	err = d.Set("class", record.Class)
	if err != nil {
		return nil, err
	}
	err = d.Set("type", string(record.Type))
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
