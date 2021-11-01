package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xlrte/core/pkg/api"
	"github.com/xlrte/core/pkg/api/secrets"
	"github.com/xlrte/core/pkg/runtime/gcp"
)

type runArgs struct {
	rootDir       string
	deployVersion string
	environment   string
	targetDir     string
}

type runInputs struct {
	basePath string
	selector api.EnvResolver
	runtimes *api.Runtimes
}

func BuildCommand(ctx context.Context) *cobra.Command {

	rootCmd := &cobra.Command{
		Use:   "xlrte",
		Short: "xlrte gives you Dev & Ops super powers!",
		Long: `xlrte - A Cloud infrastructure builder for developers that gives you Dev & Ops super powers!
	        Documentation is available at http://github.com/xlrte/core`,
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				panic(err)
			}
		},
	}

	rootCmd.AddCommand(versionCommand(), providersCommand(), initProject(ctx),
		planCommand(ctx), applyCommand(ctx), deleteCommand(ctx), initSecretsCommand())

	return rootCmd
}

func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of xlrte",
		Long:  `All software has versions. This is xlrte's`,
		Run: func(cmd *cobra.Command, args []string) {
			PrintVersionInfo()
		},
	}
}

func providersCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "providers",
		Short: "lists installed cloud providers",
		Long:  `lists all the installed cloud providers. To install a new one, see the "install" command`,
		Run: func(cmd *cobra.Command, args []string) {
			providers := runtimes(".", ".")
			fmt.Println("Installed Cloud Providers:")
			for _, p := range providers.Runtimes {
				fmt.Println("  - " + p.Name())
			}
		},
	}
}

func initProject(ctx context.Context) *cobra.Command {
	var rootDir string
	var environment string
	var provider string
	var context string
	var region string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "initialize your cloud provider for an xlrte environment",
		Long:  `initializes an environment for a given cloud provider (but does not deploy), setting up the project/account and enabling the required services for the cloud provider`,
		Run: func(cmd *cobra.Command, args []string) {

			initSecretSystem(&rootDir, environment)
			targetDir := filepath.Join(".xlrte", "environments", environment)
			providers := runtimes(rootDir, targetDir)
			for _, p := range providers.Runtimes {
				if p.Name() == provider {
					initSecretSystem(&rootDir, environment)
					err := api.InitEnvironment(ctx, rootDir, environment, context, region, p)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					fmt.Println("Your cloud provider has been initialized to support xlrte, happy deploying!")
				}
			}
		},
	}
	cmd.Flags().StringVarP(&environment, "environment", "e", "", "Environment name")
	cmd.Flags().StringVarP(&provider, "provider", "p", "", "Cloud provider name, such as 'gcp'")
	cmd.Flags().StringVarP(&context, "context", "c", "", "Context, such as the GCP project name to run the environment in")
	cmd.Flags().StringVarP(&region, "region", "r", "", "The Cloud provider region to run the environment in")

	err := cmd.MarkFlagRequired("environment")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = cmd.MarkFlagRequired("provider")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = cmd.MarkFlagRequired("context")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = cmd.MarkFlagRequired("region")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return cmd
}

