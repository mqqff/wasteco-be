package usecase

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"path/filepath"
	"sort"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	donationRepo "github.com/mqqff/wasteco-be/internal/app/donation/repository"
	impactRepo "github.com/mqqff/wasteco-be/internal/app/impact/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/pkg/supabase"
)

type DonationUsecaseItf interface {
	CreateDonation(userID uuid.UUID, req dto.CreateDonationRequest, files []*multipart.FileHeader) (dto.DonationResponse, error)
	GetAvailableDonations(category, search string, userLat, userLong, radiusKM float64) ([]dto.DonationResponse, error)
	GetDonationDetail(id uuid.UUID, userLat, userLong float64) (dto.DonationResponse, error)
	GetMyDonations(userID uuid.UUID) (dto.MyDonationsResponse, error)
	UpdateStatus(userID uuid.UUID, id uuid.UUID, status string) (dto.DonationResponse, error)
	ConfirmHandedOver(userID uuid.UUID, id uuid.UUID) (dto.DonationResponse, error)
	CancelDonation(userID uuid.UUID, id uuid.UUID) error
}

type DonationUsecase struct {
	donationRepo donationRepo.DonationRepoItf
	impactRepo   impactRepo.ImpactRepoItf
	supabase     supabase.SupabaseItf
}

func NewDonationUsecase(d donationRepo.DonationRepoItf, i impactRepo.ImpactRepoItf, s supabase.SupabaseItf) DonationUsecaseItf {
	return &DonationUsecase{
		donationRepo: d,
		impactRepo:   i,
		supabase:     s,
	}
}

func (u *DonationUsecase) CreateDonation(userID uuid.UUID, req dto.CreateDonationRequest, files []*multipart.FileHeader) (dto.DonationResponse, error) {
	var photoURLs []string

	// Upload photos (max 3)
	for i, fileHeader := range files {
		if i >= 3 {
			break
		}

		file, err := fileHeader.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			continue
		}

		ext := filepath.Ext(fileHeader.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		path := fmt.Sprintf("donations/%s-%d%s", uuid.NewString(), i, ext)
		mimeType := fileHeader.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = "image/jpeg"
		}

		publicURL, err := u.supabase.UploadFile("donations", path, mimeType, bytes.NewReader(data))
		if err == nil && publicURL != "" {
			photoURLs = append(photoURLs, publicURL)
		}
	}

	photoJSON, _ := sonic.Marshal(photoURLs)

	var expiryDate *time.Time
	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err == nil {
			expiryDate = &parsed
		}
	}

	var pickupDeadline *time.Time
	if req.PickupDeadline != "" {
		parsed, err := time.Parse(time.RFC3339, req.PickupDeadline)
		if err == nil {
			pickupDeadline = &parsed
		} else {
			// fallback attempt
			parsed2, err2 := time.Parse("2006-01-02 15:04:05", req.PickupDeadline)
			if err2 == nil {
				pickupDeadline = &parsed2
			}
		}
	}

	portions := req.EstimatedPortions
	if portions <= 0 {
		portions = 2
	}

	donation := entity.Donation{
		ID:                uuid.New(),
		UserID:            userID,
		Title:             req.Title,
		Category:          req.Category,
		Quantity:          req.Quantity,
		Unit:              req.Unit,
		PhotoURLs:         string(photoJSON),
		ConditionNote:     req.ConditionNote,
		StorageMethod:     req.StorageMethod,
		ExpiryDate:        expiryDate,
		PickupDeadline:    pickupDeadline,
		AreaName:          req.AreaName,
		MeetingPoint:      req.MeetingPoint,
		Latitude:          req.Latitude,
		Longitude:         req.Longitude,
		WhatsAppNumber:    req.WhatsAppNumber,
		EstimatedPrice:    req.EstimatedPrice,
		EstimatedPortions: portions,
		Status:            "available",
	}

	err := u.donationRepo.Create(&donation)
	if err != nil {
		return dto.DonationResponse{}, err
	}

	return u.mapToResponse(donation, 0, 0), nil
}

func (u *DonationUsecase) GetAvailableDonations(category, search string, userLat, userLong, radiusKM float64) ([]dto.DonationResponse, error) {
	donations, err := u.donationRepo.FindAllAvailable(category, search)
	if err != nil {
		return nil, err
	}

	var results []dto.DonationResponse

	for _, d := range donations {
		var dist float64
		hasCoord := userLat != 0 && userLong != 0 && d.Latitude != 0 && d.Longitude != 0

		if hasCoord {
			dist = haversine(userLat, userLong, d.Latitude, d.Longitude)
			if radiusKM > 0 && dist > radiusKM {
				continue
			}
		}

		resp := u.mapToResponse(d, userLat, userLong)
		results = append(results, resp)
	}

	// Sort by distance if user coords provided
	if userLat != 0 && userLong != 0 {
		sort.Slice(results, func(i, j int) bool {
			if results[i].DistanceKM == nil {
				return false
			}
			if results[j].DistanceKM == nil {
				return true
			}
			return *results[i].DistanceKM < *results[j].DistanceKM
		})
	}

	return results, nil
}

func (u *DonationUsecase) GetDonationDetail(id uuid.UUID, userLat, userLong float64) (dto.DonationResponse, error) {
	donation, err := u.donationRepo.FindByID(id)
	if err != nil {
		return dto.DonationResponse{}, err
	}

	return u.mapToResponse(donation, userLat, userLong), nil
}

