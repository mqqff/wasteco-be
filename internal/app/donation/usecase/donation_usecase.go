package usecase

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	donationRepo "github.com/mqqff/wasteco-be/internal/app/donation/repository"
	pantryRepo "github.com/mqqff/wasteco-be/internal/app/pantry/repository"
	userRepo "github.com/mqqff/wasteco-be/internal/app/user/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("donation request not found")
	ErrForbidden     = errors.New("forbidden")
	ErrInvalidState  = errors.New("donation request is no longer available for this action")
	ErrPhoneRequired = errors.New("whatsapp number is required")
	ErrInvalidItems  = errors.New("selected items do not belong to this donation request")
)

const defaultRadiusKM = 5.0

type DonationUsecaseItf interface {
	CreateRequest(userID uuid.UUID, req dto.CreateDonationRequest) (dto.DonationResponse, error)
	GetNearbyRequests(userID uuid.UUID, search string, userLat, userLong, radiusKM float64) ([]dto.DonationResponse, error)
	GetRequestDetail(userID, id uuid.UUID, userLat, userLong float64) (dto.DonationResponse, error)
	GetMyRequests(userID uuid.UUID) (dto.MyDonationsResponse, error)
	Fulfill(donorID, id uuid.UUID, req dto.FulfillDonationRequest) (dto.DonationResponse, error)
	ConfirmReceived(requesterID, id uuid.UUID) (dto.DonationResponse, error)
	Cancel(requesterID, id uuid.UUID) error
}

type DonationUsecase struct {
	donationRepo donationRepo.DonationRepoItf
	pantryRepo   pantryRepo.PantryRepoItf
	userRepo     userRepo.UserRepoItf
}

func NewDonationUsecase(d donationRepo.DonationRepoItf, p pantryRepo.PantryRepoItf, us userRepo.UserRepoItf) DonationUsecaseItf {
	return &DonationUsecase{
		donationRepo: d,
		pantryRepo:   p,
		userRepo:     us,
	}
}

func (u *DonationUsecase) CreateRequest(userID uuid.UUID, req dto.CreateDonationRequest) (dto.DonationResponse, error) {
	requester, err := u.userRepo.Show(dto.UserParam{ID: userID})
	if err != nil {
		return dto.DonationResponse{}, err
	}
	if requester.PhoneNumber == "" {
		return dto.DonationResponse{}, ErrPhoneRequired
	}

	request := entity.DonationRequest{
		ID:           uuid.New(),
		RequesterID:  userID,
		Title:        strings.TrimSpace(req.Title),
		Note:         req.Note,
		AreaName:     req.AreaName,
		MeetingPoint: req.MeetingPoint,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		ReceiveStart: req.ReceiveStart,
		ReceiveEnd:   req.ReceiveEnd,
		Status:       "open",
	}
	if request.AreaName == "" {
		request.AreaName = requester.AreaName
	}

	names := make([]string, 0, len(req.Items))
	for _, item := range req.Items {
		request.Items = append(request.Items, entity.DonationRequestItem{
			ID:       uuid.New(),
			Name:     item.Name,
			Quantity: item.Quantity,
			Unit:     item.Unit,
		})
		names = append(names, item.Name)
	}
	if request.Title == "" {
		request.Title = defaultTitle(names)
	}

	if err := u.donationRepo.Create(&request); err != nil {
		return dto.DonationResponse{}, err
	}

	request.Requester = requester
	return mapToResponse(request, userID, req.Latitude, req.Longitude, true), nil
}

func (u *DonationUsecase) GetNearbyRequests(userID uuid.UUID, search string, userLat, userLong, radiusKM float64) ([]dto.DonationResponse, error) {
	if userLat == 0 || userLong == 0 || radiusKM <= 0 {
		user, err := u.userRepo.Show(dto.UserParam{ID: userID})
		if err != nil {
			return nil, err
		}
		if userLat == 0 || userLong == 0 {
			userLat, userLong = user.Latitude, user.Longitude
		}
		if radiusKM <= 0 {
			radiusKM = float64(user.SharingRadiusKM)
		}
	}
	if radiusKM <= 0 {
		radiusKM = defaultRadiusKM
	}

	requests, err := u.donationRepo.FindOpen(userID, search)
	if err != nil {
		return nil, err
	}

	results := make([]dto.DonationResponse, 0, len(requests))
	for _, r := range requests {
		resp := mapToResponse(r, userID, userLat, userLong, false)
		if resp.DistanceKM != nil && *resp.DistanceKM > radiusKM {
			continue
		}
		results = append(results, resp)
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].DistanceKM == nil {
			return false
		}
		if results[j].DistanceKM == nil {
			return true
		}
		return *results[i].DistanceKM < *results[j].DistanceKM
	})

	return results, nil
}

