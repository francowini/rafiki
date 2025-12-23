// Package main provides a seed data tool for local development testing.
package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"

	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/google/uuid"
)

const (
	AdminEmail    = "admin@rafiki.lat"
	AdminPassword = "admin123"
	AdminName     = "Francisco Admin"
)

// ObjectiveScenario defines test scenarios for result-type objectives.
type ObjectiveScenario struct {
	Title         string
	BeginMetric   int
	TargetMetric  int
	CurrentMetric int
	Description   string
}

// Test scenarios covering various begin/target/current combinations
var objectiveScenarios = []ObjectiveScenario{
	{
		Title:         "Leer 100 libros este year",
		BeginMetric:   0,
		TargetMetric:  100,
		CurrentMetric: 50,
		Description:   "Increase goal - 50% complete",
	},
	{
		Title:         "Reducir peso (50 a 30)",
		BeginMetric:   50,
		TargetMetric:  30,
		CurrentMetric: 50, // Just started, at begin value
		Description:   "Decrease goal - 0% complete",
	},
	{
		Title:         "Reducir cafeina (80 a 70)",
		BeginMetric:   80,
		TargetMetric:  70,
		CurrentMetric: 75, // Halfway
		Description:   "Decrease goal - 50% complete",
	},
	{
		Title:         "Reducir tiempo pantalla (100 a 80)",
		BeginMetric:   100,
		TargetMetric:  80,
		CurrentMetric: 80, // Goal reached
		Description:   "Decrease goal - 100% complete",
	},
	{
		Title:         "Reducir azucar (50 a 30) - superado",
		BeginMetric:   50,
		TargetMetric:  30,
		CurrentMetric: 20, // Overshot target
		Description:   "Decrease goal - overshot target",
	},
	{
		Title:         "Ahorrar 10000 pesos",
		BeginMetric:   0,
		TargetMetric:  10000,
		CurrentMetric: 2500,
		Description:   "Increase goal - 25% complete",
	},
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: seeddata <db_url> <encryption_key_hex>")
		os.Exit(1)
	}

	dbURL := os.Args[1]
	encryptionKeyHex := os.Args[2]

	// Parse encryption key
	encryptionKey, err := hex.DecodeString(encryptionKeyHex)
	if err != nil {
		log.Fatalf("Invalid encryption key (must be hex): %v", err)
	}

	// Create encryptor
	encryptor, err := encrypt.NewAESEncryptor(encryptionKey)
	if err != nil {
		log.Fatalf("Failed to create encryptor: %v", err)
	}

	// Connect to database
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	ctx := context.Background()

	// Begin transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	log.Println("Deleting existing test data...")
	if err := deleteExistingData(ctx, tx); err != nil {
		log.Fatalf("Failed to delete existing data: %v", err)
	}

	log.Println("Creating admin user...")
	userID, err := createAdminUser(ctx, tx)
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}
	log.Printf("  User ID: %s", userID)

	log.Println("Creating values...")
	valueIDs, err := createValues(ctx, tx, userID, encryptor)
	if err != nil {
		log.Fatalf("Failed to create values: %v", err)
	}
	log.Printf("  Created %d values", len(valueIDs))

	log.Println("Creating life visions...")
	lifeVisionIDs, err := createLifeVisions(ctx, tx, userID, valueIDs, encryptor)
	if err != nil {
		log.Fatalf("Failed to create life visions: %v", err)
	}
	log.Printf("  Created %d life visions", len(lifeVisionIDs))

	log.Println("Creating objectives (result-type with various scenarios)...")
	objectiveIDs, err := createObjectives(ctx, tx, userID, lifeVisionIDs, encryptor)
	if err != nil {
		log.Fatalf("Failed to create objectives: %v", err)
	}
	log.Printf("  Created %d objectives", len(objectiveIDs))

	log.Println("Creating frequency objective with 90 days of records...")
	freqObjectiveID, err := createFrequencyObjective(ctx, tx, userID, lifeVisionIDs[0], encryptor)
	if err != nil {
		log.Fatalf("Failed to create frequency objective: %v", err)
	}
	log.Printf("  Frequency objective ID: %s", freqObjectiveID)

	log.Println("Creating objective records (90 days history)...")
	if err := createObjectiveRecords(ctx, tx, userID, freqObjectiveID); err != nil {
		log.Fatalf("Failed to create objective records: %v", err)
	}
	log.Println("  Created 90 days of records")

	log.Println("Creating tasks...")
	if err := createTasks(ctx, tx, userID, objectiveIDs, encryptor); err != nil {
		log.Fatalf("Failed to create tasks: %v", err)
	}
	log.Println("  Created sample tasks")

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Println("")
	log.Println("=== Seed Data Summary ===")
	log.Printf("User: %s", AdminEmail)
	log.Printf("Values: %d", len(valueIDs))
	log.Printf("Life Visions: %d", len(lifeVisionIDs))
	log.Printf("Result Objectives: %d", len(objectiveIDs))
	log.Println("Frequency Objectives: 1")
	log.Println("Records: 90 days")
	log.Println("")
}

