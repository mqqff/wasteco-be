package main

import (
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/internal/infra/mysql"
	"github.com/mqqff/wasteco-be/pkg/log"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	db, err := mysql.NewConn()
	if err != nil {
		return
	}

	log.Info(nil, "[Seeder] Starting data seeding...")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	affanID := uuid.New()
	budiID := uuid.New()
	sitiID := uuid.New()

	users := []entity.User{
		{
			ID:              affanID,
			Name:            "Affan Maulana",
			Email:           "affan@wasteco.id",
			PasswordHash:    string(hashedPassword),
			PhoneNumber:     "+6281234567890",
			AreaName:        "Dinoyo, Lowokwaru, Malang",
			Latitude:        -7.9528,
			Longitude:       112.6074,
			SharingRadiusKM: 3,
		},
		{
			ID:              budiID,
			Name:            "Budi Santoso",
			Email:           "budi@wasteco.id",
			PasswordHash:    string(hashedPassword),
			PhoneNumber:     "+6281358902144",
			AreaName:        "Dinoyo, Malang",
			Latitude:        -7.9535,
			Longitude:       112.6080,
			SharingRadiusKM: 2,
		},
		{
			ID:              sitiID,
			Name:            "Siti Aminah",
			Email:           "siti@wasteco.id",
			PasswordHash:    string(hashedPassword),
			PhoneNumber:     "+6285712345678",
			AreaName:        "Lowokwaru, Malang",
			Latitude:        -7.9450,
			Longitude:       112.6150,
			SharingRadiusKM: 3,
		},
	}

	for _, u := range users {
		var existing entity.User
		if err := db.Where("email = ?", u.Email).First(&existing).Error; err == nil {
			if u.Email == "affan@wasteco.id" {
				affanID = existing.ID
			} else if u.Email == "budi@wasteco.id" {
				budiID = existing.ID
			} else if u.Email == "siti@wasteco.id" {
				sitiID = existing.ID
			}
		} else {
			db.Create(&u)
		}
	}

	now := time.Now()
	nearExpiry := now.AddDate(0, 0, 2)
	freshDate := now.AddDate(0, 0, 7)

	pantryItems := []entity.PantryIngredient{
		{
			ID:              uuid.New(),
			UserID:          affanID,
			Name:            "Nasi Putih",
			Quantity:        2,
			Unit:            "porsi",
			Category:        "Bahan Pokok",
			StorageLocation: "Kulkas bagian bawah",
			ExpiryDate:      &freshDate,
			Status:          "fresh",
		},
		{
			ID:              uuid.New(),
			UserID:          affanID,
			Name:            "Telur Ayam",
			Quantity:        2,
			Unit:            "butir",
			Category:        "Lauk / Protein",
			StorageLocation: "Pintu kulkas",
			ExpiryDate:      &freshDate,
			Status:          "fresh",
		},
		{
			ID:              uuid.New(),
			UserID:          affanID,
			Name:            "Wortel",
			Quantity:        100,
			Unit:            "gram",
			Category:        "Sayuran",
			StorageLocation: "Kulkas bagian bawah",
			ExpiryDate:      &nearExpiry,
			Status:          "near_expiry",
		},
		{
			ID:              uuid.New(),
			UserID:          affanID,
			Name:            "Roti Tawar",
			Quantity:        3,
			Unit:            "lembar",
			Category:        "Bahan Pokok",
			StorageLocation: "Suhu Ruang",
			ExpiryDate:      &nearExpiry,
			Status:          "near_expiry",
		},
	}

	for _, item := range pantryItems {
		var count int64
		db.Model(&entity.PantryIngredient{}).Where("user_id = ? AND name = ?", item.UserID, item.Name).Count(&count)
		if count == 0 {
			db.Create(&item)
		}
	}

	receiveStart := now.Truncate(time.Hour).Add(time.Hour)
	receiveEnd := receiveStart.Add(7 * 24 * time.Hour)

	requests := []entity.DonationRequest{
		{
			ID:           uuid.New(),
			RequesterID:  budiID,
			Title:        "Wortel & Telur",
			Note:         "Untuk masak makan malam keluarga. Bantu sebagian item juga boleh.",
			AreaName:     "Dinoyo, Malang",
			MeetingPoint: "Depan Minimarket Dinoyo",
			Latitude:     -7.9535,
			Longitude:    112.6080,
			ReceiveStart: receiveStart,
			ReceiveEnd:   receiveEnd,
			Status:       "open",
			Items: []entity.DonationRequestItem{
				{ID: uuid.New(), Name: "Wortel", Quantity: 500, Unit: "gram"},
				{ID: uuid.New(), Name: "Telur Ayam", Quantity: 4, Unit: "butir"},
				{ID: uuid.New(), Name: "Nasi Putih", Quantity: 2, Unit: "porsi"},
			},
		},
		{
			ID:           uuid.New(),
			RequesterID:  sitiID,
			Title:        "Kentang",
			Note:         "Mau bikin perkedel untuk bekal anak.",
			AreaName:     "Dinoyo, Malang",
			MeetingPoint: "Teras rumah dekat Masjid Dinoyo",
			Latitude:     -7.9500,
			Longitude:    112.6050,
			ReceiveStart: receiveStart,
			ReceiveEnd:   receiveEnd,
			Status:       "open",
			Items: []entity.DonationRequestItem{
				{ID: uuid.New(), Name: "Kentang", Quantity: 1, Unit: "kg"},
			},
		},
		{
			ID:           uuid.New(),
			RequesterID:  budiID,
			Title:        "Tomat & Bumbu",
			AreaName:     "Dinoyo, Malang",
			MeetingPoint: "Gerbang Kos Hijau Jl. MT Haryono",
			Latitude:     -7.9560,
			Longitude:    112.6110,
			ReceiveStart: receiveStart,
			ReceiveEnd:   receiveEnd,
			Status:       "open",
			Items: []entity.DonationRequestItem{
				{ID: uuid.New(), Name: "Tomat", Quantity: 3, Unit: "buah"},
				{ID: uuid.New(), Name: "Bawang Merah", Quantity: 100, Unit: "gram"},
				{ID: uuid.New(), Name: "Cabai Rawit", Quantity: 50, Unit: "gram"},
			},
		},
	}

	for _, r := range requests {
		var count int64
		db.Model(&entity.DonationRequest{}).Where("requester_id = ? AND title = ?", r.RequesterID, r.Title).Count(&count)
		if count == 0 {
			db.Create(&r)
		}
	}

	breakdown1, _ := sonic.Marshal([]map[string]interface{}{
		{"name": "Nasi Putih", "quantity": 2, "unit": "porsi"},
		{"name": "Telur Ayam", "quantity": 2, "unit": "butir"},
		{"name": "Wortel", "quantity": 100, "unit": "gram"},
	})

	breakdown2, _ := sonic.Marshal([]map[string]interface{}{
		{"name": "Labu Siam & Buncis", "quantity": 700, "unit": "gram"},
	})

	breakdown3, _ := sonic.Marshal([]map[string]interface{}{
		{"name": "Sup Telur Puyuh & Wortel", "quantity": 3, "unit": "porsi"},
	})

	breakdown4, _ := sonic.Marshal([]map[string]interface{}{
		{"name": "Wortel Segar", "quantity": 500, "unit": "gram"},
	})

	activities := []entity.ImpactActivity{
		{
			ID:             uuid.New(),
			UserID:         affanID,
			ActivityType:   "cooking",
			Title:          "Nasi Goreng Wortel Telur",
			PortionsSaved:  2,
			CO2PreventedKg: 0.42,
			ItemsBreakdown: string(breakdown1),
			CreatedAt:      now.Add(-2 * time.Hour),
		},
		{
			ID:             uuid.New(),
			UserID:         affanID,
			ActivityType:   "cooking",
			Title:          "Tumis Labu Siam & Buncis",
			PortionsSaved:  2,
			CO2PreventedKg: 0.42,
			ItemsBreakdown: string(breakdown2),
			CreatedAt:      now.Add(-24 * time.Hour),
		},
		{
			ID:             uuid.New(),
			UserID:         affanID,
			ActivityType:   "cooking",
			Title:          "Sup Telur Puyuh & Wortel",
			PortionsSaved:  3,
			CO2PreventedKg: 0.63,
			ItemsBreakdown: string(breakdown3),
			CreatedAt:      now.Add(-72 * time.Hour),
		},
		{
			ID:             uuid.New(),
			UserID:         affanID,
			ActivityType:   "donation",
			Title:          "Wortel Segar (500 gram)",
			PortionsSaved:  2,
			CO2PreventedKg: 0.42,
			ItemsBreakdown: string(breakdown4),
			CreatedAt:      now.Add(-5 * time.Hour),
		},
	}

	for _, act := range activities {
		var count int64
		db.Model(&entity.ImpactActivity{}).Where("user_id = ? AND title = ?", act.UserID, act.Title).Count(&count)
		if count == 0 {
			db.Create(&act)
		}
	}

	log.Info(nil, "[Seeder] All dummy seed data inserted successfully!")
}
