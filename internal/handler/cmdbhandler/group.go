package cmdbhandler

import "saurfang/internal/service/cmdbservice"

type GroupHandler struct {
	cmdbservice.GroupsService
}

func NewGroupHandler(svc *cmdbservice.GroupsService) *GroupHandler {
	return &GroupHandler{*svc}
}
