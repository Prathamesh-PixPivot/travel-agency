// cmd/server/main.go
package main

import (
	"log"
	"net/http"
	"time"

	"travel-agency/internal/auth"
	"travel-agency/internal/config"
	"travel-agency/internal/db"
	"travel-agency/internal/handlers"
	"travel-agency/internal/jobs"
	"travel-agency/internal/models"
	"travel-agency/internal/notifications"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize DB
	database, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Ensure pgcrypto extension for gen_random_uuid()
	if err := database.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto;").Error; err != nil {
		log.Fatalf("Failed to enable pgcrypto extension: %v", err)
	}

	// Auto‑migrate all models except Invoice (created only if missing)
	toMigrate := []interface{}{
		&models.Tenant{},
		&models.User{},
		&models.Lead{},
		&models.Itinerary{},
		&models.ItineraryItem{},
		&models.Booking{},
		&models.Vendor{},
		&models.Payment{},
		&models.Task{},
		&models.Ticket{},
		&models.AuditLog{},
	}
	if err := database.AutoMigrate(toMigrate...); err != nil {
		log.Fatalf("Auto migration failed: %v", err)
	}
	if !database.Migrator().HasTable(&models.Invoice{}) {
		if err := database.AutoMigrate(&models.Invoice{}); err != nil {
			log.Fatalf("Failed to create invoices table: %v", err)
		}
	}

	// Start background jobs
	jobs.StartCronJobs(database)

	// Handlers
	jwtSecret := cfg.JWTSecret
	authHandler := handlers.NewAuthHandler(database, jwtSecret)
	smtpSender := notifications.NewSMTPSender(
		cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom,
	)
	adminHandler := handlers.NewAdminHandler(database, smtpSender)

	r := chi.NewRouter()

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3001"}, // In production, lock this down to your front-end origin(s).
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Rate limit
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	// Public auth
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/refresh", authHandler.RefreshToken)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware(jwtSecret))

		// Admin: agent CRUD
		r.Route("/api/admin/agents", func(r chi.Router) {
			r.Post("/", adminHandler.CreateAgent)
			r.Get("/", adminHandler.ListAgents)
			r.Get("/{agentID}", adminHandler.GetAgent)
			r.Put("/{agentID}", adminHandler.UpdateAgent)
			r.Delete("/{agentID}", adminHandler.DeleteAgent)
		})

		// User self‑service
		r.Get("/api/user/profile", authHandler.GetProfile)
		r.Put("/api/user/profile", authHandler.UpdateProfile)
		r.Put("/api/user/reset-password", authHandler.ResetPassword)

		// Leads
		leadsHandler := handlers.NewLeadsHandler(database)
		r.Route("/api/leads", func(r chi.Router) {
			r.Post("/", leadsHandler.CreateLead)
			r.Get("/", leadsHandler.ListLeads)
			r.Get("/{leadID}", leadsHandler.GetLead)
			r.Put("/{leadID}", leadsHandler.UpdateLead)
		})

		// Itineraries
		itinHandler := handlers.NewItineraryHandler(database)
		r.Route("/api/itineraries", func(r chi.Router) {
			r.Post("/", itinHandler.CreateItinerary)
			r.Get("/", itinHandler.ListItineraries)
			r.Get("/{itineraryID}", itinHandler.GetItinerary)
			r.Put("/{itineraryID}", itinHandler.UpdateItinerary)
		})

		// Bookings
		bookingHandler := handlers.NewBookingHandler(database)
		r.Route("/api/bookings", func(r chi.Router) {
			r.Post("/", bookingHandler.CreateBooking)
			r.Get("/", bookingHandler.ListBookings)
			r.Get("/{bookingID}", bookingHandler.GetBooking)
			r.Put("/{bookingID}", bookingHandler.UpdateBooking)
		})

		// Vendors
		vendorHandler := handlers.NewVendorHandler(database)
		r.Route("/api/vendors", func(r chi.Router) {
			r.Post("/", vendorHandler.CreateVendor)
			r.Get("/", vendorHandler.ListVendors)
			r.Get("/{vendorID}", vendorHandler.GetVendor)
			r.Put("/{vendorID}", vendorHandler.UpdateVendor)
		})

		// Invoices
		invoiceHandler := handlers.NewInvoiceHandler(database)
		r.Route("/api/invoices", func(r chi.Router) {
			r.Post("/", invoiceHandler.CreateInvoice)
			r.Get("/", invoiceHandler.ListInvoices)
			r.Get("/{invoiceID}", invoiceHandler.GetInvoice)
			r.Put("/{invoiceID}", invoiceHandler.UpdateInvoice)
			r.Get("/{invoiceID}/pdf", invoiceHandler.DownloadInvoicePDF)
		})

		// Payments
		paymentHandler := handlers.NewPaymentHandler(database)
		r.Route("/api/payments", func(r chi.Router) {
			r.Post("/", paymentHandler.CreatePayment)
			r.Get("/", paymentHandler.ListPayments)
			r.Get("/{paymentID}", paymentHandler.GetPayment)
			r.Put("/{paymentID}", paymentHandler.UpdatePayment)
		})

		// Tasks
		taskHandler := handlers.NewTaskHandler(database)
		r.Route("/api/tasks", func(r chi.Router) {
			r.Post("/", taskHandler.CreateTask)
			r.Get("/", taskHandler.ListTasks)
			r.Get("/{taskID}", taskHandler.GetTask)
			r.Put("/{taskID}", taskHandler.UpdateTask)
			r.Delete("/{taskID}", taskHandler.DeleteTask)
		})

		// Tickets
		ticketHandler := handlers.NewTicketHandler(database)
		r.Route("/api/tickets", func(r chi.Router) {
			r.Post("/", ticketHandler.CreateTicket)
			r.Get("/", ticketHandler.ListTickets)
			r.Get("/{ticketID}", ticketHandler.GetTicket)
		})
	})

	// Health check
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("Travel-agency backend is up"))
	})

	addr := ":" + cfg.Port
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
