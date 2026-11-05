package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
)

type PublicInvitationService struct {
	invitationRepo repositories.InvitationRepository
	templateRepo   repositories.TemplateRepository
	guestRepo      repositories.GuestRepository
	wishRepo       repositories.WishRepository
}

func NewPublicInvitationService(
	invitationRepo repositories.InvitationRepository,
	templateRepo repositories.TemplateRepository,
	guestRepo repositories.GuestRepository,
	wishRepo repositories.WishRepository,
) *PublicInvitationService {
	return &PublicInvitationService{
		invitationRepo: invitationRepo,
		templateRepo:   templateRepo,
		guestRepo:      guestRepo,
		wishRepo:       wishRepo,
	}
}

// GetInvitationBySlug resolves public invitation data formatted for Frontend Nuxt 3 templates
func (s *PublicInvitationService) GetInvitationBySlug(slug string) (map[string]interface{}, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("slug is required")
	}

	invitation, err := s.invitationRepo.FindByIDOrSlug(slug)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, errors.New("invitation not found")
	}

	// 1. Resolve Template
	template := s.resolveTemplate(invitation)

	// 2. Resolve Features Config
	featuresConfig := s.resolveFeaturesConfig(invitation, template)

	// 3. Resolve Inner Invitation Data
	invitationPayload := s.buildInvitationPayload(invitation, featuresConfig)

	// 4. Construct Public Contract
	slugVal := ""
	if invitation.Slug != nil {
		slugVal = *invitation.Slug
	}

	var templateID *uuid.UUID
	var templateMap models.PublicInvitationTemplate
	if template != nil {
		templateID = &template.ID
		templateMap = models.PublicInvitationTemplate{
			Name:          template.Name,
			NuxtComponent: template.NuxtComponent,
			Slug:          template.Slug,
		}
	} else {
		templateMap = models.PublicInvitationTemplate{
			Name:          "Scarlet Sketch",
			NuxtComponent: "TemplateScarletSketch",
			Slug:          "scarlet-sketch",
		}
	}

	result := map[string]interface{}{
		"id":              invitation.ID,
		"slug":            slugVal,
		"template_id":     templateID,
		"template":        templateMap,
		"features_config": featuresConfig,
		"invitation":      invitationPayload,
	}

	return result, nil
}

func (s *PublicInvitationService) resolveTemplate(invitation *models.Invitation) *models.Template {
	if invitation.Template != nil {
		return invitation.Template
	}

	if invitation.TemplateID != nil {
		if tmpl, err := s.templateRepo.FindByID(*invitation.TemplateID); err == nil && tmpl != nil {
			return tmpl
		}
	}

	// Look inside Theme jsonb
	if invitation.Theme != nil {
		if themeMap, ok := toMap(invitation.Theme); ok {
			if tid, exists := themeMap["template_id"]; exists && tid != nil {
				tidStr := fmt.Sprintf("%v", tid)
				if u, err := uuid.Parse(tidStr); err == nil {
					if tmpl, err := s.templateRepo.FindByID(u); err == nil && tmpl != nil {
						return tmpl
					}
				}
				if tmpl, err := s.templateRepo.FindBySlug(tidStr); err == nil && tmpl != nil {
					return tmpl
				}
			}

			if comp, exists := themeMap["template_component"]; exists && comp != nil {
				compStr := fmt.Sprintf("%v", comp)
				if tmpl, err := s.templateRepo.FindBySlug(compStr); err == nil && tmpl != nil {
					return tmpl
				}
			}
		}
	}

	// Fallback to "scarlet-sketch" or "noir" or first active template
	if tmpl, err := s.templateRepo.FindBySlug("scarlet-sketch"); err == nil && tmpl != nil {
		return tmpl
	}
	if tmpl, err := s.templateRepo.FindBySlug("noir"); err == nil && tmpl != nil {
		return tmpl
	}

	templates, _, _ := s.templateRepo.FindAll(1, 1, "", "", nil, false)
	if len(templates) > 0 {
		return &templates[0]
	}

	return nil
}

func (s *PublicInvitationService) resolveFeaturesConfig(invitation *models.Invitation, template *models.Template) map[string]interface{} {
	// Standard defaults
	result := map[string]interface{}{
		"has_countdown": true,
		"has_gallery":   true,
		"has_story":     false,
		"has_gift":      true,
		"has_rsvp":      true,
		"has_video":     false,
		"has_maps":      true,
		"has_qr":        false,
		"has_audio":     true,
	}

	// Merge template defaults if present
	if template != nil && template.DefaultFeaturesConfig != nil {
		for k, v := range template.DefaultFeaturesConfig {
			result[k] = v
		}
	}

	// Merge package features config if present (overrides template defaults)
	if invitation.Package != nil && invitation.Package.FeaturesConfig != nil {
		for k, v := range invitation.Package.FeaturesConfig {
			result[k] = v
		}
	} else if invitation.Order != nil && invitation.Order.Package != nil && invitation.Order.Package.FeaturesConfig != nil {
		for k, v := range invitation.Order.Package.FeaturesConfig {
			result[k] = v
		}
	}

	return result
}

