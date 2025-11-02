package transport

import (
	"net/http"

	"github.com/Robotech-Org/driver-management/internal/wiring"
	"github.com/Robotech-Org/driver-management/utils/responses"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func NewRoutes(di *wiring.DependencyContainer) *chi.Mux {
	router := chi.NewRouter()

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
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
		r.Post("/logout", di.AuthHandler.LogoutHandler)
	})

	router.Route("/admin", func(adminRoute chi.Router) {
		adminRoute.Route("/user", func(r chi.Router) {
			r.Post("/", di.AuthHandler.CreateUserHandler)
			r.Post("/set-password", di.AuthHandler.ResetPasswordHandler)
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
			r.Patch("/{id}", di.DriverHandler.ChangeCarDriver)
		})
		adminRoute.Route("/trip", func(r chi.Router) {
			r.Get("/", di.DriverHandler.ListTrips)
			r.Post("/", di.DriverHandler.CreateTrip)
			r.Patch("/{id}", di.DriverHandler.UpdateTrip)
			r.Delete("/{id}", di.DriverHandler.DeleteTrip)
		})
	})

	router.Route("/user", func(userRoute chi.Router) {
		userRoute.Route("/password", func(r chi.Router) {
			r.Post("/", di.AuthHandler.ResetPasswordHandler)
		})
	})

	return router
}
