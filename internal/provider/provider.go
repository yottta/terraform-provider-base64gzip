package provider

import (
	"context"
	"terraform-provider-base64gzip/internal/provider/functions"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ provider.Provider              = &base64gzip{}
	_ provider.ProviderWithFunctions = &base64gzip{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &base64gzip{
			version: version,
		}
	}
}

type base64gzip struct {
	version string
}

func (p *base64gzip) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "base64gzip"
	resp.Version = p.version
}

func (p *base64gzip) Schema(_ context.Context, _ provider.SchemaRequest, _ *provider.SchemaResponse) {
}

func (p *base64gzip) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

func (p *base64gzip) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *base64gzip) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *base64gzip) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *base64gzip) Functions(context.Context) []func() function.Function {
	return []func() function.Function{
		func() function.Function {
			return functions.Base64GzipFunction{}
		},
		func() function.Function {
			return functions.Base64GunzipFunction{}
		},
	}
}
