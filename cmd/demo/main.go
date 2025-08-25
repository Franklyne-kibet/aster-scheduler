package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Franklyne-kibet/aster-scheduler/internal/config"
	"github.com/Franklyne-kibet/aster-scheduler/internal/db"
	"github.com/Franklyne-kibet/aster-scheduler/internal/db/store"
	"github.com/Franklyne-kibet/aster-scheduler/internal/types"
	"github.com/google/uuid"
)

func main() {
	fmt.Println("🚀 Aster Demo - Testing CRUD Operations")
	fmt.Println("=====================================")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	database, err := db.NewConnection(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	fmt.Println("✅ Connected to database")

	// Create job store
	jobStore := store.NewJobStore(database.Pool())

	// Demo all CRUD operations
	if err := runCRUDDemo(ctx, jobStore); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Println("\n🎉 All operations completed successfully!")
}

func runCRUDDemo(ctx context.Context, store *store.JobStore) error {
	// 1. CREATE - Create a new job
	fmt.Println("\n1️⃣ Creating a new job...")
	
	timeout := 5 * time.Minute // 5 minute timeout
	job := &types.Job{
		ID:          uuid.New(),
		Name:        fmt.Sprintf("demo_job_%d", time.Now().Unix()),
		Description: "A demo job that says hello",
		CronExpr:    "0 */10 * * *", // Every 10 minutes
		Command:     "echo",
		Args:        []string{"Hello", "from", "Aster!"},
		Env: map[string]string{
			"ENVIRONMENT": "demo",
			"VERSION":     "1.0",
		},
		Status:     types.JobStatusActive,
		MaxRetries: 3,
		Timeout:    &timeout,
	}

	if err := store.CreateJob(ctx, job); err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	fmt.Printf("   ✅ Created job: %s (ID: %s)\n", job.Name, job.ID)

	// 2. READ - Get the job back
	fmt.Println("\n2️⃣ Reading the job back...")
	
	retrieved, err := store.GetJob(ctx, job.ID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	fmt.Printf("   ✅ Retrieved job: %s\n", retrieved.Name)
	fmt.Printf("   📋 Description: %s\n", retrieved.Description)
	fmt.Printf("   ⏰ Cron: %s\n", retrieved.CronExpr)
	fmt.Printf("   💻 Command: %s %v\n", retrieved.Command, retrieved.Args)
	fmt.Printf("   🌍 Environment: %v\n", retrieved.Env)
	fmt.Printf("   📊 Status: %s\n", retrieved.Status)
	fmt.Printf("   🔄 Max Retries: %d\n", retrieved.MaxRetries)
	if retrieved.Timeout != nil {
		fmt.Printf("   ⏱️  Timeout: %s\n", *retrieved.Timeout)
	}

	// 3. UPDATE - Modify the job
	fmt.Println("\n3️⃣ Updating the job...")
	
	retrieved.Description = "Updated demo job"
	retrieved.Command = "curl"
	retrieved.Args = []string{"-s", "https://httpbin.org/json"}
	retrieved.Env["VERSION"] = "1.1"
	retrieved.Status = types.JobStatusInactive

	if err := store.UpdateJob(ctx, retrieved); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	fmt.Println("   ✅ Job updated successfully")

	// Verify update
	updated, err := store.GetJob(ctx, job.ID)
	if err != nil {
		return fmt.Errorf("failed to get updated job: %w", err)
	}

	fmt.Printf("   📋 New Description: %s\n", updated.Description)
	fmt.Printf("   💻 New Command: %s %v\n", updated.Command, updated.Args)
	fmt.Printf("   📊 New Status: %s\n", updated.Status)
	fmt.Printf("   🌍 Updated Env: %v\n", updated.Env)

	// 4. LIST - Show all jobs
	fmt.Println("\n4️⃣ Listing all jobs...")
	
	jobs, err := store.ListJobs(ctx, 5, 0) // Get first 5 jobs
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	fmt.Printf("   📝 Found %d jobs:\n", len(jobs))
	for i, j := range jobs {
		fmt.Printf("   %d. %s (%s) - %s\n", 
			i+1, j.Name, j.Status, j.Command)
	}

	// 5. DELETE - Remove the job
	fmt.Println("\n5️⃣ Deleting the job...")
	
	if err := store.DeleteJob(ctx, job.ID); err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	fmt.Println("   ✅ Job deleted successfully")

	// Verify deletion
	_, err = store.GetJob(ctx, job.ID)
	if err == nil {
		return fmt.Errorf("job should have been deleted but still exists")
	}

	fmt.Println("   ✅ Verified job no longer exists")

	return nil
}