func (u *DonationUsecase) GetRequestDetail(userID, id uuid.UUID, userLat, userLong float64) (dto.DonationResponse, error) {
	request, err := u.findRequest(id)
	if err != nil {
		return dto.DonationResponse{}, err
	}

	if userLat == 0 || userLong == 0 {
		if user, err := u.userRepo.Show(dto.UserParam{ID: userID}); err == nil {
			userLat, userLong = user.Latitude, user.Longitude
		}
	}

	return mapToResponse(request, userID, userLat, userLong, true), nil
}

func (u *DonationUsecase) GetMyRequests(userID uuid.UUID) (dto.MyDonationsResponse, error) {
	requests, err := u.donationRepo.FindByRequesterID(userID)
	if err != nil {
		return dto.MyDonationsResponse{}, err
	}

	res := dto.MyDonationsResponse{
		Active:  []dto.DonationResponse{},
		History: []dto.DonationResponse{},
	}
	for _, r := range requests {
		resp := mapToResponse(r, userID, 0, 0, true)
		if r.Status == "open" || r.Status == "handed_over" {
			res.Active = append(res.Active, resp)
		} else {
			res.History = append(res.History, resp)
		}
	}

	return res, nil
}

func (u *DonationUsecase) Fulfill(donorID, id uuid.UUID, req dto.FulfillDonationRequest) (dto.DonationResponse, error) {
	request, err := u.findRequest(id)
	if err != nil {
		return dto.DonationResponse{}, err
	}
	if request.RequesterID == donorID {
		return dto.DonationResponse{}, ErrForbidden
	}
	if request.Status != "open" || time.Now().After(request.ReceiveEnd) {
		return dto.DonationResponse{}, ErrInvalidState
	}

	donor, err := u.userRepo.Show(dto.UserParam{ID: donorID})
	if err != nil {
		return dto.DonationResponse{}, err
	}
	if donor.PhoneNumber == "" {
		return dto.DonationResponse{}, ErrPhoneRequired
	}

	byID := make(map[string]entity.DonationRequestItem, len(request.Items))
	for _, item := range request.Items {
		byID[item.ID.String()] = item
	}

	pantry, err := u.pantryRepo.FindAllByUserID(donorID, "")
	if err != nil {
		return dto.DonationResponse{}, err
	}

	selected := make([]entity.DonationRequestItem, 0, len(req.ItemIDs))
	seen := make(map[string]bool, len(req.ItemIDs))
	for _, itemID := range req.ItemIDs {
		item, ok := byID[itemID]
		if !ok {
			return dto.DonationResponse{}, ErrInvalidItems
		}
		if seen[itemID] {
			continue
		}
		seen[itemID] = true

		if match := matchPantryItem(item.Name, pantry); match != nil {
			item.PantryIngredientID = &match.ID
		}
		selected = append(selected, item)
	}

	if err := u.donationRepo.Fulfill(id, donorID, selected); err != nil {
		if errors.Is(err, donationRepo.ErrStateChanged) {
			return dto.DonationResponse{}, ErrInvalidState
		}
		return dto.DonationResponse{}, err
	}

	return u.GetRequestDetail(donorID, id, 0, 0)
}

func (u *DonationUsecase) ConfirmReceived(requesterID, id uuid.UUID) (dto.DonationResponse, error) {
	request, err := u.findRequest(id)
	if err != nil {
		return dto.DonationResponse{}, err
	}
	if request.RequesterID != requesterID {
		return dto.DonationResponse{}, ErrForbidden
	}
	if request.Status != "handed_over" || request.DonorID == nil {
		return dto.DonationResponse{}, ErrInvalidState
	}

	donorID := *request.DonorID
	portions := 0
	breakdown := make([]dto.ImpactItemBreakdown, 0, len(request.Items))
	deductions := make([]donationRepo.PantryDeduction, 0, len(request.Items))

	for _, item := range request.Items {
		if !item.Selected {
			continue
		}
		portions++
		breakdown = append(breakdown, dto.ImpactItemBreakdown{
			Name:     item.Name,
			Quantity: item.Quantity,
			Unit:     item.Unit,
		})

		if item.PantryIngredientID != nil {
			deductions = append(deductions, donationRepo.PantryDeduction{
				PantryIngredientID: *item.PantryIngredientID,
				Quantity:           item.Quantity,
			})
		}
	}

	breakdownJSON, _ := sonic.Marshal(breakdown)
	activity := entity.ImpactActivity{
		ID:             uuid.New(),
		UserID:         donorID,
		ActivityType:   "donation",
		ReferenceID:    &request.ID,
		Title:          fmt.Sprintf("%s (%d item)", request.Title, portions),
		PortionsSaved:  portions,
		CO2PreventedKg: float64(portions) * 0.21,
		ItemsBreakdown: string(breakdownJSON),
	}
	if err := u.donationRepo.CompleteReceipt(id, requesterID, donorID, &activity, deductions); err != nil {
		if errors.Is(err, donationRepo.ErrStateChanged) {
			return dto.DonationResponse{}, ErrInvalidState
		}
		return dto.DonationResponse{}, err
	}

	return u.GetRequestDetail(requesterID, id, 0, 0)
}

