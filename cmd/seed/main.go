// cmd/seed/main.go
//
// Creates MercadoPago preapproval plans for each active local plan that does not
// yet have an mp_preapproval_plan_id, then persists the returned MP plan ID and
// init_point back to the database.
//
// Usage:
//
//	go run ./cmd/seed/main.go
//	make seed-mp-plans
//
// Run once per environment (dev / staging / prod) with the correct credentials.
// Idempotent: plans that already have an mp_preapproval_plan_id are skipped.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	sdkcfg        "github.com/mercadopago/sdk-go/pkg/config"
	sdkplan       "github.com/mercadopago/sdk-go/pkg/preapprovalplan"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	appconfig "github.com/jnieto01/payments-ms-01/internal/config"
	"github.com/jnieto01/payments-ms-01/internal/domain/entity"
)

const configPath = "./config/"

func main() {
	cfg := appconfig.Load(configPath)

	// ── DB connection ──────────────────────────────────────────────────────────
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.MySQL.User, cfg.MySQL.Password,
		cfg.MySQL.Host, cfg.MySQL.Port,
		cfg.MySQL.DBName,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	// ── MP SDK client ──────────────────────────────────────────────────────────
	sdkConfig, err := sdkcfg.New(cfg.MercadoPago.AccessToken)
	if err != nil {
		log.Fatalf("MP SDK init: %v", err)
	}
	planClient := sdkplan.NewClient(sdkConfig)

	// ── Fetch plans that still need seeding ───────────────────────────────────
	var plans []entity.Plan
	if err := db.Where("is_active = 1 AND (mp_preapproval_plan_id IS NULL OR mp_preapproval_plan_id = '')").
		Find(&plans).Error; err != nil {
		log.Fatalf("fetch plans: %v", err)
	}

	if len(plans) == 0 {
		log.Println("All active plans already have an mp_preapproval_plan_id. Nothing to do.")
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	successCount := 0
	for _, plan := range plans {
		log.Printf("→ Creating MP preapproval plan for [%s] %s (%.2f %s/month)...",
			plan.PlanKey, plan.Name, plan.Price, plan.Currency)

		req := sdkplan.Request{
			Reason:  fmt.Sprintf("NVF Sports — Plan %s", plan.Name),
			BackURL: cfg.MercadoPago.BackURL,
			AutoRecurring: &sdkplan.AutoRecurringRequest{
				Frequency:         1,
				FrequencyType:     "months",
				TransactionAmount: plan.Price,
				CurrencyID:        plan.Currency,
			},
		}

		resp, err := planClient.Create(ctx, req)
		if err != nil {
			log.Printf("  ✗ FAILED for plan %s: %v", plan.PlanKey, err)
			continue
		}

		if err := db.Model(&entity.Plan{}).
			Where("id = ?", plan.ID).
			Updates(map[string]any{
				"mp_preapproval_plan_id":  resp.ID,
				"mp_preapproval_plan_url": resp.InitPoint,
			}).Error; err != nil {
			log.Printf("  ✗ DB update failed for plan %s (mp_id=%s): %v", plan.PlanKey, resp.ID, err)
			continue
		}

		log.Printf("  ✓ plan_key=%s  mp_preapproval_plan_id=%s", plan.PlanKey, resp.ID)
		successCount++
	}

	log.Printf("\nDone: %d/%d plans seeded successfully.", successCount, len(plans))
	if successCount < len(plans) {
		os.Exit(1)
	}
}
