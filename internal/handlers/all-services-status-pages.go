package handlers

import (
	"log"
	"net/http"
	"vigilate/internal/helpers"

	"github.com/CloudyKit/jet/v6"
)

// AllHealthyServices lists all healthy services
func (repo *DBRepo) AllHealthyServices(w http.ResponseWriter, r *http.Request) {
	hostServices, err := repo.DB.GetServicesByStatus("healthy")
	if err != nil {
		helpers.ServerError(w, r, err)
		log.Println(err)
		return
	}
	vars := make(jet.VarMap)
	vars.Set("hostServices", hostServices)
	if err := helpers.RenderPage(w, r, "healthy", vars, nil); err != nil {
		printTemplateError(w, err)
	}
}

// AllWarningServices lists all warning services
func (repo *DBRepo) AllWarningServices(w http.ResponseWriter, r *http.Request) {
	hostServices, err := repo.DB.GetServicesByStatus("warning")
	if err != nil {
		helpers.ServerError(w, r, err)
		log.Println(err)
		return
	}
	vars := make(jet.VarMap)
	vars.Set("hostServices", hostServices)
	if err := helpers.RenderPage(w, r, "warning", vars, nil); err != nil {
		printTemplateError(w, err)
	}
}

// AllProblemServices lists all problem services
func (repo *DBRepo) AllProblemServices(w http.ResponseWriter, r *http.Request) {
	hostServices, err := repo.DB.GetServicesByStatus("problem")
	if err != nil {
		helpers.ServerError(w, r, err)
		log.Println(err)
		return
	}
	vars := make(jet.VarMap)
	vars.Set("hostServices", hostServices)
	if err := helpers.RenderPage(w, r, "problems", vars, nil); err != nil {
		printTemplateError(w, err)
	}
}

// AllPendingServices lists all pending services
func (repo *DBRepo) AllPendingServices(w http.ResponseWriter, r *http.Request) {
	hostServices, err := repo.DB.GetServicesByStatus("pending")
	if err != nil {
		helpers.ServerError(w, r, err)
		log.Println(err)
		return
	}
	vars := make(jet.VarMap)
	vars.Set("hostServices", hostServices)
	if err := helpers.RenderPage(w, r, "pending", vars, nil); err != nil {
		printTemplateError(w, err)
	}
}
