package server

import (
	"context"
	"testing"
	"time"
)

func TestDomainStoreCreateListAndPrimaryAssignment(t *testing.T) {
	db := mustOpenTestDB(t)
	siteStore := NewSiteStore(db)
	domainStore := NewDomainStore(db)
	serverID := mustInsertServerWithStatus(t, db, "ready")

	siteID, err := siteStore.Create(context.Background(), CreateSiteInput{ServerID: serverID, Name: "Northwind", Status: SiteStatusDraft})
	if err != nil {
		t.Fatalf("create site: %v", err)
	}
	baseID, err := domainStore.Create(context.Background(), CreateDomainInput{
		Hostname:  "sandbox.pressluft.test",
		Kind:      DomainKindBase,
		Ownership: DomainOwnershipPlatform,
		Source:    DomainSourceSandbox,
		Status:    DomainStatusActive,
	})
	if err != nil {
		t.Fatalf("create base domain: %v", err)
	}
	firstID, err := domainStore.Create(context.Background(), CreateDomainInput{
		Hostname:       "northwind.sandbox.pressluft.test",
		Kind:           DomainKindHostname,
		Ownership:      DomainOwnershipPlatform,
		Source:         DomainSourceSandbox,
		Status:         DomainStatusActive,
		SiteID:         siteID,
		ParentDomainID: baseID,
		IsPrimary:      true,
	})
	if err != nil {
		t.Fatalf("create first hostname: %v", err)
	}
	secondID, err := domainStore.Create(context.Background(), CreateDomainInput{
		Hostname:  "www.northwind.example.com",
		Kind:      DomainKindHostname,
		Ownership: DomainOwnershipCustomer,
		Source:    DomainSourceCustom,
		Status:    DomainStatusActive,
		SiteID:    siteID,
		IsPrimary: true,
	})
	if err != nil {
		t.Fatalf("create second hostname: %v", err)
	}
	first, err := domainStore.GetByID(context.Background(), firstID)
	if err != nil {
		t.Fatalf("get first hostname: %v", err)
	}
	second, err := domainStore.GetByID(context.Background(), secondID)
	if err != nil {
		t.Fatalf("get second hostname: %v", err)
	}
	if first.IsPrimary {
		t.Fatal("expected first hostname to no longer be primary")
	}
	if !second.IsPrimary {
		t.Fatal("expected second hostname to be primary")
	}
	storedSite, err := siteStore.GetByID(context.Background(), siteID)
	if err != nil {
		t.Fatalf("get site: %v", err)
	}
	if storedSite.PrimaryDomain != "www.northwind.example.com" {
		t.Fatalf("primary_domain = %q, want %q", storedSite.PrimaryDomain, "www.northwind.example.com")
	}
	domains, err := domainStore.ListBySite(context.Background(), siteID)
	if err != nil {
		t.Fatalf("list domains by site: %v", err)
	}
	if len(domains) != 2 {
		t.Fatalf("site domain count = %d, want 2", len(domains))
	}
}

func TestDomainStoreBackfillsLegacyPrimaryDomains(t *testing.T) {
	db := mustOpenTestDB(t)
	domainStore := NewDomainStore(db)
	serverID := mustInsertServerWithStatus(t, db, "ready")
	siteID := nextTestPublicID(t, db, "sites")
	if _, err := db.Exec(
		`INSERT INTO sites (id, server_id, name, primary_domain, status, created_at, updated_at) VALUES (?, ?, ?, ?, 'active', ?, ?)`,
		siteID,
		serverID,
		"Legacy Site",
		"legacy.example.test",
		time.Now().UTC().Format(time.RFC3339),
		time.Now().UTC().Format(time.RFC3339),
	); err != nil {
		t.Fatalf("insert legacy site: %v", err)
	}
	if err := domainStore.BackfillLegacyPrimaryDomains(context.Background()); err != nil {
		t.Fatalf("backfill legacy domains: %v", err)
	}
	domains, err := domainStore.ListBySite(context.Background(), siteID)
	if err != nil {
		t.Fatalf("list domains by site: %v", err)
	}
	if len(domains) != 1 {
		t.Fatalf("backfilled domain count = %d, want 1", len(domains))
	}
	if domains[0].Source != DomainSourceLegacy {
		t.Fatalf("source = %q, want %q", domains[0].Source, DomainSourceLegacy)
	}
	if !domains[0].IsPrimary {
		t.Fatal("expected backfilled domain to be primary")
	}
}

func TestDomainStoreEnsurePlatformBaseDomains(t *testing.T) {
	db := mustOpenTestDB(t)
	domainStore := NewDomainStore(db)

	if err := domainStore.EnsurePlatformBaseDomains(context.Background()); err != nil {
		t.Fatalf("ensure platform base domains: %v", err)
	}
	if err := domainStore.EnsurePlatformBaseDomains(context.Background()); err != nil {
		t.Fatalf("ensure platform base domains second pass: %v", err)
	}

	domains, err := domainStore.List(context.Background())
	if err != nil {
		t.Fatalf("list domains: %v", err)
	}
	if len(domains) != 2 {
		t.Fatalf("domain count = %d, want 2", len(domains))
	}

	byHostname := map[string]StoredDomain{}
	for _, domain := range domains {
		byHostname[domain.Hostname] = domain
	}

	if byHostname["pressluft.bombig.app"].Status != DomainStatusActive {
		t.Fatalf("pressluft.bombig.app status = %q, want %q", byHostname["pressluft.bombig.app"].Status, DomainStatusActive)
	}
	if byHostname["pressluft.dev"].Status != DomainStatusPending {
		t.Fatalf("pressluft.dev status = %q, want %q", byHostname["pressluft.dev"].Status, DomainStatusPending)
	}
}
