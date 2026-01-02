package services

import (
	"net/http"

	"react-golang-starter/internal/response"
)

func init() {
	// Organization errors
	response.RegisterSentinelError(ErrOrgNotFound, http.StatusNotFound, response.ErrCodeNotFound)
	response.RegisterSentinelError(ErrOrgSlugTaken, http.StatusConflict, response.ErrCodeConflict)
	response.RegisterSentinelError(ErrInvalidSlug, http.StatusBadRequest, response.ErrCodeValidation)
	response.RegisterSentinelError(ErrNotMember, http.StatusForbidden, response.ErrCodeForbidden)
	response.RegisterSentinelError(ErrInsufficientRole, http.StatusForbidden, response.ErrCodeForbidden)
	response.RegisterSentinelError(ErrCannotRemoveOwner, http.StatusBadRequest, response.ErrCodeBadRequest)
	response.RegisterSentinelError(ErrInvitationNotFound, http.StatusNotFound, response.ErrCodeNotFound)
	response.RegisterSentinelError(ErrInvitationExpired, http.StatusBadRequest, response.ErrCodeBadRequest)
	response.RegisterSentinelError(ErrInvitationAccepted, http.StatusConflict, response.ErrCodeConflict)
	response.RegisterSentinelError(ErrAlreadyMember, http.StatusConflict, response.ErrCodeConflict)
	response.RegisterSentinelError(ErrCannotChangeOwnRole, http.StatusBadRequest, response.ErrCodeBadRequest)
	response.RegisterSentinelError(ErrMustHaveOwner, http.StatusBadRequest, response.ErrCodeBadRequest)
	response.RegisterSentinelError(ErrInvitationEmailTaken, http.StatusConflict, response.ErrCodeConflict)
	response.RegisterSentinelError(ErrSeatLimitExceeded, http.StatusForbidden, response.ErrCodeForbidden)

	// File errors
	response.RegisterSentinelError(ErrAccessDenied, http.StatusForbidden, response.ErrCodeForbidden)
}
