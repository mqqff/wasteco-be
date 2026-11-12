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
			EstimatedPrice:  5000,
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
			EstimatedPrice:  4000,
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
			EstimatedPrice:  3000,
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
			EstimatedPrice:  6000,
		},
	}

	for _, item := range pantryItems {
		var count int64
		db.Model(&entity.PantryIngredient{}).Where("user_id = ? AND name = ?", item.UserID, item.Name).Count(&count)
		if count == 0 {
			db.Create(&item)
		}
	}

	today18 := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, time.Local)
	tomorrow12 := today18.Add(24 * time.Hour)

	donations := []entity.Donation{
		{
			ID:                uuid.New(),
			UserID:            budiID,
			Title:             "Wortel Segar Kebun",
			Category:          "Sayuran",
			Quantity:          500,
			Unit:              "gram",
			PhotoURLs:         `["https://images.unsplash.com/photo-1598170845058-32b9d6a5da37?w=500"]`,
			ConditionNote:     "Sudah dicuci bersih, kulit masih renyah dan tidak lembek. Disimpan di chiller kulkas dalam wadah tertutup.",
			StorageMethod:     "Chiller kulkas wadah tertutup",
			PickupDeadline:    &today18,
			AreaName:          "Dinoyo, Malang",
			MeetingPoint:      "Depan minimarket Alfamart MT Haryono / Gerbang Kos Hijau",
			Latitude:          -7.9535,
			Longitude:         112.6080,
			WhatsAppNumber:    "+6281358902144",
			EstimatedPrice:    10000,
			EstimatedPortions: 2,
			Status:            "available",
		},
		{
			ID:                uuid.New(),
			UserID:            sitiID,
			Title:             "Kentang Dieng",
			Category:          "Bahan Pokok",
			Quantity:          1000,
			Unit:              "gram",
			PhotoURLs:         `["https://images.unsplash.com/photo-1518977676601-b53f82aba655?w=500"]`,
			ConditionNote:     "Kondisi mulus, tidak bertunas, kering dan bersih.",
			StorageMethod:     "Tempat kering sejuk",
			PickupDeadline:    &tomorrow12,
			AreaName:          "Dinoyo, Malang",
			MeetingPoint:      "Teras rumah dekat Masjid Dinoyo",
			Latitude:          -7.9500,
			Longitude:         112.6050,
			WhatsAppNumber:    "+6285712345678",
			EstimatedPrice:    15000,
			EstimatedPortions: 3,
			Status:            "available",
		},
	}

	for _, d := range donations {
		var count int64
		db.Model(&entity.Donation{}).Where("title = ?", d.Title).Count(&count)
		if count == 0 {
			db.Create(&d)
		}
	}

	breakdown1, _ := sonic.Marshal([]map[string]interface{}{
		{"name": "Nasi Putih", "quantity": 2, "unit": "porsi", "estimated_price": 5000},
		{"name": "Telur Ayam", "quantity": 2, "unit": "butir", "estimated_price": 4000},
		{"name": "Wortel", "quantity": 100, "unit": "gram", "estimated_price": 3000},
	})

	breakdown2, _ := sonic.Marshal([]map[string]interface{}{
		{"name": "Labu Siam & Buncis", "quantity": 700, "unit": "gram", "estimated_price": 14000},
	})

	breakdown3, _ := sonic.Marshal([]map[string]interface{}{
		{"name": "Sup Telur Puyuh & Wortel", "quantity": 3, "unit": "porsi", "estimated_price": 18000},
	})

	breakdown4, _ := sonic.Marshal([]map[string]interface{}{
		{"name": "Wortel Segar", "quantity": 500, "unit": "gram", "estimated_price": 10000},
	})

	activities := []entity.ImpactActivity{
		{
			ID:             uuid.New(),
			UserID:         affanID,
			ActivityType:   "cooking",
			Title:          "Nasi Goreng Wortel Telur",
			PortionsSaved:  2,
			MoneySaved:     12000,
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
			MoneySaved:     14000,
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
			MoneySaved:     18000,
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
			MoneySaved:     10000,
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
