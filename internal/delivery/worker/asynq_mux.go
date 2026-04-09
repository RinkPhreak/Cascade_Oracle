package worker

import (
	"github.com/hibiken/asynq"

	"cascade/internal/application/usecase"
)

// NewServerMux registers background job routing functions
func NewServerMux(waterfallUC *usecase.WaterfallUseCase) *asynq.ServeMux {
	mux := asynq.NewServeMux()

	// Bind task types to handlers
	mux.HandleFunc("cascade:waterfall:process", HandleWaterfallTask(waterfallUC))

	return mux
}
