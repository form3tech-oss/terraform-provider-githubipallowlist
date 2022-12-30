package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	entryDescription = "Managed by Terraform"
)

func resourceGitHubIPAllowListEntry() *schema.Resource {
	return &schema.Resource{
		Description: "GitHub IP allow list entry managed on an organization level.",

		CreateContext: resourceGitHubIPAllowListEntryCreate,
		ReadContext:   resourceGitHubIPAllowListEntryRead,
		UpdateContext: resourceGitHubIPAllowListEntryUpdate,
		DeleteContext: resourceGitHubIPAllowListEntryDelete,

		Schema: map[string]*schema.Schema{
			"is_active": {
				Description: "Whether the entry is currently active.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"allow_list_value": {
				Description: "A single IP address or range of IP addresses in CIDR notation.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceGitHubIPAllowListEntryCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*apiClient)

	isActive := d.Get("is_active").(bool)
	value := d.Get("allow_list_value").(string)

	entry, err := client.github.CreateIPAllowListEntry(ctx, client.ownerID, entryDescription, value, isActive)
	if err != nil {
		diag.FromErr(err)
	}

	d.SetId(entry.ID)

	tflog.Trace(ctx, "created a resource githubipallowlist_ip_allow_list_entry")

	return nil
}

func resourceGitHubIPAllowListEntryRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}

func resourceGitHubIPAllowListEntryUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}

func resourceGitHubIPAllowListEntryDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}
