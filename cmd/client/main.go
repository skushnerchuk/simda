package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/skushnerchuk/simda/internal/client"
	"github.com/skushnerchuk/simda/internal/clientui"
	"github.com/skushnerchuk/simda/internal/clientui/splash"
	"github.com/skushnerchuk/simda/internal/clientui/theme"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const ClientVersion = "0.0.1"

var (
	receive uint
	warm    uint
	server  string
	port    string
)

const maxWarm = 120

var App *tview.Application

func validateParams() {
	if warm > maxWarm {
		fatal("warm cannot be greater than %d seconds\n", maxWarm)
	}
	if warm < receive {
		fatal("warm cannot be less than receive\n")
	}
}

var rootCmd = &cobra.Command{
	Use:     "simda",
	Short:   "System Information Monitoring DAemon client",
	Version: ClientVersion,
	Run: func(_ *cobra.Command, _ []string) {
		validateParams()

		App = tview.NewApplication()
		mainWindow := clientui.NewMainView(server, port, int(warm), int(receive))
		warmWindow := splash.NewWarmingWindow(int(warm))

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		defer cancel()
		errors, ctx := errgroup.WithContext(ctx)

		ch := make(chan *pb.Snapshot)
		defer close(ch)

		c := client.NewClient(warm, receive, server, port)

		App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() { //nolint:exhaustive
			case tcell.KeyCtrlC: // Block Ctrl-C for exit
				return nil
			case tcell.KeyCtrlQ: // exit
				cancel()
				App.Stop()
				return event
			default:
				return event
			}
		})

		errors.Go(
			func() error {
				if err := c.Run(ctx, ch, cancel); err != nil {
					return err
				}
				return nil
			},
		)

		t := time.NewTicker(1 * time.Second)
		elapsed := int(warm)
		defer t.Stop()

		errors.Go(
			func() error {
				if err := App.SetRoot(warmWindow.View, true).EnableMouse(true).Run(); err != nil {
					return err
				}
				return nil
			},
		)

	L:
		for {
			select {
			case <-t.C:
				elapsed--
				if elapsed >= 0 {
					warmWindow.Update(elapsed)
					App.Draw()
				} else {
					App.SetRoot(mainWindow.View, true)
					t.Stop()
				}
			case <-ctx.Done():
				cancel()
				App.Stop()
				break L
			case value := <-ch:
				mainWindow.SetData(value)
				App.Draw()
			}
		}

		err := errors.Wait()
		if err != nil {
			fatal("%s\n", err.Error())
		}
	},
}

func fatal(msg string, args ...any) {
	fmt.Printf(msg, args...)
	os.Exit(1)
}

func init() {
	cobra.OnInitialize()

	rootCmd.Flags().UintVarP(&receive, "receive", "r", 5, "receive snapshots every N seconds")
	rootCmd.Flags().UintVarP(&warm, "warm", "w", 5, "warm up time in seconds")
	rootCmd.Flags().StringVarP(&server, "server", "s", "127.0.0.1", "server ip")
	rootCmd.Flags().StringVarP(&port, "port", "p", "50051", "server port")
}

func main() {
	theme.ApplyTheme()

	if err := rootCmd.Execute(); err != nil {
		fatal("%s\n", err.Error())
	}
}