func deleteCommand(ctx context.Context) *cobra.Command {
	yes := ""
	theArgs := runArgs{}
	export := &cobra.Command{
		Use:   "delete",
		Short: "deletes the infrastructure for an environment",
		Long:  `Deletes the entire infrastructure for an environment`,
		Run: func(cmd *cobra.Command, args []string) {
			input := theArgs.toRunInputs()
			var err error
			text := yes
			if yes != "yes" {
				fmt.Println("Are you sure you want to delete all resources? ('yes', or any other input for no)")
				reader := bufio.NewReader(os.Stdin)
				text, err = reader.ReadString('\n')
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
			if text == "yes\n" || text == "yes" {
				fmt.Println("Deleting based on configuration directory: " + theArgs.rootDir)
				err := api.Execute(ctx, api.Delete, input.basePath, input.selector, input.runtimes)
				if err != nil {

					checkSecretInit(err, theArgs.environment)
					fmt.Println(err)
					os.Exit(1)
				}
				return
			}
			fmt.Println("Delete cancelled")
		},
	}
	export.Flags().StringVarP(&yes, "confirm", "y", "", "Confirms delete (non-interactive run), give 'yes' as an argument")
	addRunTags(export, &theArgs)
	return export
}

func planCommand(ctx context.Context) *cobra.Command {
	theArgs := runArgs{}
	plan := &cobra.Command{
		Use:   "plan",
		Short: "shows the changes of a deployment without applying the changes",
		Long:  `calculates & shows the changes of a deployment, without applying the changes`,
		Run: func(cmd *cobra.Command, args []string) {
			input := theArgs.toRunInputs()
			fmt.Println("Building from configuration directory: " + theArgs.rootDir)
			err := api.Execute(ctx, api.Plan, input.basePath, input.selector, input.runtimes)
			if err != nil {
				checkSecretInit(err, theArgs.environment)
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	addRunTags(plan, &theArgs)
	return plan
}

func applyCommand(ctx context.Context) *cobra.Command {
	yes := ""
	theArgs := runArgs{}
	apply := &cobra.Command{
		Use:   "apply",
		Short: "apply deployment",
		Long:  `apply the deployment`,
		Run: func(cmd *cobra.Command, args []string) {
			input := theArgs.toRunInputs()
			var err error
			text := yes
			if yes != "yes" {
				fmt.Println("Are you sure you want to apply? ('yes', or any other input for no)")
				reader := bufio.NewReader(os.Stdin)
				text, err = reader.ReadString('\n')
				if err != nil {
					checkSecretInit(err, theArgs.environment)
					fmt.Println(err)
					os.Exit(1)
				}
			}

			if text == "yes\n" || text == "yes" {
				fmt.Println("Building from configuration directory: " + theArgs.rootDir)
				err := api.Execute(ctx, api.Apply, input.basePath, input.selector, input.runtimes)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				return
			}
			fmt.Println("apply cancelled.")

		},
	}

	apply.Flags().StringVarP(&yes, "confirm", "y", "", "Confirms apply (non-interactive run), give 'yes' as an argument")
	addRunTags(apply, &theArgs)
	return apply
}

func initSecretSystem(rootDir *string, environment string) {
	if rootDir == nil || *rootDir == "" {
		*rootDir = ".xlrte/config"
	}
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	initialized, err := secrets.InitSecrets(dirname, *rootDir, environment)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if initialized {
		fmt.Println("")
		fmt.Println("A private key has been created for you in $HOME/.xlrte/private-key.asc")
		fmt.Println("Please make sure to keep your private key and passphrase secure from prying eyes and backed up.")
		fmt.Println("A lost private key & passphrase cannot be recovered. If no user remains with access to at least one functioning key, secrets data will be lost.")
		fmt.Println("")
		fmt.Println("Keys & passphrases should not be shared. Each user requiring access should create their own with `xlrte secret init`, then commit their public key to the repository and request an existing user to re-encrypt all secrets with `xlrte secret refresh`")
	}
}

func initSecretsCommand() *cobra.Command {
	var rootDir string
	var environment string
	var name string
	command := &cobra.Command{
		Use:   "secret",
		Short: "manage secrets & secret config",
		Long:  `manage secrets & secret config`,
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				panic(err)
			}
		},
	}
	subCommands := []*cobra.Command{
		{
			Use:   "init",
			Short: "initialize secret management for environment",
			Long:  `initialize secret management for environment`,
			Run: func(cmd *cobra.Command, args []string) {
				initSecretSystem(&rootDir, environment)
			},
		},
		{
			Use:   "add",
			Short: "add a new secret",
			Long:  `add a new secret`,
			Run: func(cmd *cobra.Command, args []string) {
				if rootDir == "" {
					rootDir = ".xlrte/config"
				}
				err := secrets.AddSecret(rootDir, environment, name)
				if err != nil {
					checkSecretInit(err, environment)
					fmt.Println(err)
					os.Exit(1)
				}
			},
		},
		{
			Use:   "refresh",
			Short: "re-encrypts all secrets with currently available public keys",
			Long:  `re-encrypts all secrets with currently available public keys, useful when new keys are added or removed`,
			Run: func(cmd *cobra.Command, args []string) {
				if rootDir == "" {
					rootDir = ".xlrte/config"
				}
				dirname, err := os.UserHomeDir()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				err = secrets.Refresh(dirname, rootDir, environment)
				if err != nil {
					checkSecretInit(err, environment)
					fmt.Println(err)
					os.Exit(1)
				}
			},
		},
	}
	for _, cmd := range subCommands {
		cmd.Flags().StringVarP(&environment, "environment", "e", "", "Environment name")
		err := cmd.MarkFlagRequired("environment")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if cmd.Name() == "add" {
			cmd.Flags().StringVarP(&name, "name", "n", "", "Name of secret")
			err = cmd.MarkFlagRequired("name")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	command.AddCommand(
		subCommands...,
	)

	command.Flags().StringVarP(&environment, "environment", "e", "", "Environment name")
	err := command.MarkFlagRequired("environment")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return command
}

func (theArgs *runArgs) toRunInputs() runInputs {
	if theArgs.rootDir == "" {
		theArgs.rootDir = ".xlrte/config"
	}
	var selector api.EnvResolver
	selector = &api.ArgEnvResolver{
		EnvName:    theArgs.environment,
		ArgVersion: theArgs.deployVersion,
	}

	if theArgs.deployVersion == "" {
		selector = &api.FileEnvResolver{
			EnvName: theArgs.environment,
			BaseDir: theArgs.rootDir,
		}
	}

	modulesDir := theArgs.targetDir
	if theArgs.targetDir == "" {
		theArgs.targetDir = filepath.Join(".xlrte", "environments", selector.Env())
		modulesDir = filepath.Join(".xlrte", "environments")
		err := os.MkdirAll(theArgs.targetDir, 0750)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	rtes := runtimes(modulesDir, theArgs.targetDir)

	return runInputs{
		basePath: theArgs.rootDir,
		selector: selector,
		runtimes: rtes,
	}
}

func addRunTags(command *cobra.Command, args *runArgs) {
	command.Flags().StringVarP(&args.deployVersion, "version", "v", "", "Version to deploy")
	command.Flags().StringVarP(&args.environment, "environment", "e", "", "Environment name")

	err := command.MarkFlagRequired("environment")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runtimes(modulesDir, baseDir string) *api.Runtimes {
	rt := gcp.NewRuntime(modulesDir, baseDir)
	return &api.Runtimes{
		Runtimes: []api.Runtime{
			rt,
		},
	}
}

func checkSecretInit(err error, environment string) {
	if strings.Contains(err.Error(), "private-key.asc: no such file or directory") {
		fmt.Printf("No private key found. Have you previously run `xlrte secret init -e %s` to initialise the secrets for the environment?", environment)
		fmt.Println("")
		fmt.Println("If you have run `xlrte secret init`, but secrets exist from before, they may need to be re-encrypted by a colleague with `xlrte secret refresh`")
		os.Exit(1)
	}
}
