package app

import (
	"context"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	"github.com/eurofurence/reg-attendee-service/internal/repository/authservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
)

type Application interface {
	Run() int
	Datagen() int
}

type Impl struct{}

func New() Application {
	return &Impl{}
}

func (i *Impl) Run() int {
	config.ParseCommandLineFlags()
	setupLogging("attendee-service", config.UseEcsLogging())

	if err := config.StartupLoadConfiguration(); err != nil {
		return 1
	}
	setLoglevel(config.LoggingSeverity())

	if err := database.Open(); err != nil {
		return 1
	}
	defer database.Close()
	if err := database.MigrateIfSwitchedOn(); err != nil {
		return 1
	}

	if err := paymentservice.Create(); err != nil {
		return 1
	}
	if err := mailservice.Create(); err != nil {
		return 1
	}
	if err := authservice.Create(); err != nil {
		return 1
	}

	attendeeService := attendeesrv.New()
	if err := runServerWithGracefulShutdown(attendeeService); err != nil {
		return 2
	}

	return 0
}

func (i *Impl) Datagen() int {
	config.ParseCommandLineFlags()
	setupLogging("attendee-service-datagen", config.UseEcsLogging())

	if err := config.StartupLoadConfiguration(); err != nil {
		return 1
	}
	setLoglevel(config.LoggingSeverity())

	if err := database.Open(); err != nil {
		return 1
	}
	defer database.Close()

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	attendeeService := attendeesrv.New()
	count := config.GenerateCount()
	if count == 0 {
		aulogging.Logger.Ctx(ctx).Error().Print("must specify generate-count option with value > 0. BAILING OUT.")
		return 1
	}

	aulogging.Logger.Ctx(ctx).Info().Printf("Now generating %d fake registrations...", count)

	err := attendeeService.GenerateFakeRegistrations(ctx, count)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("error while generating fake registrations: %s", err.Error())
		return 2
	}

	aulogging.Logger.Ctx(ctx).Info().Print("Remember to turn off email sending when working with these registrations!")
	aulogging.Logger.Ctx(ctx).Info().Print("Registrations generated successfully. Done.")

	return 0
}
