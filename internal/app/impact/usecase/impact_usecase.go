package usecase

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/impact/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
)

type ImpactUsecaseItf interface {
	GetMonthlySummary(userID uuid.UUID, month, year int) (dto.ImpactSummaryResponse, error)
	GetHistory(userID uuid.UUID, page, limit int) (dto.ImpactHistoryResponse, error)
	GetDetail(userID uuid.UUID, id uuid.UUID) (dto.ImpactActivityResponse, error)
}

type ImpactUsecase struct {
	impactRepo repository.ImpactRepoItf
}

func NewImpactUsecase(r repository.ImpactRepoItf) ImpactUsecaseItf {
	return &ImpactUsecase{
		impactRepo: r,
	}
}

var englishMonths = [...]string{
	"", "January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

func (u *ImpactUsecase) GetMonthlySummary(userID uuid.UUID, month, year int) (dto.ImpactSummaryResponse, error) {
	now := time.Now()
	if year == 0 {
		year = now.Year()
	}
	if month < 1 || month > 12 {
		month = int(now.Month())
	}

	summary, err := u.impactRepo.GetMonthlySummary(userID, month, year)
	if err != nil {
		return dto.ImpactSummaryResponse{}, err
	}

	monthName := fmt.Sprintf("%s %d", englishMonths[month], year)

	var cookingPct, donationPct float64
	totalMoney := summary.CookingMoney + summary.DonationMoney
	if totalMoney > 0 {
		cookingPct = math.Round((summary.CookingMoney/totalMoney)*1000) / 10
		donationPct = math.Round((summary.DonationMoney/totalMoney)*1000) / 10
	}

	recentDB, _, _ := u.impactRepo.FindAllByUserID(userID, 1, 3)
	var recentResp []dto.ImpactActivityResponse
	for _, act := range recentDB {
		recentResp = append(recentResp, u.mapToResponse(act, false))
	}

	return dto.ImpactSummaryResponse{
		Month:               month,
		Year:                year,
		MonthName:           monthName,
		TotalPortionsSaved:  summary.TotalPortions,
		TotalMoneySaved:     summary.TotalMoney,
		TotalCO2PreventedKg: summary.TotalCO2,
		Composition: dto.ImpactComposition{
			CookingPortions:    summary.CookingPortions,
			CookingMoneySaved:  summary.CookingMoney,
			CookingPercentage:  cookingPct,
			DonationPortions:   summary.DonationPortions,
			DonationMoneySaved: summary.DonationMoney,
			DonationPercentage: donationPct,
		},
		RecentActivities: recentResp,
	}, nil
}

func (u *ImpactUsecase) GetHistory(userID uuid.UUID, page, limit int) (dto.ImpactHistoryResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	items, total, err := u.impactRepo.FindAllByUserID(userID, page, limit)
	if err != nil {
		return dto.ImpactHistoryResponse{}, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	var activities []dto.ImpactActivityResponse
	for _, act := range items {
		activities = append(activities, u.mapToResponse(act, false))
	}

	return dto.ImpactHistoryResponse{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
		Activities: activities,
	}, nil
}

func (u *ImpactUsecase) GetDetail(userID uuid.UUID, id uuid.UUID) (dto.ImpactActivityResponse, error) {
	item, err := u.impactRepo.FindByID(id)
	if err != nil {
		return dto.ImpactActivityResponse{}, err
	}

	if item.UserID != userID {
		return dto.ImpactActivityResponse{}, errors.New("unauthorized")
	}

	return u.mapToResponse(item, true), nil
}

func (u *ImpactUsecase) mapToResponse(act entity.ImpactActivity, includeBreakdown bool) dto.ImpactActivityResponse {
	resp := dto.ImpactActivityResponse{
		ID:             act.ID.String(),
		ActivityType:   act.ActivityType,
		Title:          act.Title,
		PortionsSaved:  act.PortionsSaved,
		MoneySaved:     act.MoneySaved,
		CO2PreventedKg: act.CO2PreventedKg,
		FormattedTime:  formatActivityTime(act.CreatedAt),
		CreatedAt:      act.CreatedAt.Format(time.RFC3339),
	}

	if includeBreakdown && act.ItemsBreakdown != "" {
		var breakdown []dto.ImpactItemBreakdown
		_ = sonic.Unmarshal([]byte(act.ItemsBreakdown), &breakdown)
		resp.ItemsBreakdown = breakdown
	}

	return resp
}

func formatActivityTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "Just now"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour && t.Day() == now.Day() {
		return fmt.Sprintf("Today, %s", t.Format("15:04"))
	}
	if diff < 48*time.Hour && t.Day() == now.Day()-1 {
		return fmt.Sprintf("Yesterday, %s", t.Format("15:04"))
	}

	return fmt.Sprintf("%d days ago, %s", int(diff.Hours()/24), t.Format("15:04"))
}
