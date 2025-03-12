package provider

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/shirou/gopsutil/v4/process"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource = &SessionResource{}
)

func NewSessionResource() resource.Resource {
	return &SessionResource{}
}

// SessionResource defines the resource implementation for a session of Doom.
type SessionResource struct {
	path string
}

// SessionResourceModel describes the resource data model for a session of Doom.
type SessionResourceModel struct {
	Wad types.String `tfsdk:"wad"`
	Pid types.Int32  `tfsdk:"pid"`
}

func (r *SessionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_session"
}

func (r *SessionResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Doom Play Session",

		Attributes: map[string]schema.Attribute{
			"wad": schema.StringAttribute{
				MarkdownDescription: "WAD file to load, e.g. freedoom1.wad",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pid": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "Process ID of the Doom session",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *SessionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	path, ok := req.ProviderData.(string)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected string, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.path = path
}

func (r *SessionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SessionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	cmd := exec.CommandContext(ctx, r.path, "-iwad", data.Wad.ValueString()) // #nosec G204
	if err := cmd.Start(); err != nil {
		resp.Diagnostics.AddError("Failed to start Doom", fmt.Sprintf("... details ... %s", err))
		return
	}

	pid := cmd.Process.Pid

	if err := cmd.Process.Release(); err != nil {
		resp.Diagnostics.AddError("Failed to start Doom", fmt.Sprintf("... details ... %s", err))
		return
	}

	data.Pid = types.Int32Value(int32(pid)) // #nosec G115

	tflog.Trace(ctx, fmt.Sprintf("started %s with pid %d", r.path, pid))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SessionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SessionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if ok, _ := process.PidExistsWithContext(ctx, data.Pid.ValueInt32()); !ok {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SessionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Please report this issue to the provider developers.")
}

func (r *SessionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SessionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	p, err := process.NewProcessWithContext(ctx, data.Pid.ValueInt32())
	if err != nil {
		// Process was killed externally during CLI execution, we consider this a success.
		return
	}
	if err := p.KillWithContext(ctx); err != nil {
		resp.Diagnostics.AddError("Failed to kill Doom session", fmt.Sprintf("... details ... %s", err))
		return
	}
}
