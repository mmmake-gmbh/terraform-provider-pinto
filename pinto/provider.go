package pinto

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	gopinto "github.com/camaoag/project-pinto-sdk-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cc "golang.org/x/oauth2/clientcredentials"
)

type IPintoProvider interface{}
type PintoProvider struct {
	IPintoProvider

	client        *gopinto.APIClient
	apiKey        string
	provider      string
	environment   string
	credentialsId string
	xApiOptions   string
}

const (
	schemaBaseUrl       = "base_url"
	schemaTokenUrl      = "token_url"
	schemaClientId      = "client_id"
	schemaClientSecret  = "client_secret"
	schemaClientScope   = "client_scope"
	schemaApiKey        = "api_key"
	schemaCredentialsId = "credentials_id"

	envKeyBaseUrl       = "PINTO_BASE_URL"
	envKeyTokenUrl      = "PINTO_TOKEN_URL"
	envKeyProvider      = "PINTO_PROVIDER"
	envKeyEnvironment   = "PINTO_ENVIRONMENT"
	envKeyApiKey        = "PINTO_API_KEY"
	envKeyClientId      = "PINTO_CLIENT_ID"
	envKeyClientSecret  = "PINTO_CLIENT_SECRET"
	envKeyClientScope   = "PINTO_CLIENT_SCOPE"
	envKeyCredentialsId = "PINTO_CREDENTIALS_ID"
)

func NewDefaultProvider() *schema.Provider {
	return Provider(nil)
}

// Provider -
func Provider(client *gopinto.APIClient) *schema.Provider {
	log.Printf("[DEBUG] Pinto: Starting Provider")
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			schemaProvider: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyProvider, nil),
			},
			schemaEnvironment: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyEnvironment, nil),
			},
			schemaCredentialsId: {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyCredentialsId, nil),
			},
			//TODO: make optional when a fixed base url exists
			schemaBaseUrl: {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyBaseUrl, nil),
			},
			schemaTokenUrl: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyTokenUrl, nil),
			},
			schemaClientId: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyClientId, nil),
			},
			schemaClientSecret: {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyClientSecret, nil),
			},
			schemaClientScope: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyClientScope, nil),
			},
			schemaApiKey: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyApiKey, nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"pinto_dns_zone":   resourceDnsZone(),
			"pinto_dns_record": resourceDnsRecord(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"pinto_dns_zone":    dataSourceDnsZone(),
			"pinto_dns_zones":   dataSourceDnsZones(),
			"pinto_dns_record":  dataSourceDnsRecord(),
			"pinto_dns_records": dataSourceDnsRecords(),
		},
		ConfigureContextFunc: func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
			// override the provider client e.g. with a mock client used during tests and disable diagnostics
			if client != nil {
				return &PintoProvider{
					IPintoProvider: nil,
					client:         client,
					apiKey:         "",
					provider:       "",
					environment:    "",
				}, nil
			}
			return providerConfigure(ctx, data)
		},
	}
}

func configureOAuthClient(d *schema.ResourceData) (cc.Config, error) {
	tokenUrl := d.Get(schemaTokenUrl).(string)

	clientId, ok := d.GetOk(schemaClientId)
	if !ok {
		return cc.Config{}, fmt.Errorf("using client-credentials requires %s and %s to be set too for pinto", envKeyClientId, envKeyClientSecret)
	}
	clientSecret, ok := d.GetOk(schemaClientSecret)
	if !ok {
		return cc.Config{}, fmt.Errorf("using client-credentials requires %s and %s to be set too for pinto", envKeyClientId, envKeyClientSecret)
	}

	var oauthConfig cc.Config
	clientScope, ok := d.GetOk(schemaClientScope)
	if ok {
		var scopes []string
		s := strings.Split(clientScope.(string), ",")
		for _, scope := range s {
			scopes = append(scopes, scope)
		}

		oauthConfig = cc.Config{
			TokenURL:     tokenUrl,
			ClientID:     clientId.(string),
			ClientSecret: clientSecret.(string),
			Scopes:       scopes,
		}
	} else {
		oauthConfig = cc.Config{
			ClientID:     clientId.(string),
			ClientSecret: clientSecret.(string),
			TokenURL:     tokenUrl,
		}
	}

	return oauthConfig, nil
}
func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	provider := PintoProvider{
		apiKey: "",
	}
	_, ok := d.GetOk(schemaProvider)
	if ok {
		provider.provider = d.Get(schemaProvider).(string)
	} else {
		provider.provider = ""
	}
	_, ok = d.GetOk(schemaEnvironment)
	if ok {
		provider.environment = d.Get(schemaEnvironment).(string)
	} else {
		provider.environment = ""
	}
	_, ok = d.GetOk(schemaCredentialsId)
	if ok {
		provider.credentialsId = d.Get(schemaCredentialsId).(string)
	} else {
		provider.credentialsId = ""
	}

	var xApiOptions, err = json.Marshal(XApiOptions{
		AccessOptions: AccessOptions{
			Provider:      provider.provider,
			Environment:   provider.environment,
			CredentialsId: provider.credentialsId,
		},
	})

	if err != nil {
		diags = append(
			diags,
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "CredentialsId not set",
				Detail:   "Unable to setup xApiOptions",
			})
	}

	provider.xApiOptions = string(xApiOptions)

	clientConf := gopinto.NewConfiguration()
	clientConf.Servers[0].URL = d.Get(schemaBaseUrl).(string)

	val, ok := d.GetOk(schemaApiKey)
	if ok {
		provider.apiKey = val.(string)
	}
	_, ok = d.GetOk(schemaClientId)
	if ok {
		oAuthConf, err := configureOAuthClient(d)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Pinto client",
				Detail:   "Require " + schemaClientId + " and " + schemaClientSecret + " to be set together for client-credentials authentication",
			})
			return nil, diags
		}
		// TODO: Do we need to move this into a utils class and use a different context each time?
		clientConf.HTTPClient = oAuthConf.Client(context.Background())

	}
	client := gopinto.NewAPIClient(clientConf)
	provider.client = client

	log.Printf("[DEBUG] Pinto: %s, %s, %s \n", d.Get(schemaBaseUrl).(string), provider.provider, provider.environment)
	return &provider, diags
}
