package chronos

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"net/url"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider {
		Schema: map[string]*schema.Schema{
			"url": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
				DefaultFunc: schema.EnvDefaultFunc("CHRONOS_URL", nil),
				Description: "Chronos base URL",
			},
			"basic_auth_user": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Default: "",
				Description: "HTTP basic auth user",
			},
			"basic_auth_password": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Default: "",
				Description: "HTTP basic auth password",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"chronos_job": resourceChronosJob(),
		},
		ConfigureFunc: generateChronosConfig,
	}
}

func generateChronosConfig(rd *schema.ResourceData) (interface{}, error) {
	chronosUrl, err := url.Parse(rd.Get("url").(string))
	if err != nil {
		return nil, err
	}

	config := Config{
		Url: chronosUrl,
		UserInfo: url.UserPassword(rd.Get("basic_auth_user").(string), rd.Get("basic_auth_password").(string)),
	}
	return config, nil
}
