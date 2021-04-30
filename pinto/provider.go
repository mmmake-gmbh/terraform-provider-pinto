package pinto

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/whizus/gopinto"
	cc "golang.org/x/oauth2/clientcredentials"
)

type IPintoProvider interface{}
type PintoProvider struct {
	IPintoProvider

	client      *gopinto.APIClient
	apiKey      string
	provider    string
	environment string
}

const (
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

func NewProvider(
	client *gopinto.APIClient,
	apiKey string,
	provider string,
	environment string,
) *PintoProvider {
	return &PintoProvider{
		IPintoProvider: nil,
		client:         client,
		apiKey:         apiKey,
		provider:       provider,
		environment:    environment,
	}
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
				return NewProvider(
					client,
					"",
					"",
					"",
					), nil
			}
			var diags diag.Diagnostics

			clientConf := gopinto.NewConfiguration()
			// TODO: Override this with an DEBUG env if needed
			clientConf.Debug = false
			client := gopinto.NewAPIClient(clientConf)

			// clientConf.UserAgent = fmt.Sprintf("HashiCorp Terraform/%s (+https://www.terraform.io) Terraform Plugin SDK/%s", provider.TerraformVersion, meta.SDKVersionString())
			clientConf.Servers[0].URL = data.Get(schemaBaseUrl).(string)

			// global settings
			schemaProvider := data.Get(schemaProvider).(string)
			schemaEnvironment := data.Get(schemaEnvironment).(string)

			// oauth configuration
			schemaClientId := data.Get(schemaClientId).(string)
			schemaClientSecret := data.Get(schemaClientSecret).(string)

			// oauth validation
			if schemaClientId != "" && schemaClientSecret != "" {
				oAuthConf, err := configureOAuthClient(data)
				clientConf.HTTPClient = oAuthConf.Client(context.Background())

				client := gopinto.NewAPIClient(clientConf)

				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Unable to create Pinto client",
						Detail:   err.Error(),
					})
					return nil, diags
				}

				apiKey := data.Get(schemaApiKey).(string)
				if apiKey != "" {
					context.WithValue(ctx, gopinto.ContextAPIKeys, apiKey)
				} else {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Warning,
						Summary:  "Api key missing",
						Detail:   "No api key was given. Please set one to use the ",
					})
				}

				return NewProvider(
					client,
					apiKey,
					schemaProvider,
					schemaEnvironment,
				), diags
			}

			// returns the default provider without oAuth configured
			return NewProvider(
				client,
				"",
				schemaProvider,
				schemaEnvironment,
			), diags
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
			TokenURL:     tokenUrl,
			ClientID:     clientId.(string),
			ClientSecret: clientSecret.(string),
		}
	}

	return oauthConfig, nil
}