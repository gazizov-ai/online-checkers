package handler

import (
	"errors"
	"net/http"

	"github.com/gazizov-ai/online-checkers/pkg/httpx"
	appjwt "github.com/gazizov-ai/online-checkers/pkg/jwt"
	"github.com/gazizov-ai/online-checkers/services/profile/internal/domain"
	"github.com/gazizov-ai/online-checkers/services/profile/internal/repository"
	"github.com/gazizov-ai/online-checkers/services/profile/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	service *service.ProfileService
}

func NewProfileHandler(service *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{service: service}
}

type profileResponse struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name,omitempty"`
	CountryCode *string `json:"country_code,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type updateMyProfileRequest struct {
	DisplayName *string `json:"display_name"`
	CountryCode *string `json:"country_code"`
	AvatarURL   *string `json:"avatar_url"`
	Bio         *string `json:"bio"`
}

type batchProfilesRequest struct {
	UserIDs []string `json:"user_ids"`
}

type batchProfilesResponse struct {
	Profiles []profileResponse `json:"profiles"`
}

func (h *ProfileHandler) GetProfileByUserID(w http.ResponseWriter, r *http.Request) {
	rawUserID := chi.URLParam(r, "user_id")

	userID, err := uuid.Parse(rawUserID)
	if err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid user_id")
		return
	}

	profile, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			_ = httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "profile not found")
			return
		}

		_ = httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, toProfileResponse(profile))
}

func toProfileResponse(profile domain.Profile) profileResponse {
	return profileResponse{
		UserID:      profile.UserID.String(),
		Username:    profile.Username,
		DisplayName: profile.DisplayName,
		CountryCode: profile.CountryCode,
		AvatarURL:   profile.AvatarURL,
		Bio:         profile.Bio,
		CreatedAt:   profile.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   profile.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *ProfileHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := appjwt.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	profile, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			_ = httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "profile not found")
			return
		}

		_ = httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, toProfileResponse(profile))
}

func (h *ProfileHandler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := appjwt.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var req updateMyProfileRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	profile, err := h.service.UpdateProfile(r.Context(), userID, domain.UpdateProfileInput{
		DisplayName: req.DisplayName,
		CountryCode: req.CountryCode,
		AvatarURL:   req.AvatarURL,
		Bio:         req.Bio,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrProfileNotFound):
			_ = httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "profile not found")
		case errors.Is(err, service.ErrInvalidDisplayName):
			_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_display_name", "invalid display name")
		case errors.Is(err, service.ErrInvalidCountryCode):
			_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_country_code", "invalid country code")
		case errors.Is(err, service.ErrInvalidAvatarURL):
			_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_avatar_url", "invalid avatar url")
		case errors.Is(err, service.ErrInvalidBio):
			_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_bio", "invalid bio")
		default:
			_ = httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
		}
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, toProfileResponse(profile))
}

func (h *ProfileHandler) BatchProfiles(w http.ResponseWriter, r *http.Request) {
	var req batchProfilesRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	userIDs := make([]uuid.UUID, 0, len(req.UserIDs))
	for _, rawID := range req.UserIDs {
		id, err := uuid.Parse(rawID)
		if err != nil {
			_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid user_id")
			return
		}

		userIDs = append(userIDs, id)
	}

	profiles, err := h.service.GetProfiles(r.Context(), userIDs)
	if err != nil {
		if errors.Is(err, service.ErrTooManyProfilesRequested) {
			_ = httpx.WriteError(w, http.StatusBadRequest, "too_many_profiles", "too many profiles requested")
			return
		}

		_ = httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	responseProfiles := make([]profileResponse, 0, len(profiles))
	for _, profile := range profiles {
		responseProfiles = append(responseProfiles, toProfileResponse(profile))
	}

	_ = httpx.WriteJSON(w, http.StatusOK, batchProfilesResponse{
		Profiles: responseProfiles,
	})
}
