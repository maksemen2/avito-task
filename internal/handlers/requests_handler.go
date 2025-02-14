package handlers

import (
	"github.com/maksemen2/avito-task/internal/dao"
	"gorm.io/gorm"
)

type RequestsHandler struct {
	dao *dao.HolderDAO
}

func NewRequestsHandler(db *gorm.DB) *RequestsHandler {
	return &RequestsHandler{dao: dao.NewHolderDAO(db)}
}