func (u *DonationUsecase) Cancel(requesterID, id uuid.UUID) error {
	request, err := u.findRequest(id)
	if err != nil {
		return err
	}
	if request.RequesterID != requesterID {
		return ErrForbidden
	}

	if err := u.donationRepo.Cancel(id, requesterID); err != nil {
		if errors.Is(err, donationRepo.ErrStateChanged) {
			return ErrInvalidState
		}
		return err
	}

	return nil
}

func (u *DonationUsecase) findRequest(id uuid.UUID) (entity.DonationRequest, error) {
	request, err := u.donationRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return entity.DonationRequest{}, ErrNotFound
	}

	return request, err
}

func matchPantryItem(name string, pantry []entity.PantryIngredient) *entity.PantryIngredient {
	target := strings.ToLower(strings.TrimSpace(name))
	for i := range pantry {
		if strings.ToLower(pantry[i].Name) == target {
			return &pantry[i]
		}
	}
	for i := range pantry {
		candidate := strings.ToLower(pantry[i].Name)
		if strings.Contains(candidate, target) || strings.Contains(target, candidate) {
			return &pantry[i]
		}
	}

	return nil
}

func defaultTitle(names []string) string {
	if len(names) > 2 {
		names = names[:2]
	}
	return strings.Join(names, " & ")
}

func mapToResponse(r entity.DonationRequest, viewerID uuid.UUID, userLat, userLong float64, withContact bool) dto.DonationResponse {
	resp := dto.DonationResponse{
		ID:           r.ID.String(),
		Title:        r.Title,
		Note:         r.Note,
		AreaName:     r.AreaName,
		MeetingPoint: r.MeetingPoint,
		Latitude:     r.Latitude,
		Longitude:    r.Longitude,
		ReceiveStart: r.ReceiveStart.Format(time.RFC3339),
		ReceiveEnd:   r.ReceiveEnd.Format(time.RFC3339),
		Status:       r.Status,
		ItemCount:    len(r.Items),
		Items:        make([]dto.DonationItemResponse, 0, len(r.Items)),
		CreatedAt:    r.CreatedAt.Format(time.RFC3339),
	}

	for _, item := range r.Items {
		resp.Items = append(resp.Items, dto.DonationItemResponse{
			ID:       item.ID.String(),
			Name:     item.Name,
			Quantity: item.Quantity,
			Unit:     item.Unit,
			Selected: item.Selected,
		})
	}

	isParticipant := viewerID == r.RequesterID || (r.DonorID != nil && viewerID == *r.DonorID)
	showContact := withContact && (isParticipant || r.Status == "open")

	if r.Requester.ID != uuid.Nil {
		resp.Requester = toDonationUser(r.Requester, showContact)
	}
	if r.Donor != nil && r.Donor.ID != uuid.Nil {
		resp.Donor = toDonationUser(*r.Donor, showContact)
	}

	if r.HandedOverAt != nil {
		resp.HandedOverAt = r.HandedOverAt.Format(time.RFC3339)
	}
	if r.ReceivedAt != nil {
		resp.ReceivedAt = r.ReceivedAt.Format(time.RFC3339)
	}

	if userLat != 0 && userLong != 0 && r.Latitude != 0 && r.Longitude != 0 {
		dist := haversine(userLat, userLong, r.Latitude, r.Longitude)
		resp.DistanceKM = &dist
		resp.FormattedDistance = formatDistance(dist)
	}

	return resp
}

func toDonationUser(user entity.User, withContact bool) *dto.DonationUser {
	res := &dto.DonationUser{
		ID:       user.ID.String(),
		Name:     user.Name,
		AreaName: user.AreaName,
	}
	if withContact && user.PhoneNumber != "" {
		res.WhatsAppNumber = user.PhoneNumber
		res.WhatsAppURL = whatsAppURL(user.PhoneNumber)
	}

	return res
}

func whatsAppURL(phone string) string {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)
	if strings.HasPrefix(digits, "0") {
		digits = "62" + digits[1:]
	}

	return "https://wa.me/" + digits
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