func (s *PublicInvitationService) buildInvitationPayload(invitation *models.Invitation, featuresConfig map[string]interface{}) map[string]interface{} {
	payload := make(map[string]interface{})

	// 1. Audio URL
	if invitation.AudioURL != nil && *invitation.AudioURL != "" {
		payload["audio_url"] = *invitation.AudioURL
	} else if themeMap, ok := toMap(invitation.Theme); ok && themeMap["audio_url"] != nil {
		payload["audio_url"] = themeMap["audio_url"]
	} else {
		payload["audio_url"] = "https://storage.domain.com/audio/God-Bless-The-Broken-Road.mp3"
	}

	// 2. Groom
	groomMap := make(map[string]interface{})
	if g, ok := toMap(invitation.Groom); ok {
		for k, v := range g {
			groomMap[k] = v
		}
	} else if g, ok := toMap(invitation.GroomData); ok {
		for k, v := range g {
			groomMap[k] = v
		}
	}
	if _, ok := groomMap["name"]; !ok {
		if fn, exists := groomMap["full_name"]; exists {
			groomMap["name"] = fn
		} else {
			groomMap["name"] = "Mulyo Aryawijaya"
		}
	}
	if _, ok := groomMap["nickname"]; !ok {
		groomMap["nickname"] = "Mulyo"
	}
	if _, ok := groomMap["father"]; !ok {
		groomMap["father"] = "Bapak Arya"
	}
	if _, ok := groomMap["mother"]; !ok {
		groomMap["mother"] = "Ibu Wijaya"
	}
	if _, ok := groomMap["instagram"]; !ok {
		groomMap["instagram"] = "mulyoarya"
	}
	if _, ok := groomMap["photo_url"]; !ok {
		groomMap["photo_url"] = "https://storage.domain.com/photos/groom.jpg"
	}
	payload["groom"] = groomMap

	// 3. Bride
	brideMap := make(map[string]interface{})
	if b, ok := toMap(invitation.Bride); ok {
		for k, v := range b {
			brideMap[k] = v
		}
	} else if b, ok := toMap(invitation.BrideData); ok {
		for k, v := range b {
			brideMap[k] = v
		}
	}
	if _, ok := brideMap["name"]; !ok {
		if fn, exists := brideMap["full_name"]; exists {
			brideMap["name"] = fn
		} else {
			brideMap["name"] = "Putu Cempaka Giriyanti"
		}
	}
	if _, ok := brideMap["nickname"]; !ok {
		brideMap["nickname"] = "Cempaka"
	}
	if _, ok := brideMap["father"]; !ok {
		brideMap["father"] = "Bapak Giri"
	}
	if _, ok := brideMap["mother"]; !ok {
		brideMap["mother"] = "Ibu Yanti"
	}
	if _, ok := brideMap["instagram"]; !ok {
		brideMap["instagram"] = "cempakagiri"
	}
	if _, ok := brideMap["photo_url"]; !ok {
		brideMap["photo_url"] = "https://storage.domain.com/photos/bride.jpg"
	}
	payload["bride"] = brideMap

	// 4. Verse
	if invitation.VerseText != nil && *invitation.VerseText != "" {
		payload["verse_text"] = *invitation.VerseText
	} else {
		payload["verse_text"] = "Dan di atas semuanya itu: kenakanlah kasih, sebagai pengikat yang mempersatukan dan menyempurnakan."
	}
	if invitation.VerseRef != nil && *invitation.VerseRef != "" {
		payload["verse_ref"] = *invitation.VerseRef
	} else {
		payload["verse_ref"] = "KOLOSE 3:14"
	}

	// 5. Events
	eventsMap := make(map[string]interface{})
	if invitation.Events != nil {
		if ev, ok := toMap(invitation.Events); ok {
			eventsMap = ev
		}
	}
	if len(eventsMap) == 0 {
		if ev, ok := toMap(invitation.Event); ok {
			// Extract holy_matrimony / akad
			akadDate := getString(ev, "akad_date", "2026-05-18")
			akadTime := getString(ev, "akad_time", "09:00 - 11:00 WITA")
			venueName := getString(ev, "akad_venue_name", getString(ev, "venue_name", "Gereja Katedral Santo Roh Kudus"))
			address := getString(ev, "akad_address", getString(ev, "address", "Jl. Tukad Musi No. 1, Denpasar, Bali"))
			mapsURL := getString(ev, "akad_maps_url", getString(ev, "maps_url", "https://maps.google.com/?q=-8.670458,115.227484"))

			receptionDate := getString(ev, "reception_date", "2026-05-18")
			receptionTime := getString(ev, "reception_time", "18:00 WITA - Selesai")
			receptionVenue := getString(ev, "reception_venue_name", getString(ev, "venue_name", "The Glass House Bali"))
			receptionAddress := getString(ev, "reception_address", getString(ev, "address", "Jl. Danau Tamblingan No. 89, Sanur, Bali"))
			receptionMaps := getString(ev, "reception_maps_url", getString(ev, "maps_url", "https://maps.google.com/?q=-8.670458,115.227484"))

			eventsMap["holy_matrimony"] = map[string]interface{}{
				"title":      "Holy Matrimony",
				"date":       akadDate,
				"time_range": akadTime,
				"location":   venueName,
				"address":    address,
				"maps_url":   mapsURL,
			}
			eventsMap["reception"] = map[string]interface{}{
				"title":      "Wedding Reception",
				"date":       receptionDate,
				"time_range": receptionTime,
				"location":   receptionVenue,
				"address":    receptionAddress,
				"maps_url":   receptionMaps,
			}
		} else {
			eventsMap["holy_matrimony"] = map[string]interface{}{
				"title":      "Holy Matrimony",
				"date":       "2026-05-18",
				"time_range": "09:00 - 11:00 WITA",
				"location":   "Gereja Katedral Santo Roh Kudus",
				"address":    "Jl. Tukad Musi No. 1, Denpasar, Bali",
				"maps_url":   "https://maps.google.com/?q=...",
			}
			eventsMap["reception"] = map[string]interface{}{
				"title":      "Wedding Reception",
				"date":       "2026-05-18",
				"time_range": "18:00 WITA - Selesai",
				"location":   "The Glass House Bali",
				"address":    "Jl. Danau Tamblingan No. 89, Sanur, Bali",
				"maps_url":   "https://maps.google.com/?q=...",
			}
		}
	}
	payload["events"] = eventsMap

	// 6. Countdown Target Date
	if invitation.CountdownTargetDate != nil && *invitation.CountdownTargetDate != "" {
		payload["countdown_target_date"] = *invitation.CountdownTargetDate
	} else {
		payload["countdown_target_date"] = "2026-05-18T09:00:00+08:00"
	}

	// 7. Gallery (Conditional Feature Handling)
	var galleryList []string
	if isFeatureFlagEnabled(featuresConfig, "has_gallery") {
		if len(invitation.GalleryURLs) > 0 {
			galleryList = invitation.GalleryURLs
		} else if invitation.Gallery != nil {
			if rawSlice, ok := invitation.Gallery.([]interface{}); ok {
				for _, item := range rawSlice {
					if s, ok := item.(string); ok {
						galleryList = append(galleryList, s)
					}
				}
			} else if strSlice, ok := invitation.Gallery.([]string); ok {
				galleryList = strSlice
			}
		}
		if len(galleryList) == 0 {
			galleryList = []string{
				"https://storage.domain.com/gallery/photo1.jpg",
				"https://storage.domain.com/gallery/photo2.jpg",
				"https://storage.domain.com/gallery/photo3.jpg",
			}
		}
	} else {
		galleryList = []string{}
	}
	payload["gallery"] = galleryList

	// 8. Gift (Conditional Feature Handling)
	var giftMap map[string]interface{}
	if isFeatureFlagEnabled(featuresConfig, "has_gift") {
		giftMap = make(map[string]interface{})
		if g, ok := toMap(invitation.Gift); ok && len(g) > 0 {
			giftMap = g
		} else if invitation.Gifts != nil {
			if giftsList, ok := invitation.Gifts.([]interface{}); ok && len(giftsList) > 0 {
				if firstGift, ok := toMap(giftsList[0]); ok {
					giftMap = firstGift
				}
			}
		}
		if len(giftMap) == 0 {
			giftMap = map[string]interface{}{
				"bank_name":      "BCA",
				"account_number": "127326134",
				"account_name":   "Putu Cempaka Giriyanti Aryawijaya",
				"qr_code_url":    nil,
			}
		} else {
			// Ensure standard keys: bank_name, account_number, account_name
			if _, ok := giftMap["account_name"]; !ok {
				if ah, exists := giftMap["account_holder"]; exists {
					giftMap["account_name"] = ah
				}
			}
			if _, ok := giftMap["qr_code_url"]; !ok {
				giftMap["qr_code_url"] = nil
			}
		}
	}
	payload["gift"] = giftMap

	// 9. Story (Conditional Feature Handling)
	if isFeatureFlagEnabled(featuresConfig, "has_story") {
		if invitation.Story != nil {
			payload["story"] = invitation.Story
		} else if invitation.StoryData != nil {
			payload["story"] = invitation.StoryData
		} else {
			payload["story"] = nil
		}
	} else {
		payload["story"] = nil
	}

	return payload
}

