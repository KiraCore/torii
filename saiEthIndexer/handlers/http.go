package handlers

import (
	"net/http"

	valid "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/saiset-co/saiEthIndexer/config"
	"github.com/saiset-co/saiEthIndexer/tasks"
	"go.uber.org/zap"
)

type HttpHandler struct {
	Logger      *zap.Logger
	TaskManager *tasks.TaskManager
}

type addContractsRequest struct {
	Contracts []config.Contract `json:"contracts" valid:",required"`
}

type deleteContractsRequest struct {
	Addreses []string `json:"addresses" valid:",required"`
}

// Validation of contracts struct
func (r *addContractsRequest) validate() error {
	_, err := valid.ValidateStruct(r)

	return err
}

type addContractResponse struct {
	Created bool `json:"is_added" example:"true"`
}
type deleteContractResponse struct {
	Created bool `json:"is_deleted" example:"true"`
}

func HandleHTTP(g *gin.RouterGroup, logger *zap.Logger, t *tasks.TaskManager) {
	handler := &HttpHandler{
		Logger:      logger,
		TaskManager: t,
	}
	{
		g.POST("/add_contract", handler.addContract)
		g.POST("/delete_contract", handler.deleteContracts)
	}
}

// @Summary     add contract
// @Description add contract
// @ID          add contract
// @Tags  	    Contract
// @Accept      json
// @Produce     json
// @Success     200 {object} addContractResponse
// @Failure     500 {object} errInternalServer
// @Failure     400 {object} errBadRequest
// @Router      /add_contract [post]
func (h *HttpHandler) addContract(c *gin.Context) {
	dto := addContractsRequest{}
	err := c.ShouldBindJSON(&dto)
	if err != nil {
		h.Logger.Error("http  - add contract - bind", zap.Error(err))
		c.JSON(http.StatusBadRequest, errBadRequest)
	}

	for _, contract := range dto.Contracts {
		err = contract.Validate()
		if err != nil {
			h.Logger.Error("http  - add contract - validate", zap.Error(err))
			c.JSON(http.StatusBadRequest, errBadRequest)
			return
		}
	}
	err = h.TaskManager.AddContract(dto.Contracts)
	if err != nil {
		h.Logger.Error("http - add contract", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errInternalServer)
		return
	}

	c.JSON(http.StatusOK, &addContractResponse{Created: true})
}

// @Summary     delete contract
// @Description delete contract
// @ID          delete contract
// @Tags  	    Contract
// @Accept      json
// @Produce     json
// @Success     200 {object} deleteContractResponse
// @Failure     500 {object} errInternalServer
// @Failure     400 {object} errBadRequest
// @Router      /add_contract [post]
func (h *HttpHandler) deleteContracts(c *gin.Context) {
	dto := deleteContractsRequest{}
	err := c.ShouldBindJSON(&dto)
	if err != nil {
		h.Logger.Error("http  - delete contract - bind", zap.Error(err))
		c.JSON(http.StatusBadRequest, errBadRequest)
	}

	if len(dto.Addreses) == 0 {
		h.Logger.Error("http  - delete contract - zero request length")
		c.JSON(http.StatusBadRequest, errBadRequest)
	}

	err = h.TaskManager.DeleteContract(dto.Addreses)
	if err != nil {
		h.Logger.Error("http - delete contract ", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errInternalServer)
		return
	}

	c.JSON(http.StatusOK, &deleteContractResponse{Created: true})
}
