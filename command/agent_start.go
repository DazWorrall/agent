package command

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/buildbox/agent/buildbox"
	"github.com/codegangsta/cli"
	"os"
)

func AgentStartCommandAction(c *cli.Context) {
	// Init debugging
	if c.Bool("debug") {
		buildbox.LoggerInitDebug()
	}

	// Create the agent
	if c.String("access-token") != "" {
		fmt.Println("buildbox: use of --access-token is now deprecated\nSee 'buildbox agent --help'")
		os.Exit(1)
	}

	agentRegistrationToken := c.String("token")
	if agentRegistrationToken == "" {
		fmt.Println("buildbox: missing token\nSee 'buildbox agent --help'")
		os.Exit(1)
	}

	// Create a client so we can register the agent
	var client buildbox.Client
	client.AuthorizationToken = agentRegistrationToken
	client.URL = c.String("url")

	// Register the agent
	agentAccessToken, err := client.AgentRegister(c.String("name"), c.StringSlice("meta-data"))
	if err != nil {
		buildbox.Logger.Fatal(err)
	}

	// Start the agent using the token we have
	agent := setupAgent(agentAccessToken, c.String("bootstrap-script"), c.String("url"))

	// Setup signal monitoring
	agent.MonitorSignals()

	// Start the agent
	agent.Start()
}

func setupAgent(agentAccessToken string, bootstrapScript string, url string) *buildbox.Agent {
	// Expand the envionment variable.
	bootstrapScript = os.ExpandEnv(bootstrapScript)

	// Make sure the boostrap script exists.
	if _, err := os.Stat(bootstrapScript); os.IsNotExist(err) {
		print("buildbox: no such file " + bootstrapScript + "\n")
		os.Exit(1)
	}

	// Set the agent options
	var agent buildbox.Agent
	agent.BootstrapScript = bootstrapScript

	// Client specific options
	agent.Client.AuthorizationToken = agentAccessToken
	agent.Client.URL = url

	// Setup the agent
	agent.Setup()

	// A nice welcome message
	buildbox.Logger.WithFields(logrus.Fields{
		"pid":     os.Getpid(),
		"version": buildbox.Version,
	}).Infof("Started agent `%s`", agent.Name)

	return &agent
}