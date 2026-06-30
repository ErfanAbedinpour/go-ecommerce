package storefront

import (
	"context"

	domainsettings "app/internal/domain/settings"
)

// AboutPage is the public about page aggregate.
type AboutPage struct {
	Hero       domainsettings.AboutHero         `json:"hero"`
	Story      domainsettings.AboutStory        `json:"story"`
	Mission    domainsettings.AboutTextBlock    `json:"mission"`
	Vision     domainsettings.AboutTextBlock    `json:"vision"`
	Milestones []domainsettings.AboutMilestone  `json:"milestones"`
	Team       []domainsettings.AboutTeamMember `json:"team"`
	Stats      domainsettings.AboutStats        `json:"stats"`
	Contact    domainsettings.PublicContact     `json:"contact"`
	Social     domainsettings.PublicSocial      `json:"social"`
	SEO        domainsettings.AboutSEO          `json:"seo"`
}

// GetAboutPage returns aggregated content for the public about page.
func (s *Service) GetAboutPage(ctx context.Context) (*AboutPage, error) {
	store, err := s.settings.Get(ctx)
	if err != nil {
		return nil, err
	}

	productsCount, err := s.products.CountActive(ctx)
	if err != nil {
		return nil, err
	}

	about := store.About
	stats := about.Stats
	if stats.ProductsCount == 0 {
		stats.ProductsCount = productsCount
	}

	return &AboutPage{
		Hero:       about.Hero,
		Story:      about.Story,
		Mission:    about.Mission,
		Vision:     about.Vision,
		Milestones: about.Milestones,
		Team:       about.Team,
		Stats:      stats,
		Contact: domainsettings.PublicContact{
			Phone:        store.Contact.Phone,
			Mobile:       store.Contact.Phone,
			Email:        store.Contact.Email,
			Address:      store.Contact.Address,
			WorkingHours: "",
		},
		Social: domainsettings.PublicSocial{
			Instagram: store.Social.Instagram,
			WhatsApp:  store.Social.Instagram,
			Telegram:  store.Social.Twitter,
			Facebook:  store.Social.Facebook,
			Twitter:   store.Social.Twitter,
			LinkedIn:  store.Social.LinkedIn,
			YouTube:   store.Social.YouTube,
			TikTok:    store.Social.TikTok,
		},
		SEO: domainsettings.AboutSEO{
			MetaTitle:       store.SEO.MetaTitle,
			MetaDescription: store.SEO.MetaDescription,
		},
	}, nil
}
