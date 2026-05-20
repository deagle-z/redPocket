package repository

import (
	"BaseGoUni/core/pojo"
	"testing"
)

func TestShouldPersistBannerI18n_AllowsTextOnlyPopupAnnouncement(t *testing.T) {
	desc := "Platform maintenance notice"
	item := pojo.SysBannerI18n{
		LanguageCode: "all",
		Description:  &desc,
		ImageURL:     "",
	}
	banner := pojo.SysBanner{
		Position:    "popup",
		BannerType:  "popup",
		DisplayType: "popup",
	}

	if !shouldPersistBannerI18n(banner, item) {
		t.Fatal("expected text-only popup announcement i18n to be persisted")
	}
}

func TestShouldPersistBannerI18n_RequiresImageForNonPopupBanner(t *testing.T) {
	desc := "Homepage copy without image"
	item := pojo.SysBannerI18n{
		LanguageCode: "all",
		Description:  &desc,
		ImageURL:     "",
	}
	banner := pojo.SysBanner{
		Position:    "home",
		BannerType:  "image",
		DisplayType: "banner",
	}

	if shouldPersistBannerI18n(banner, item) {
		t.Fatal("expected non-popup banner without image to be skipped")
	}
}
