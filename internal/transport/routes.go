package transport

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/Robotech-Org/driver-management/internal/auth"
	"github.com/Robotech-Org/driver-management/internal/wiring"
	"github.com/Robotech-Org/driver-management/utils/responses"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func findDocsDir() string {
	candidates := []string{
		"./docs",
		"../docs",
		"../../docs",
	}
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			return c
		}
	}
	if exe, err := os.Executable(); err == nil {
		exeDocs := filepath.Join(filepath.Dir(exe), "docs")
		if fi, err := os.Stat(exeDocs); err == nil && fi.IsDir() {
			return exeDocs
		}
	}
	return "./docs"
}

func NewRoutes(di *wiring.DependencyContainer) *chi.Mux {
	router := chi.NewRouter()

	cors := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:*",
			"https://localhost:*",
			"http://localhost:3000",
			"http://localhost:3001",
			"https://erp-system-eta-nine.vercel.app",
			"https://mahiseid.et",
			"https://www.mahiseid.et",
			"https://*.up.railway.app",
			"https://driver-management-client-staging.up.railway.app",
			"https://driver-management-client-production.up.railway.app",
			"https://dmc-seven-kappa.vercel.app",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	router.Use(cors.Handler)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		message := "healthy"
		responses.SuccessJSONResponse(message, nil, http.StatusOK, w)
	})

	docsDir := findDocsDir()
	serveSwaggerYaml := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		http.ServeFile(w, r, filepath.Join(docsDir, "swagger.yaml"))
	}

	router.Get("/swagger.yaml", serveSwaggerYaml)
	router.Get("/swagger/swagger.yaml", serveSwaggerYaml)

	// Serve the Swagger UI
	fs := http.FileServer(http.Dir(filepath.Join(docsDir, "swagger-ui")))
	router.Handle("/swagger/*", http.StripPrefix("/swagger/", fs))

	// Documentation convenience redirects
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusFound)
	})
	router.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	router.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	router.Get("/docs/*", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})

	// Traccar Client posts driver positions here; the device token is the credential.
	router.Get("/track", di.TrackingHandler.Ingest)
	router.Post("/track", di.TrackingHandler.Ingest)

	router.Group(func(r chi.Router) {
		r.Post("/login", di.AuthHandler.Login)
		r.Post("/refresh", di.AuthHandler.RefreshTokenHandler)
	})

	// Authenticated routes
	router.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Get("/me", di.AuthHandler.ProfileHandler)
		r.Patch("/me", di.AuthHandler.UpdateProfileHandler)
		r.Post("/me/password", di.AuthHandler.ChangePasswordHandler)
		r.Post("/logout", di.AuthHandler.LogoutHandler)
		r.Post("/account/set_password", di.AuthHandler.ResetPasswordHandler)
	})

	router.Route("/admin", func(adminRoute chi.Router) {
		adminRoute.Use(auth.AuthMiddleware)
		adminRoute.Route("/user", func(r chi.Router) {
			r.Use(auth.RequireRole("admin"))
			r.Post("/", di.AuthHandler.CreateUserHandler)
			r.Get("/list", di.AuthHandler.GetUsersHandler)
			r.Patch("/{id}", di.AuthHandler.UpdateUserHandler)
			r.Delete("/{id}", di.AuthHandler.DeleteUserHandler)
			r.Post("/reset-password", di.AuthHandler.AdminResetPasswordHandler)
		})
		adminRoute.Route("/role", func(r chi.Router) {
			r.Use(auth.RequireRole("admin"))
			r.Get("/", di.AuthHandler.GetRolesHandler)
			r.Post("/", di.AuthHandler.CreateRoleHandler)
			r.Patch("/{id}", di.AuthHandler.UpdateRoleHandler)
			r.Delete("/{id}", di.AuthHandler.DeleteRoleHandler)
		})
		adminRoute.Route("/driver", func(r chi.Router) {
			r.Get("/", di.DriverHandler.ListDrivers)
			r.Post("/", di.DriverHandler.CreateDriver)
			r.Get("/{id}", di.DriverHandler.GetDriverDetail)
			r.Patch("/{id}", di.DriverHandler.UpdateDriver)
			r.Delete("/{id}", di.DriverHandler.DeleteDriver)
			r.Post("/{id}/payroll", di.DriverHandler.AddPayroll)
			r.Get("/{id}/payroll", di.DriverHandler.GetDriverLedger)
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole("admin", "fleet_manager"))
				r.Get("/{id}/tracking", di.TrackingHandler.TrackingStatus)
				r.Post("/{id}/tracking", di.TrackingHandler.EnableTracking)
				r.Delete("/{id}/tracking", di.TrackingHandler.DisableTracking)
				r.Get("/{id}/locations", di.TrackingHandler.DriverLocations)
			})
		})
		adminRoute.Route("/cars", func(r chi.Router) {
			r.Get("/", di.DriverHandler.ListCars)
			r.Post("/", di.DriverHandler.CreateCar)
			r.Patch("/{id}", di.DriverHandler.UpdateCar)
			r.Delete("/{id}", di.DriverHandler.DeleteCar)
			r.Patch("/{id}/driver", di.DriverHandler.ChangeCarDriver)
		})
		adminRoute.Route("/trip", func(r chi.Router) {
			r.Get("/", di.DriverHandler.ListTrips)
			r.Post("/", di.DriverHandler.CreateTrip)
			r.Get("/{id}", di.DriverHandler.TripDetail)
			r.Patch("/{id}", di.DriverHandler.UpdateTrip)
			r.Delete("/{id}", di.DriverHandler.DeleteTrip)
			r.With(auth.RequireRole("admin", "fleet_manager")).Get("/{id}/locations", di.TrackingHandler.TripLocations)
		})
		adminRoute.With(auth.RequireRole("admin", "fleet_manager")).Get("/tracking/live", di.TrackingHandler.LivePositions)
		adminRoute.Route("/expense-type", func(r chi.Router) {
			r.Get("/", di.DriverHandler.GetExpenseTypes)
			r.Post("/", di.DriverHandler.CreateExpense)
			r.Patch("/{id}", di.DriverHandler.UpdateExpense)
			r.Delete("/{id}", di.DriverHandler.DeleteExpense)
		})
		adminRoute.Route("/stats", func(r chi.Router) {
			r.Get("/", di.DriverHandler.GetStats)
		})
		adminRoute.Route("/income", func(r chi.Router) {
			r.Post("/advance", di.DriverHandler.RegisterAdvancePayment)
			r.Patch("/advance/{id}", di.DriverHandler.UpdateAdvancePayment)
			r.Post("/settlement", di.DriverHandler.RegisterFinalPayment)
			r.Patch("/settlement/{id}", di.DriverHandler.UpdateSettlementPayment)
			r.Post("/maintenance", di.DriverHandler.RegisterMaintenance)
			r.Get("/summary", di.DriverHandler.GetFinancialSummary)
			r.Get("/trip/{tripId}", di.DriverHandler.GetTripProfitability)
		})
	})

	router.Route("/finance", func(financeRoute chi.Router) {
		financeRoute.Use(auth.AuthMiddleware)
		financeRoute.Route("/trip", func(r chi.Router) {
			r.Get("/", di.DriverHandler.ListTrips)
			r.Get("/{id}", di.DriverHandler.TripDetail)
			r.Post("/{id}", di.DriverHandler.CreateTripExpenseHandler)
		})
		financeRoute.Patch("/trip-expense/{id}", di.DriverHandler.UpdateTripExpense)
		financeRoute.Post("/driver-pay", di.DriverHandler.RegisterDriverPayment)
		financeRoute.Post("/trip-overhead", di.DriverHandler.RegisterTripOverhead)
		financeRoute.Post("/trip-return", di.DriverHandler.RegisterReturnPayment)
	})

	router.Route("/user", func(userRoute chi.Router) {
		userRoute.Route("/password", func(r chi.Router) {
			r.Post("/", di.AuthHandler.ResetPasswordHandler)
		})
	})

	return router
}
