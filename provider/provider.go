package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/whizus/go-stackit"
	cc "golang.org/x/oauth2/clientcredentials"
)

type IPintoProvider interface{}
type PintoProvider struct {
	IPintoProvider

	client      *stackit.APIClient
	apiKey      string
	provider    string
	environment string
}

const (
	schemaProvider     = "provider"
	schemaEnvironment  = "environment"
	schemaBaseUrl      = "base_url"
	schemaTokenUrl     = "token_url"
	schemaClientId     = "client_id"
	schemaClientSecret = "client_secret"
	schemaClientScope  = "client_scope"
	schemaApiKey       = "api_key"

	envKeyBaseUrl      = "PINTO_BASE_URL"
	envKeyTokenUrl     = "PINTO_TOKEN_URL"
	envKeyProvider     = "PINTO_PROVIDER"
	envKeyEnvironment  = "PINTO_ENVIRONMENT"
	envKeyApiKey       = "PINTO_API_KEY"
	envKeyClientId     = "PINTO_CLIENT_ID"
	envKeyClientSecret = "PINTO_CLIENT_SECRET"
	envKeyClientScope  = "PINTO_CLIENT_SCOPE"
)

// Provider -
func Provider() *schema.Provider {
	log.Printf("[DEBUG] Pinto: Starting Provider")
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			//TODO: Should we make this optional and allow user to set this on resource level?
			schemaProvider: {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyProvider, nil),
			},
			//TODO: Should we make this optional and allow user to set this on resource level?
			schemaEnvironment: {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyEnvironment, nil),
			},
			schemaBaseUrl: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyBaseUrl, "https://pinto.irgendwo.co"),
			},
			schemaTokenUrl: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyTokenUrl, "https://auth.pinto.irgendwo.co/connect/token"),
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
		ResourcesMap: map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"pinto_dns_zone": dataSourceDnsZone(),
		},
		ConfigureContextFunc: providerConfigure,
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
			TokenURL:     tokenUrl,
			ClientID:     clientId.(string),
			ClientSecret: clientSecret.(string),
		}
	}

	return oauthConfig, nil
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	provider := PintoProvider{
		apiKey: "",
	}
	provider.provider = d.Get(schemaProvider).(string)
	provider.environment = d.Get(schemaEnvironment).(string)

	clientConf := stackit.NewConfiguration()
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
		//TODO: Do we need to move this into a utils class and use a different context each time?
		clientConf.HTTPClient = oAuthConf.Client(context.Background())

	}
	client := stackit.NewAPIClient(clientConf)
	provider.client = client

	log.Printf("[DEBUG] Pinto: %s, %s, %s \n", d.Get(schemaBaseUrl).(string), provider.provider, provider.environment)
	return &provider, diags
}
