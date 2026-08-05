package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yinebebt/deeplink"
)

var testEndpoint string

func TestMain(m *testing.M) {
	store := deeplink.NewMemoryStore()
	svc, err := deeplink.New(deeplink.Config{
		BaseURL: "http://deeplink.test/",
		Store:   store,
	})
	if err != nil {
		panic(err)
	}
	svc.Register(deeplink.RedirectProcessor{})

	srv := httptest.NewServer(svc.Handler())
	defer srv.Close()
	defer svc.Close() //nolint:errcheck

	testEndpoint = srv.URL
	os.Exit(m.Run())
}

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"deeplink": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func TestAccLinkResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLinkConfig(testEndpoint, "https://example.com/docs", "Docs"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("deeplink_link.docs", "type", "redirect"),
					resource.TestCheckResourceAttr("deeplink_link.docs", "url", "https://example.com/docs"),
					resource.TestCheckResourceAttr("deeplink_link.docs", "title", "Docs"),
					resource.TestCheckResourceAttrSet("deeplink_link.docs", "short_id"),
					resource.TestCheckResourceAttrSet("deeplink_link.docs", "short_link"),
					resource.TestCheckResourceAttrSet("deeplink_link.docs", "id"),
				),
			},
			{
				Config: testAccLinkConfig(testEndpoint, "https://example.com/guide", "Guide"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("deeplink_link.docs", "url", "https://example.com/guide"),
					resource.TestCheckResourceAttr("deeplink_link.docs", "title", "Guide"),
				),
			},
		},
	})
}

func TestShortIDFromURL(t *testing.T) {
	got, err := shortIDFromURL("http://localhost:8090/AbCd123")
	if err != nil {
		t.Fatal(err)
	}
	if got != "AbCd123" {
		t.Fatalf("got %q", got)
	}
}

func testAccLinkConfig(endpoint, dest, title string) string {
	return fmt.Sprintf(`
provider "deeplink" {
  endpoint = %q
}

resource "deeplink_link" "docs" {
  type  = "redirect"
  url   = %q
  title = %q
}
`, endpoint, dest, title)
}

func testAccCheckLinkDestroy(s *terraform.State) error {
	client, err := newClient(testEndpoint, "")
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "deeplink_link" {
			continue
		}
		_, err := client.GetLink(context.Background(), rs.Primary.Attributes["type"], rs.Primary.Attributes["short_id"])
		if err == nil {
			return fmt.Errorf("link %s still exists", rs.Primary.ID)
		}
		var apiErr *apiError
		if !errors.As(err, &apiErr) || apiErr.Status != http.StatusNotFound {
			return err
		}
	}
	return nil
}
