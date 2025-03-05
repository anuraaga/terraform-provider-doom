package provider

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ProcessProvider satisfies various provider interfaces.
var (
	_ provider.Provider = &DoomProvider{}
)

// DoomProvider implements terraform-provider-process.
type DoomProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// DoomProviderModel describes the configuration for running Doom.
type DoomProviderModel struct {
	Path types.String `tfsdk:"path"`
}

func (p *DoomProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "doom"
	resp.Version = p.version
}

func (p *DoomProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				MarkdownDescription: "Path to Doom executable",
				Required:            true,
			},
		},
	}
}

func (p *DoomProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data DoomProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	path, err := exec.LookPath(data.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to find Doom executable", fmt.Sprintf("... details ... %s", err))
		return
	}

	resp.ResourceData = path
}

func (p *DoomProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSessionResource,
	}
}

func (p *DoomProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DoomProvider{
			version: version,
		}
	}
}
