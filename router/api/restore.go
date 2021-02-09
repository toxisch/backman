package api

import (
	"fmt"
	"net/http"
	"net/url"

	echo "github.com/labstack/echo/v4"
	"github.com/swisscom/backman/log"
	"github.com/swisscom/backman/service/util"
)

// swagger:route POST /api/v1/restore/{service_type}/{service_name} restore restoreBackup
// Triggers a restore for given service.
//
// produces:
// - application/json
//
// schemes: http, https
//
// responses:
//   202:
func (h *Handler) RestoreBackup(c echo.Context) error {
	serviceType := c.Param("service_type")
	serviceName, err := url.QueryUnescape(c.Param("service_name"))
	if err != nil {
		log.Errorf("%v", err)
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("invalid service name: %v", err))
	}
	targetName, err := url.QueryUnescape(c.Param("target_name"))
	if err != nil {
		log.Errorf("%v", err)
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("invalid target service name: %v", err))
	}
	filename, err := url.QueryUnescape(c.Param("file"))
	if err != nil {
		log.Errorf("%v", err)
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("invalid filename: %v", err))
	}

	if !util.IsValidServiceType(serviceType) {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("unsupported service type: %s", serviceType))
	}

	cfService := h.Service.GetService(serviceType, serviceName)
	if len(cfService.Name) == 0 {
		err := fmt.Errorf("could not find service [%s] to restore", serviceName)
		log.Errorf("%v", err)
		return c.JSON(http.StatusNotFound, err.Error())
	}

	targetService := util.Service{}
	if len(targetName) > 0 {
		targetService = h.Service.GetService(serviceType, targetName)
		if len(targetService.Name) == 0 {
			err := fmt.Errorf("could not find target service [%s] to restore", targetName)
			log.Errorf("%v", err)
			return c.JSON(http.StatusNotFound, err.Error())
		}
	}

	go func() { // async
		if err := h.Service.Restore(cfService, targetService, filename); err != nil {
			log.Errorf("requested restore for service [%s] failed: %v", serviceName, err)
		}
	}()
	return c.JSON(http.StatusAccepted, nil)
}
