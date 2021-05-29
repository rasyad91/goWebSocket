package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"vigilate/internal/helpers"
	"vigilate/internal/models"

	"github.com/go-chi/chi"
)

const (
	HTTP           = 1
	HTTPS          = 2
	SSLCertificate = 3
)

type jsonResp struct {
	OK            bool      `json:"ok"`
	Message       string    `json:"message"`
	ServiceID     int       `json:"service_id"`
	HostServiceID int       `json:"host_service_id"`
	HostID        int       `json:"host_id"`
	OldStatus     string    `json:"old_status"`
	NewStatus     string    `json:"new_status"`
	LastCheck     time.Time `json:"last_check"`
}

func (repo *DBRepo) TestCheck(w http.ResponseWriter, r *http.Request) {
	hostServiceID, err := strconv.Atoi(chi.URLParam(r, "id"))
	ok := true
	if err != nil {
		log.Println(err)
		ok = false
	}
	oldStatus := chi.URLParam(r, "oldStatus")

	// use id to get host service
	hs, err := repo.DB.GetHostServiceByID(hostServiceID)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		helpers.ServerError(w, r, err)
		ok = false
	}

	// get host?
	h, err := repo.DB.GetHostByID(hs.HostID)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		helpers.ServerError(w, r, err)
		ok = false
	}
	// test the service
	msg, newStatus := repo.testServiceForHost(h, hs)
	log.Println(msg, newStatus)

	// update the host service in the db with new status if changed and last check
	hs.Status = newStatus
	hs.UpdatedAt = time.Now()
	hs.LastCheck = time.Now()
	if err := repo.DB.UpdateHostService(hs); err != nil {
		log.Println(err)
		ok = false
	}
	// broadcast service status changed event

	resp := jsonResp{}
	// create json
	if ok {
		resp = jsonResp{
			OK:            true,
			Message:       msg,
			ServiceID:     hs.ServiceID,
			HostServiceID: hs.ID,
			HostID:        hs.HostID,
			OldStatus:     oldStatus,
			NewStatus:     newStatus,
			LastCheck:     time.Now(),
		}
	} else {
		resp.OK = ok
		resp.Message = "Something went wrong"
	}

	// send json to client
	out, _ := json.MarshalIndent(resp, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (repo *DBRepo) testServiceForHost(h models.Host, hs models.HostService) (msg, newStatus string) {

	switch hs.ServiceID {
	case HTTP:
		msg, newStatus = testHTTPForHost(h.URL)
	case HTTPS:
	case SSLCertificate:
	}

	return
}

func testHTTPForHost(url string) (string, string) {
	if strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}

	url = strings.Replace(url, "https://", "http://", -1)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("%s - %s ", url, "error connecting"), "problem"
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("%s - %s ", url, resp.Status), "problem"
	}

	return fmt.Sprintf("%s - %s ", url, resp.Status), "healthy"

}