// SubmitRSVP records guest attendance and wishes
func (s *PublicInvitationService) SubmitRSVP(idOrSlug string, req models.PublicRSVPRequest) (map[string]interface{}, error) {
	invitation, err := s.invitationRepo.FindByIDOrSlug(idOrSlug)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, errors.New("invitation not found")
	}

	// Standardize inputs
	guestName := strings.TrimSpace(req.GuestName)
	if guestName == "" {
		guestName = strings.TrimSpace(req.Name)
	}
	if guestName == "" {
		return nil, errors.New("guest name is required")
	}

	attendanceStatus := strings.ToLower(strings.TrimSpace(req.AttendanceStatus))
	if attendanceStatus == "" {
		attendanceStatus = strings.ToLower(strings.TrimSpace(req.Status))
	}
	if attendanceStatus == "" {
		attendanceStatus = strings.ToLower(strings.TrimSpace(req.RSVPStatus))
	}
	if attendanceStatus == "" {
		attendanceStatus = "hadir"
	}

	numberOfGuests := req.NumberOfGuests
	if numberOfGuests <= 0 {
		numberOfGuests = req.Pax
	}
	if numberOfGuests <= 0 {
		numberOfGuests = 1
	}

	message := strings.TrimSpace(req.Message)

	// Save or update Guest
	guests, _ := s.guestRepo.FindByInvitationID(invitation.ID)
	var matchedGuest *models.Guest
	for i := range guests {
		if strings.EqualFold(strings.TrimSpace(guests[i].Name), guestName) {
			matchedGuest = &guests[i]
			break
		}
	}

	if matchedGuest != nil {
		matchedGuest.RSVPStatus = attendanceStatus
		matchedGuest.Pax = numberOfGuests
		if message != "" {
			matchedGuest.Message = &message
		}
		_ = s.guestRepo.Update(matchedGuest)
	} else {
		newGuest := models.Guest{
			InvitationID: invitation.ID,
			Name:         guestName,
			Pax:          numberOfGuests,
			QRToken:      fmt.Sprintf("qr_%s", uuid.New().String()[:8]),
			RSVPStatus:   attendanceStatus,
		}
		if message != "" {
			newGuest.Message = &message
		}
		_ = s.guestRepo.Create(&newGuest)
	}

	// If message provided, record into wishes table
	if message != "" {
		wish := models.Wish{
			InvitationID:     invitation.ID,
			GuestName:        guestName,
			AttendanceStatus: attendanceStatus,
			NumberOfGuests:   numberOfGuests,
			Message:          message,
		}
		_ = s.wishRepo.Create(&wish)
	}

	return map[string]interface{}{
		"guest_name":        guestName,
		"attendance_status": attendanceStatus,
		"number_of_guests":  numberOfGuests,
		"message":           message,
	}, nil
}

// GetWishes returns the guest book stream
func (s *PublicInvitationService) GetWishes(idOrSlug string) ([]models.Wish, error) {
	invitation, err := s.invitationRepo.FindByIDOrSlug(idOrSlug)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, errors.New("invitation not found")
	}

	return s.wishRepo.FindByInvitationID(invitation.ID)
}

// Helper utilities
func toMap(v interface{}) (map[string]interface{}, bool) {
	if v == nil {
		return nil, false
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m, true
	}
	// Re-marshal if it's a struct or json.RawMessage
	b, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var res map[string]interface{}
	if err := json.Unmarshal(b, &res); err == nil && res != nil {
		return res, true
	}
	return nil, false
}

func getString(m map[string]interface{}, key, fallback string) string {
	if val, ok := m[key]; ok && val != nil {
		s := fmt.Sprintf("%v", val)
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return fallback
}

func isFeatureFlagEnabled(features map[string]interface{}, featureKey string) bool {
	if features == nil {
		return false
	}
	val, exists := features[featureKey]
	if !exists || val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true") || v == "1"
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	default:
		return false
	}
}

