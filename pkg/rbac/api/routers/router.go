package routers

import (
	"net/http"

	"github.com/gorilla/mux"

	handlers "github.com/omc-college/management-system/pkg/rbac/api/handlers"
	postgres "github.com/omc-college/management-system/pkg/rbac/repository/postgres"
)

// NewCrudRouter Inits RBAC CRUD service router
func NewCrudRouter(repository *postgres.RolesRepository) *mux.Router {
	// Init handlers DB wrap
	var handlersDB handlers.RolesRepositoryWrap
	handlersDB.RolesRepository = repository

	router := mux.NewRouter()

	router.HandleFunc("/roles", handlersDB.GetAllRoles).Methods(http.MethodGet)
	router.HandleFunc("/roles", handlersDB.CreateRole).Methods(http.MethodPost)
	router.HandleFunc("/roles/{id}", handlersDB.GetRole).Methods(http.MethodGet)
	router.HandleFunc("/roles/{id}", handlersDB.UpdateRole).Methods(http.MethodPut)
	router.HandleFunc("/roles/{id}", handlersDB.DeleteRole).Methods(http.MethodDelete)
	router.HandleFunc("/roletmpl", handlersDB.GetRoleTemplate).Methods(http.MethodGet)

	return router
}
