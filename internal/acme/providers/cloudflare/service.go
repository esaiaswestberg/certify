package cloudflare

import (
	"certify/internal/acme/providers/provider_utils"
	"certify/internal/acme/zone_configuration"
	"certify/internal/configuration"
	"fmt"
	legoCertificate "github.com/go-acme/lego/v4/certificate"
	cloudflareChallenge "github.com/go-acme/lego/v4/providers/dns/cloudflare"
)

type Provider struct{}

func NewProvider() Provider {
	return Provider{}
}

func (p Provider) ObtainCertificate(configuration *configuration.Configuration, zoneConfiguration *zone_configuration.ZoneConfiguration) (*legoCertificate.Resource, error) {
	acmeUser, exists, err := provider_utils.GetACMEUser(configuration, zoneConfiguration.IdentityEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	acmeClient, err := provider_utils.GetACMEClient(acmeUser, configuration, zoneConfiguration)
	if err != nil {
		return nil, fmt.Errorf("failed to get ACME client: %w", err)
	}

	defer provider_utils.UnsetEnvironmentVariable("CLOUDFLARE_DNS_API_TOKEN")
	if err = provider_utils.SetEnvironmentVariable("CLOUDFLARE_DNS_API_TOKEN", zoneConfiguration, "api_token"); err != nil {
		return nil, fmt.Errorf("failed to set Cloudflare API token: %w", err)
	}

	dnsProvider, err := cloudflareChallenge.NewDNSProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create DNS provider: %w", err)
	}

	if err := acmeClient.Challenge.SetDNS01Provider(dnsProvider); err != nil {
		return nil, fmt.Errorf("failed to set DNS provider: %w", err)
	}

	// If the user does not exist, register them
	if !exists {
		if err := provider_utils.RegisterACMEUser(acmeClient, acmeUser); err != nil {
			return nil, fmt.Errorf("failed to register user: %w", err)
		}
	}

	return provider_utils.ObtainACMECertificate(acmeClient, zoneConfiguration)
}
