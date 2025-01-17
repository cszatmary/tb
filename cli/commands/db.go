package commands

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/TouchBistro/goutils/command"
	"github.com/TouchBistro/goutils/errors"
	"github.com/TouchBistro/goutils/fatal"
	"github.com/TouchBistro/tb/cli"
	"github.com/TouchBistro/tb/engine"
	"github.com/spf13/cobra"
)

func newDBCommand(c *cli.Container) *cobra.Command {
	return &cobra.Command{
		Use:        "db <service-name>",
		Deprecated: "it will be removed soon",
		Short:      "Connects to the database in a service",
		Long: `Connects to the database in a service using a cli database client.

	Examples:
	- Connect to the partners-config-service's database.
		tb db partners-config-service`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c.Tracker.Info("checking required env vars.")

			serviceName := args[0]
			dbConf, err := getDbConf(c.Ctx, c, serviceName)
			if err != nil {
				return &fatal.Error{
					Msg: "Could not retrieve database config for this service.",
					Err: err,
				}
			}

			var cliName string
			var connArg string

			switch dbConf.dbType {
			case "postgresql":
				cliName = "pgcli"
				connArg = fmt.Sprintf("%s://%s:%s@localhost:%s/%s", dbConf.dbType, dbConf.user, dbConf.password, dbConf.port, dbConf.name)
			case "mysql":
				cliName = "mycli"
				connArg = fmt.Sprintf("%s://%s:%s@localhost:%s/%s", dbConf.dbType, dbConf.user, dbConf.password, dbConf.port, dbConf.name)
			case "mssql":
				cliName = "mssql-cli"
				connArg = fmt.Sprintf("-U %s -P %s -S localhost -d %s", dbConf.user, dbConf.password, dbConf.name)
				fmt.Println(connArg)
			default:
				return &fatal.Error{
					Msg: fmt.Sprintf("DB_TYPE %s is not currently supported by tb db. Please consider making a pull request or let the maintainers know about your use case.", dbConf.dbType),
				}
			}

			if !command.Exists(cliName) {
				return &fatal.Error{
					Msg: fmt.Sprintf("This command requires %s for %s, which uses a %s database.\n Please install it then run tb db.", cliName, serviceName, dbConf.dbType),
				}
			}

			c.Tracker.Infof("starting %s...", cliName)

			err = command.New(command.WithStdin(os.Stdin), command.WithStdout(os.Stdout), command.WithStderr(os.Stderr)).
				Exec(cliName, strings.Fields(connArg)...)
			if err != nil {
				return &fatal.Error{
					Msg: fmt.Sprintf("could not start database client %s", cliName),
					Err: err,
				}
			}
			return nil
		},
	}
}

type dbConfig struct {
	dbType   string
	name     string
	port     string
	user     string
	password string
}

func getDbConf(ctx context.Context, c *cli.Container, serviceName string) (dbConfig, error) {
	required := []string{"DB_TYPE", "DB_NAME", "DB_PORT", "DB_USER", "DB_PASSWORD"}
	missing := "missing"

	// This is ugly, but less ugly than using printenv and much faster than doing individual execs for every var
	// generates a command in the following format: sh -c echo ${var1:-missing} ${var2:-missing} ...${varN:-missing}
	// mssing is used as a blank value instead of an empty string to make producing nicer errors to the user much easier.
	var sb strings.Builder
	sb.WriteString("echo")
	for _, req := range required {
		sb.WriteString(fmt.Sprintf(" ${%s:-%s}", req, missing))
	}
	args := []string{"sh", "-c", sb.String()}

	buf := &bytes.Buffer{}
	exitCode, err := c.Engine.Exec(ctx, serviceName, engine.ExecOptions{
		Cmd:    args,
		Stdin:  os.Stdin,
		Stdout: buf,
		Stderr: os.Stderr,
	})
	if err != nil {
		return dbConfig{}, errors.Wrap(err, errors.Meta{Reason: "failed execing command inside this service's container"})
	}
	if exitCode != 0 {
		return dbConfig{}, fmt.Errorf("failed execing command inside this service's container")
	}

	result := strings.Split(strings.TrimSpace(buf.String()), " ")

	// Validate that all required env vars were found.
	notFound := make([]string, 0)
	for i, res := range result {
		if res == missing {
			notFound = append(notFound, required[i])
		}
	}
	if len(notFound) != 0 {
		return dbConfig{}, fmt.Errorf("the following required env vars were not defined: [%s]", strings.Join(notFound, ", "))
	}

	conf := dbConfig{
		dbType:   result[0],
		name:     result[1],
		port:     result[2],
		user:     result[3],
		password: result[4],
	}

	return conf, nil
}
