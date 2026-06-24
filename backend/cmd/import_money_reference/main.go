package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/service"
	"github.com/alexscott64/woulder/backend/internal/storage"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	db, err := database.New()
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	project, err := db.Money().GetProjectBySlug(ctx, "money-creek")
	if err != nil {
		log.Fatalf("load money-creek project: %v", err)
	}

	user, err := importUser(ctx, db)
	if err != nil {
		log.Fatalf("load import user: %v", err)
	}

	fixture := os.Getenv("MONEY_REFERENCE_FIXTURE")
	if fixture == "" {
		fixture = "internal/database/money/fixtures/money_creek_crag.json"
	}
	file, err := os.Open(fixture)
	if err != nil {
		log.Fatalf("open fixture %s: %v", fixture, err)
	}
	defer file.Close()

	gpxPath := os.Getenv("MONEY_REFERENCE_GPX")
	if gpxPath == "" {
		gpxPath = "internal/database/money/fixtures/onx-markups-06232026.gpx"
	}
	gpxFile, err := os.Open(gpxPath)
	if err != nil {
		log.Fatalf("open GPX %s: %v", gpxPath, err)
	}
	defer gpxFile.Close()

	svc := service.NewMoneyService(db.Money(), storage.NewLocalStorage(os.TempDir()), 25<<20)
	if err := svc.ImportReferenceCragWithGPX(ctx, project.ID, file, gpxFile, user); err != nil {
		log.Fatalf("import fixture: %v", err)
	}
	log.Printf("Imported Money Creek reference crag into project %s as %s", project.ID, user.Email)
}

func importUser(ctx context.Context, db *database.Database) (models.CurrentUser, error) {
	var u models.CurrentUser
	err := db.Conn().QueryRowContext(ctx, `
		SELECT id, email, display_name, role
		FROM woulder.users
		WHERE is_active = true AND role IN ('admin','developer')
		ORDER BY CASE WHEN role='admin' THEN 0 ELSE 1 END, created_at ASC
		LIMIT 1
	`).Scan(&u.ID, &u.Email, &u.DisplayName, &u.Role)
	return u, err
}
