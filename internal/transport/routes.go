package transport

import (
	"net/http"

	"github.com/Robotech-Org/driver-management/internal/auth"
	"github.com/Robotech-Org/driver-management/internal/wiring"
	"github.com/Robotech-Org/driver-management/utils/responses"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func NewRoutes(di *wiring.DependencyContainer) *chi.Mux {
	router := chi.NewRouter()

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "https://erp-system-eta-nine.vercel.app", "https://mahiseid.et", "https://www.mahiseid.et"},
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

	router.Group(func(r chi.Router) {
		r.Post("/login", di.AuthHandler.Login)
		r.Post("/refresh", di.AuthHandler.RefreshTokenHandler)
	})

	// Authenticated routes
	router.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Get("/me", di.AuthHandler.ProfileHandler)
		r.Patch("/me", di.AuthHandler.UpdateProfileHandler)
		r.Post("/logout", di.AuthHandler.LogoutHandler)
		r.Post("/account/set_password", di.AuthHandler.ResetPasswordHandler)
	})

	router.Route("/admin", func(adminRoute chi.Router) {
		adminRoute.Use(auth.AuthMiddleware)
		adminRoute.Route("/user", func(r chi.Router) {
			r.Post("/", di.AuthHandler.CreateUserHandler)
			r.Get("/list", di.AuthHandler.GetUsersHandler)
		})
		adminRoute.Route("/role", func(r chi.Router) {
			r.Get("/", di.AuthHandler.GetRolesHandler)
			r.Post("/", di.AuthHandler.CreateRoleHandler)
			r.Patch("/{id}", di.AuthHandler.UpdateRoleHandler)
			r.Delete("/{id}", di.AuthHandler.DeleteRoleHandler)
		})
		adminRoute.Route("/driver", func(r chi.Router) {
			r.Get("/", di.DriverHandler.ListDrivers)
			r.Post("/", di.DriverHandler.CreateDriver)
			r.Patch("/{id}", di.DriverHandler.UpdateDriver)
			r.Delete("/{id}", di.DriverHandler.DeleteDriver)
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
		})
		adminRoute.Route("/expense", func(r chi.Router) {
			r.Get("/", di.DriverHandler.GetExpenseTypes)
			r.Post("/", di.DriverHandler.CreateExpense)
		})
		adminRoute.Route("/stats", func(r chi.Router) {
			r.Get("/", di.DriverHandler.GetStats)
		})
	})

	router.Route("/finance", func(financeRoute chi.Router) {
		// financeRoute.Use(auth.AuthMiddleware)
		// What I want the finance manager to do is
		// 1. Get Trips with their expenses
		financeRoute.Route("/trip", func(r chi.Router) {
			r.Get("/", di.DriverHandler.ListTrips)
			r.Get("/{id}", di.DriverHandler.TripDetail)
			r.Post("/{id}/upload-receipt", di.DriverHandler.UploadReceipt)
		})
		// 2. Store payments made
	})

	router.Route("/user", func(userRoute chi.Router) {
		userRoute.Route("/password", func(r chi.Router) {
			r.Post("/", di.AuthHandler.ResetPasswordHandler)
		})
	})

	return router
}
