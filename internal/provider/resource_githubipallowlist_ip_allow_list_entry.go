package provider

import (
	"context"
	"github.com/form3tech-oss/terraform-provider-githubipallowlist/github"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	entryDescription  = "Managed by Terraform"
	isActiveKey       = "is_active"
	allowListValueKey = "allow_list_value"
)

func resourceGitHubIPAllowListEntry() *schema.Resource {
	return &schema.Resource{
		Description: "GitHub IP allow list entry.",

		CreateContext: resourceGitHubIPAllowListEntryCreate,
		ReadContext:   resourceGitHubIPAllowListEntryRead,
		UpdateContext: resourceGitHubIPAllowListEntryUpdate,
		DeleteContext: resourceGitHubIPAllowListEntryDelete,

		Schema: map[string]*schema.Schema{
			isActiveKey: {
				Description: "Whether the entry is currently active.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			allowListValueKey: {
				Description: "A single IP address or range of IP addresses in CIDR notation.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceGitHubIPAllowListEntryCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*apiClient)

	isActive := d.Get(isActiveKey).(bool)
	value := d.Get(allowListValueKey).(string)

	entry, err := client.github.CreateIPAllowListEntry(ctx, client.ownerID, entryDescription, github.CIDR(value), isActive)
	if err != nil {
		diag.FromErr(err)
	}

	d.SetId(entry.ID)

	tflog.Trace(ctx, "created a resource githubipallowlist_ip_allow_list_entry")

	return nil
}

func resourceGitHubIPAllowListEntryRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*apiClient)

	entries, err := client.github.GetOrganizationIPAllowListEntries(ctx, client.organization)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	entry := firstEntryByID(entries, id)
	if entry == nil {
		tflog.Warn(ctx, "githubipallowlist_ip_allow_list_entry not found", map[string]interface{}{"id": id})
		d.SetId("")
		return nil
	}
	err = d.Set(isActiveKey, entry.IsActive)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set(allowListValueKey, entry.AllowListValue)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func firstEntryByID(entries []*github.IPAllowListEntry, id string) *github.IPAllowListEntry {
	for _, e := range entries {
		if e != nil && e.ID == id {
			return e
		}
	}
	return nil
}

func resourceGitHubIPAllowListEntryUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*apiClient)

	id := d.Id()
	isActive := d.Get(isActiveKey).(bool)
	value := d.Get(allowListValueKey).(string)

	entry, err := client.github.UpdateIPAllowListEntry(ctx, id,
		github.IPAllowListEntryParameters{
			Name:     entryDescription,
			Value:    github.CIDR(value),
			IsActive: isActive,
		})
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set(isActiveKey, entry.IsActive)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set(allowListValueKey, entry.AllowListValue)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(entry.ID)

	tflog.Trace(ctx, "updated a resource githubipallowlist_ip_allow_list_entry", map[string]interface{}{"id": entry.ID})

	return nil
}

func resourceGitHubIPAllowListEntryDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}
