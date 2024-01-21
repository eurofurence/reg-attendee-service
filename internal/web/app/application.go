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
	"github.com/eurofurence/reg-attendee-service/internal/repository/selfclient"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
	"sync"
	"time"
)

type Application interface {
	Run() int
	Datagen() int
	Loadtest() int
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
	config.AdditionalGeneratorCommandLineFlags()
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
	count := config.GeneratorGenerateCount()
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

func (i *Impl) Loadtest() int {
	config.AdditionalGeneratorCommandLineFlags()
	config.ParseCommandLineFlags()
	setupLogging("attendee-service-loadtest", config.UseEcsLogging())

	if err := config.StartupLoadConfiguration(); err != nil {
		return 1
	}
	setLoglevel(config.LoggingSeverity())

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	err := selfclient.Setup()
	if err != nil {
		aulogging.Logger.Ctx(ctx).Error().Printf("failed to set up client: %s. BAILING OUT.", err.Error())
		return 1
	}

	errorCount := 0
	errorCountMutex := sync.Mutex{}

	attendeeService := attendeesrv.New()
	count := config.GeneratorGenerateCount()
	parallel := config.GeneratorParallel()
	if count == 0 {
		aulogging.Logger.Ctx(ctx).Error().Print("must specify generate-count option with value > 0. BAILING OUT.")
		return 1
	}
	if parallel == 0 {
		aulogging.Logger.Ctx(ctx).Error().Print("must specify parallel option with value > 0. BAILING OUT.")
		return 1
	}

	var wg sync.WaitGroup
	for routine := uint(1); routine <= parallel; routine++ {
		wg.Add(1)
		thisRoutine := routine
		go func() {
			defer wg.Done()
			errsCount := i.loadtestSingle(attendeeService, thisRoutine, count/parallel)

			errorCountMutex.Lock()
			defer errorCountMutex.Unlock()
			errorCount += errsCount
		}()
	}

	wg.Wait()

	aulogging.Logger.Ctx(ctx).Info().Printf("Tried to generate %d registrations, %d successful, %d failed", count, count-uint(errorCount), errorCount)

	aulogging.Logger.Ctx(ctx).Info().Print("Remember to turn off email sending when working with these registrations!")
	aulogging.Logger.Ctx(ctx).Info().Print("Done.")

	return 0
}

func (i *Impl) loadtestSingle(attsrv attendeesrv.AttendeeService, routine uint, countPerRoutine uint) int {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	aulogging.Logger.Ctx(ctx).Info().Printf("routine %4d will generate %4d registrations in 5s", routine, countPerRoutine)
	time.Sleep(5 * time.Second)

	errorCount := 0
	for count := uint(1); count <= countPerRoutine; count++ {
		id, err := attsrv.SendFakeRegistrationToAPI(ctx)
		if err != nil {
			aulogging.Logger.Ctx(ctx).Error().Printf("routine %4d ERROR: %s", routine, err.Error())
			errorCount++
		} else {
			aulogging.Logger.Ctx(ctx).Info().Printf("routine %4d success id %s", routine, id)
		}
	}

	return errorCount
}