func (u *DonationUsecase) GetMyDonations(userID uuid.UUID) (dto.MyDonationsResponse, error) {
	donations, err := u.donationRepo.FindByUserID(userID)
	if err != nil {
		return dto.MyDonationsResponse{}, err
	}

	var active []dto.DonationResponse
	var history []dto.DonationResponse

	for _, d := range donations {
		resp := u.mapToResponse(d, 0, 0)
		if d.Status == "available" || d.Status == "in_progress" {
			active = append(active, resp)
		} else {
			history = append(history, resp)
		}
	}

	return dto.MyDonationsResponse{
		Active:  active,
		History: history,
	}, nil
}

func (u *DonationUsecase) UpdateStatus(userID uuid.UUID, id uuid.UUID, status string) (dto.DonationResponse, error) {
	donation, err := u.donationRepo.FindByID(id)
	if err != nil {
		return dto.DonationResponse{}, err
	}

	if donation.UserID != userID {
		return dto.DonationResponse{}, errors.New("unauthorized")
	}

	donation.Status = status
	err = u.donationRepo.Update(&donation)
	if err != nil {
		return dto.DonationResponse{}, err
	}

	return u.mapToResponse(donation, 0, 0), nil
}

func (u *DonationUsecase) ConfirmHandedOver(userID uuid.UUID, id uuid.UUID) (dto.DonationResponse, error) {
	donation, err := u.donationRepo.FindByID(id)
	if err != nil {
		return dto.DonationResponse{}, err
	}

	if donation.UserID != userID {
		return dto.DonationResponse{}, errors.New("unauthorized")
	}

	if donation.Status == "completed" {
		return u.mapToResponse(donation, 0, 0), nil
	}

	donation.Status = "completed"
	err = u.donationRepo.Update(&donation)
	if err != nil {
		return dto.DonationResponse{}, err
	}

	portions := donation.EstimatedPortions
	if portions <= 0 {
		portions = 2
	}
	co2 := float64(portions) * 0.21

	breakdownJSON, _ := sonic.Marshal([]map[string]interface{}{
		{
			"name":            donation.Title,
			"quantity":        donation.Quantity,
			"unit":            donation.Unit,
			"estimated_price": donation.EstimatedPrice,
		},
	})

	activity := entity.ImpactActivity{
		ID:             uuid.New(),
		UserID:         userID,
		ActivityType:   "donation",
		ReferenceID:    &donation.ID,
		Title:          fmt.Sprintf("%s (%.0f %s)", donation.Title, donation.Quantity, donation.Unit),
		PortionsSaved:  portions,
		MoneySaved:     donation.EstimatedPrice,
		CO2PreventedKg: co2,
		ItemsBreakdown: string(breakdownJSON),
	}

	_ = u.impactRepo.Create(&activity)

	return u.mapToResponse(donation, 0, 0), nil
}

func (u *DonationUsecase) CancelDonation(userID uuid.UUID, id uuid.UUID) error {
	donation, err := u.donationRepo.FindByID(id)
	if err != nil {
		return err
	}

	if donation.UserID != userID {
		return errors.New("unauthorized")
	}

	donation.Status = "cancelled"
	return u.donationRepo.Update(&donation)
}

func (u *DonationUsecase) mapToResponse(d entity.Donation, userLat, userLong float64) dto.DonationResponse {
	var photos []string
	if d.PhotoURLs != "" {
		_ = sonic.Unmarshal([]byte(d.PhotoURLs), &photos)
	}

	resp := dto.DonationResponse{
		ID:                d.ID.String(),
		UserID:            d.UserID.String(),
		Title:             d.Title,
		Category:          d.Category,
		Quantity:          d.Quantity,
		Unit:              d.Unit,
		PhotoURLs:         photos,
		ConditionNote:     d.ConditionNote,
		StorageMethod:     d.StorageMethod,
		AreaName:          d.AreaName,
		MeetingPoint:      d.MeetingPoint,
		Latitude:          d.Latitude,
		Longitude:         d.Longitude,
		WhatsAppNumber:    d.WhatsAppNumber,
		EstimatedPrice:    d.EstimatedPrice,
		EstimatedPortions: d.EstimatedPortions,
		Status:            d.Status,
		CreatedAt:         d.CreatedAt.Format(time.RFC3339),
	}

	if d.User.ID != uuid.Nil {
		resp.Donor = &dto.DonorInfo{
			ID:          d.User.ID.String(),
			Name:        d.User.Name,
			AreaName:    d.User.AreaName,
			PhoneNumber: d.User.PhoneNumber,
		}
	}

	if d.ExpiryDate != nil {
		resp.ExpiryDate = d.ExpiryDate.Format("2006-01-02")
	}

	if d.PickupDeadline != nil {
		resp.PickupDeadline = d.PickupDeadline.Format("15:04 WIB")
	}

	if userLat != 0 && userLong != 0 && d.Latitude != 0 && d.Longitude != 0 {
		dist := haversine(userLat, userLong, d.Latitude, d.Longitude)
		resp.DistanceKM = &dist
		resp.FormattedDistance = formatDistance(dist)
	}

	return resp
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	lat1 = lat1 * (math.Pi / 180.0)
	lat2 = lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func formatDistance(km float64) string {
	if km < 1.0 {
		return fmt.Sprintf("%.0f m", km*1000)
	}
	return fmt.Sprintf("%.1f km", km)
}