func deleteExistingData(ctx context.Context, tx *sql.Tx) error {
	// Delete in reverse dependency order
	tables := []string{
		"objective_records",
		"tasks",
		"objectives",
		"life_visions",
		"values",
		"notification_messages",
		"moments",
		"thinks",
		"users",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE 1=1", table)
		if _, err := tx.ExecContext(ctx, query); err != nil {
			// Ignore errors for tables that might not exist
			log.Printf("  Warning: could not delete from %s: %v", table, err)
		}
	}

	return nil
}

func createAdminUser(ctx context.Context, tx *sql.Tx) (uuid.UUID, error) {
	userID := uuid.New()

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash password: %w", err)
	}

	query := `
		INSERT INTO users (
			user_id, name, email, roles, password_hash,
			department, enabled, date_created, date_updated
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	now := time.Now()
	_, err = tx.ExecContext(ctx, query,
		userID,
		AdminName,
		AdminEmail,
		"{ADMIN}",
		string(hash),
		nil,
		true,
		now,
		now,
	)

	return userID, err
}

func createValues(ctx context.Context, tx *sql.Tx, userID uuid.UUID, enc encrypt.Encryptor) ([]uuid.UUID, error) {
	values := []struct {
		content string
		facet   string
		order   int
	}{
		{"Vivir con integridad y honestidad", "personal_growth", 1},
		{"Cultivar relaciones autenticas y profundas", "relationships", 2},
		{"Contribuir al bienestar de los demas", "community", 3},
		{"Buscar el crecimiento continuo", "career", 4},
		{"Mantener la salud fisica y mental", "health", 5},
	}

	var ids []uuid.UUID
	now := time.Now()

	for _, v := range values {
		id := uuid.New()

		// Encrypt content
		encrypted, err := enc.Encrypt(v.content)
		if err != nil {
			return nil, fmt.Errorf("encrypt value content: %w", err)
		}

		query := `
			INSERT INTO values (
				value_id, user_id, content, facet, display_order,
				status, archived_at, date_created, date_updated
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

		_, err = tx.ExecContext(ctx, query,
			id, userID, encrypted, v.facet, v.order,
			"active", nil, now, now,
		)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func createLifeVisions(ctx context.Context, tx *sql.Tx, userID uuid.UUID, valueIDs []uuid.UUID, enc encrypt.Encryptor) ([]uuid.UUID, error) {
	visions := []string{
		"Ser una persona que inspira confianza y respeto en todos mis circulos",
		"Tener relaciones profundas basadas en la vulnerabilidad y el crecimiento mutuo",
		"Alcanzar la libertad financiera para poder dedicarme a lo que me apasiona",
	}

	var ids []uuid.UUID
	now := time.Now()

	for i, content := range visions {
		id := uuid.New()

		// Encrypt content
		encrypted, err := enc.Encrypt(content)
		if err != nil {
			return nil, fmt.Errorf("encrypt life vision content: %w", err)
		}

		// Link to value (cycle through available values)
		valueID := valueIDs[i%len(valueIDs)]

		query := `
			INSERT INTO life_visions (
				life_vision_id, user_id, value_id, content,
				status, archived_at, date_created, date_updated
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

		_, err = tx.ExecContext(ctx, query,
			id, userID, valueID, encrypted,
			"active", nil, now, now,
		)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func createObjectives(ctx context.Context, tx *sql.Tx, userID uuid.UUID, lifeVisionIDs []uuid.UUID, enc encrypt.Encryptor) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	now := time.Now()

	for i, scenario := range objectiveScenarios {
		id := uuid.New()

		// Encrypt title
		encrypted, err := enc.Encrypt(scenario.Title)
		if err != nil {
			return nil, fmt.Errorf("encrypt objective title: %w", err)
		}

		// Link to life vision (cycle through available)
		lifeVisionID := lifeVisionIDs[i%len(lifeVisionIDs)]

		query := `
			INSERT INTO objectives (
				objective_id, user_id, life_vision_id, title,
				tracking_type, status, target_metric, current_metric, begin_metric,
				frequency_type, frequency_count, compliance_target_pct,
				archived_at, date_created, date_updated
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

		_, err = tx.ExecContext(ctx, query,
			id, userID, lifeVisionID, encrypted,
			"result", "active", scenario.TargetMetric, scenario.CurrentMetric, scenario.BeginMetric,
			nil, nil, nil,
			nil, now, now,
		)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
		log.Printf("    - %s (begin=%d, target=%d, current=%d)", scenario.Title, scenario.BeginMetric, scenario.TargetMetric, scenario.CurrentMetric)
	}

	return ids, nil
}

func createFrequencyObjective(ctx context.Context, tx *sql.Tx, userID uuid.UUID, lifeVisionID uuid.UUID, enc encrypt.Encryptor) (uuid.UUID, error) {
	id := uuid.New()
	now := time.Now()

	title := "Meditar 10 minutos diarios"
	encrypted, err := enc.Encrypt(title)
	if err != nil {
		return uuid.Nil, fmt.Errorf("encrypt title: %w", err)
	}

	query := `
		INSERT INTO objectives (
			objective_id, user_id, life_vision_id, title,
			tracking_type, status, target_metric, current_metric, begin_metric,
			frequency_type, frequency_count, compliance_target_pct,
			archived_at, date_created, date_updated
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err = tx.ExecContext(ctx, query,
		id, userID, lifeVisionID, encrypted,
		"frequency", "active", nil, nil, nil,
		"daily", nil, 80, // 80% compliance target
		nil, now, now,
	)

	return id, err
}

func createObjectiveRecords(ctx context.Context, tx *sql.Tx, userID uuid.UUID, objectiveID uuid.UUID) error {
	// Generate 90 days of records with ~80% compliance
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -90)

	now := time.Now()
	statuses := []string{"completed", "completed", "completed", "completed", "skipped"} // 80% success

	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		recordID := uuid.New()

		// Select status (80% completed, 20% skipped)
		status := statuses[rand.Intn(len(statuses))]

		query := `
			INSERT INTO objective_records (
				objective_record_id, objective_id, user_id,
				record_date, status, notes, date_created
			) VALUES ($1, $2, $3, $4, $5, $6, $7)`

		_, err := tx.ExecContext(ctx, query,
			recordID, objectiveID, userID,
			d.Format("2006-01-02"), status, nil, now,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func createTasks(ctx context.Context, tx *sql.Tx, userID uuid.UUID, objectiveIDs []uuid.UUID, enc encrypt.Encryptor) error {
	tasks := []struct {
		title        string
		description  string
		contribution *int
		status       string
		objectiveIdx *int // nil = inbox task
	}{
		// Inbox tasks (no objective)
		{"Revisar y responder emails", "Procesar inbox y responder emails urgentes", nil, "pending", nil},
		{"Planificar la semana", "Revisar calendario y bloquear tiempo", nil, "pending", nil},
		{"Llamar al dentista", "Programar cita semestral", nil, "completed", nil},

		// Tasks for first objective (Read 100 Books)
		{"Leer 30 paginas del libro actual", "Habito de lectura diaria", intPtr(5), "completed", intPtr(0)},
		{"Comprar 3 libros de la lista", "Preparar proximas lecturas", intPtr(3), "pending", intPtr(0)},

		// Tasks for second objective (Weight reduction)
		{"Caminata de 30 minutos", "Rutina de ejercicio matutino", intPtr(2), "completed", intPtr(1)},
		{"Preparar comidas de la semana", "Meal prep saludable", intPtr(5), "pending", intPtr(1)},
	}

	now := time.Now()

	for _, task := range tasks {
		id := uuid.New()

		// Encrypt title and description
		encryptedTitle, err := enc.Encrypt(task.title)
		if err != nil {
			return fmt.Errorf("encrypt task title: %w", err)
		}

		var encryptedDesc sql.NullString
		if task.description != "" {
			encrypted, err := enc.Encrypt(task.description)
			if err != nil {
				return fmt.Errorf("encrypt task description: %w", err)
			}
			encryptedDesc = sql.NullString{String: encrypted, Valid: true}
		}

		var objectiveID *uuid.UUID
		if task.objectiveIdx != nil && *task.objectiveIdx < len(objectiveIDs) {
			oid := objectiveIDs[*task.objectiveIdx]
			objectiveID = &oid
		}

		var completedAt sql.NullTime
		if task.status == "completed" {
			completedAt = sql.NullTime{Time: now.Add(-24 * time.Hour), Valid: true}
		}

		query := `
			INSERT INTO tasks (
				task_id, user_id, objective_id, title, description,
				contribution, status, completed_at, date_created, date_updated
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

		_, err = tx.ExecContext(ctx, query,
			id, userID, objectiveID, encryptedTitle, encryptedDesc,
			task.contribution, task.status, completedAt, now, now,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func intPtr(i int) *int {
	return &i
}
