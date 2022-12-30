package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceIPAllowListEntry(t *testing.T) {
	t.Skip("Acceptance tests are supposed to reach out to a real API. Tests is skipped until we create a test GitHub organisation.")

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIPAllowListEntry,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("githubipallowlist_ip_allow_list_entry.example", "is_active", "false"),
					resource.TestCheckResourceAttr("githubipallowlist_ip_allow_list_entry.example", "allow_list_value", "1.2.3.4/32"),
				),
			},
		},
	})
}

const testAccResourceIPAllowListEntry = `
resource "githubipallowlist_ip_allow_list_entry" "example" {
  is_active        = false
  allow_list_value = "1.2.3.4/32"
}
`
