package main

import (
	"context"
	"fmt"
	"os"

	"github.com/luxeave/entropy-course/internal/course/adapter/rest"
	"github.com/spf13/cobra"
)

func newRestCommand(ctx context.Context, scope *containerScope) *cobra.Command {
	address := rest.DefaultAddress
	command := &cobra.Command{
		Use:   "rest",
		Short: "Serve the REST API",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			container, err := scope.Container()
			if err != nil {
				return err
			}

			server, err := rest.NewServer(rest.Options{
				Course:       container.Course,
				Lesson:       container.Lesson,
				Quiz:         container.Quiz,
				Practice:     container.Practice,
				Test:         container.Test,
				Import:       container.Import,
				Token:        container.Config.APIToken,
				InstructorID: container.Config.InstructorID,
			})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(os.Stdout, "Course REST API listening on http://%s\n", address)
			return server.ListenAndServe(ctx, address)
		},
	}
	command.Flags().StringVar(&address, "addr", rest.DefaultAddress, "address to bind")

	return command
}
