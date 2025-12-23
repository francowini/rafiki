package vobjectiveactivityapp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/vobjectiveactivitybus"
	"github.com/francowini/rafiki/business/types/year"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/francowini/rafiki/foundation/web"
)

type app struct {
	log                   *logger.Logger
	vobjectiveactivityBus vobjectiveactivitybus.ExtBusiness
}

func newApp(log *logger.Logger, vobjectiveactivityBus vobjectiveactivitybus.ExtBusiness) *app {
	return &app{
		log:                   log,
		vobjectiveactivityBus: vobjectiveactivityBus,
	}
}

// query handles GET /v1/objectives/{objective_id}/activity?year={year}
func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
	// Get user from auth context
	userID := mid.GetSubjectID(ctx)
	if userID == uuid.Nil {
		return errs.New(errs.Unauthenticated, errors.New("user not authenticated"))
	}

	// Parse objective_id from path
	objectiveIDStr := web.Param(r, "objective_id")
	objectiveID, err := uuid.Parse(objectiveIDStr)
	if err != nil {
		return errs.New(errs.InvalidArgument, errs.NewFieldErrors("objective_id", err))
	}

	// Parse year from query parameter (default to current year)
	yearStr := r.URL.Query().Get("year")
	var yearInt int
	if yearStr == "" {
		yearInt = time.Now().Year()
	} else {
		yearInt, err = strconv.Atoi(yearStr)
		if err != nil {
			return errs.New(errs.InvalidArgument, errs.NewFieldErrors("year", err))
		}
	}

	yearVal, err := year.Parse(yearInt)
	if err != nil {
		return errs.New(errs.InvalidArgument, errs.NewFieldErrors("year", err))
	}

	filter := vobjectiveactivitybus.QueryFilter{
		ObjectiveID: objectiveID,
		UserID:      userID,
		Year:        yearVal,
	}

	activity, err := a.vobjectiveactivityBus.Query(ctx, filter)
	if err != nil {
		if errors.Is(err, vobjectiveactivitybus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		if errors.Is(err, vobjectiveactivitybus.ErrNotObjectiveOwner) {
			return errs.New(errs.PermissionDenied, err)
		}
		a.log.Error(ctx, "vobjectiveactivityapp.query", "objectiveID", filter.ObjectiveID, "year", filter.Year, "err", err)
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	return toAppActivityResponse(activity)
}
