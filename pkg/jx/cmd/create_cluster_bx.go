package cmd

import (
	"io"

	"strings"

	"fmt"

	"errors"

	//	os_user "os/user"

	"os"

	"github.com/Pallinder/go-randomdata"
	"github.com/jenkins-x/jx/pkg/jx/cmd/bx"
	"github.com/jenkins-x/jx/pkg/jx/cmd/log"
	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
	//	"github.com/jenkins-x/jx/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
	//	"regexp"
	"strconv"
)

type CreateClusterBXOptions struct {
	CreateClusterOptions
	Flags CreateClusterBXFlags
}

type CreateClusterBXFlags struct {
	ClusterName           string
	SkipLogin             bool
	KubeVersion           string
	Location              string
	PublicVLan            string
	PrivateVLan           string
	Workers               int
	MachineType           string
	Hardware              string
	NoSubnet              bool
	DisableDiskEncryption bool
	Trusted               bool
}

var (
	createClusterBXLong = templates.LongDesc(`
		This command creates a new kubernetes cluster on IBM Cloud, installing required local dependencies and provisions the
		Jenkins X platform

		You can see a demo of this command here: [http://jenkins-x.io/demos/create_cluster_bx/](http://jenkins-x.io/demos/create_cluster_bx)
`)

	createClusterBXExample = templates.Examples(`

		jx create cluster bx

`)
)

func NewCmdCreateClusterBX(f cmdutil.Factory, out io.Writer, errOut io.Writer) *cobra.Command {
	options := CreateClusterBXOptions{
		CreateClusterOptions: createCreateClusterOptions(f, out, errOut, BX),
	}
	cmd := &cobra.Command{
		Use:     "bx",
		Short:   "Create a new kubernetes cluster on IBM Bluemix: Runs on IBM Cloud",
		Long:    createClusterBXLong,
		Example: createClusterBXExample,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			cmdutil.CheckErr(err)
		},
	}

	options.addCreateClusterFlags(cmd)
	options.addCommonFlags(cmd)

	cmd.Flags().StringVarP(&options.Flags.ClusterName, optionClusterName, "n", "", "The name of this cluster, default is a randomly generated name")
	cmd.Flags().BoolVarP(&options.Flags.SkipLogin, "skip-login", "", false, "Skip authentication if already logged in")
	cmd.Flags().StringVarP(&options.Flags.KubeVersion, "kube-version", "", "", "The Kubernetes version you want the cluster to use. If present, at least major.minor must be specified. To see available versions, run 'bx cs kube-versions'")
	cmd.Flags().StringVarP(&options.Flags.Location, "location", "", "", "The location where you want to create the cluster. To see available locations, run 'bx cs locations'. The locations that are available to you depend on the IBM Cloud region you are logged in to. For IBM Cloud Dedicated accounts, do not specify this option. The location is pre-selected based on the location of the IBM Cloud infrastructure for the Dedicated instance.")
	cmd.Flags().StringVarP(&options.Flags.PublicVLan, "public-vlan", "", "", "The ID of the public VLAN. To see available VLANs, run 'bx cs vlans <location>'. If you do not have a public VLAN yet, do not specify this option because a public VLAN is automatically created for you. For IBM Cloud Dedicated accounts, do not specify this option. The public VLAN is preconfigured by IBM for you.")
	cmd.Flags().StringVarP(&options.Flags.PrivateVLan, "private-vlan", "", "", "The ID of the private VLAN. To see available VLANs, run 'bx cs vlans <location>'. If you do not have a private VLAN yet, do not specify this option because a private VLAN is automatically created for you. For IBM Cloud Dedicated accounts, do not specify this option. The private VLAN is preconfigured by IBM.")
	cmd.Flags().IntVarP(&options.Flags.Workers, "workers", "w", 1, "The number of cluster worker nodes. Defaults to 1")
	cmd.Flags().StringVarP(&options.Flags.MachineType, "machine-type", "m", "", "The machine type of the worker node. To see available machine types, run 'bx cs machine-types <location>' (for public IBM Cloud accounts) or 'bx cs machine-types' (for IBM Cloud Dedicated accounts)")
	cmd.Flags().StringVarP(&options.Flags.Hardware, "hardware", "", "", "The level of hardware isolation for your worker node. Use 'dedicated' to have available physical resources dedicated to you only, or 'shared' to allow physical resources to be shared with other IBM customers. For IBM Cloud Public accounts, the default value is shared. For IBM Cloud Dedicated accounts, dedicated is the only available option")
	cmd.Flags().BoolVarP(&options.Flags.NoSubnet, "no-subnet", "", false, "By default, both a public and a private portable subnet are created on the associated VLAN. Use 'true' to not create a portable subnet when you create a cluster, or 'false' to create one. The default is 'false'")
	cmd.Flags().BoolVarP(&options.Flags.DisableDiskEncryption, "disable-disk-encrypt", "", false, "Optional parameter to disable encryption on a worker node")
	cmd.Flags().BoolVarP(&options.Flags.Trusted, "trusted", "", false, "Optional parameter to enable trusted cluster feature")

	return cmd
}

func (o *CreateClusterBXOptions) Run() error {

	var deps []string
	d := binaryShouldBeInstalled("bx")
	if d != "" {
		deps = append(deps, d)
	}
	err := o.installMissingDependencies(deps)
	if err != nil {
		log.Errorf("error creating cluster on IBM Cloud, %v", err)
		return err
	}

	err = o.createClusterBX()
	if err != nil {
		log.Errorf("error creating cluster %v", err)
		return err
	}

	return nil
}

func (o *CreateClusterBXOptions) createClusterBX() error {
	var err error
	if !o.Flags.SkipLogin {
		// hard coded sso authentication: needs to get input from the user!
		// more precisely we need:
		// - API endpoint
		// - username
		// - password/key
		err := o.runCommand("bx", "login", "--sso")
		if err != nil {
			return err
		}
	}

	if o.Flags.ClusterName == "" {
		o.Flags.ClusterName = strings.ToLower(randomdata.SillyName())
		log.Infof("No cluster name provided so using a generated one: %s\n", o.Flags.ClusterName)
	}
	//
	// mandatory flags are machine type, num-nodes, zone,
	//args := []string{"container", "clusters", "create", o.Flags.ClusterName, "--zone", zone, "--num-nodes", numOfNodes, "--machine-type", machineType}
	args := []string{"cs", "cluster-create", "-s", "--name", o.Flags.ClusterName}

	if o.Flags.KubeVersion != "" {
		args = append(args, "--kube-version", o.Flags.KubeVersion)
	}

	location := o.Flags.Location
	if location != "" {
		if !isValidValue(location, bx.GetBluemixLocations()) {
			prompts := &survey.Select{
				Message:  "IBM Cloud Locations:",
				Options:  bx.GetBluemixLocations(),
				Help:     "Please select a valid machine type",
				PageSize: 10,
				Default:  "n1-standard-2",
			}

			err := survey.AskOne(prompts, &location, nil)
			if err != nil {
				return err
			}
		}
		args = append(args, "--location", location)
	}

	if o.Flags.PublicVLan != "" {
		args = append(args, "--public-vlan", o.Flags.PublicVLan)
	}
	if o.Flags.PrivateVLan != "" {
		args = append(args, "--private-vlan", o.Flags.PrivateVLan)
	}

	workers := o.Flags.Workers
	if workers != 1 {
		if workers <= 0 {
			return errors.New(fmt.Sprintf("Invalid number of worker nodes (%d should be >= 1)", workers))
		}
		args = append(args, "--workers", strconv.Itoa(workers))
	}

	machineType := o.Flags.MachineType
	if machineType != "" {
		if !isValidValue(machineType, bx.GetBluemixMachineTypes()) {
			prompts := &survey.Select{
				Message:  "IBM Cloud Machine Type:",
				Options:  bx.GetBluemixMachineTypes(),
				Help:     "Please select a valid machine type",
				PageSize: 10,
				Default:  "n1-standard-2",
			}

			err := survey.AskOne(prompts, &machineType, nil)
			if err != nil {
				return err
			}
		}
		args = append(args, "--machine-type", machineType)
	}

	hardware := o.Flags.Hardware
	if hardware != "" {
		args = append(args, "--hardware", hardware)
	}

	if o.Flags.NoSubnet {
		args = append(args, "--no-subnet", "true")
	}

	if o.Flags.DisableDiskEncryption {
		args = append(args, "--disable-disk-encrypt", "true")
	}

	if o.Flags.Trusted {
		args = append(args, "--trusted", "true")
	}

	err = o.runCommand("bx", args...)
	if err != nil {
		return err
	}

	o.InstallOptions.Flags.DefaultEnvironmentPrefix = o.Flags.ClusterName
	err = o.initAndInstall(BX)
	if err != nil {
		return err
	}

	output, err := o.getCommandOutput("bx", "cs", "cluster-config", o.Flags.ClusterName)

	if err != nil {
		return err
	}

	key, value := extractEnvironmentVariable(output)
	err = os.Setenv(key, value)

	if err != nil {
		log.Errorf("error setting environment variable %v", err)
		return err
	}
	log.Errorf(" EVERYTHING FINE")

	context, err := o.getCommandOutput("", "kubectl", "config", "current-context")
	if err != nil {
		return err
	}

	ns := o.InstallOptions.Flags.Namespace
	if ns == "" {
		f := o.Factory
		_, ns, _ = f.CreateClient()
		if err != nil {
			return err
		}
	}

	err = o.runCommand("kubectl", "config", "set-context", context, "--namespace", ns)
	if err != nil {
		return err
	}

	err = o.runCommand("kubectl", "get", "ingress")
	if err != nil {
		return err
	}

	return nil
}

func extractEnvironmentVariable(text string) (string, string) {
	// replace \n with a OS dependant constant
	lines := strings.Split(text, "\n")

	for i := 0; i < len(lines); i += 1 {
		line := lines[i]
		if strings.Contains(line, " KUBECONFIG=") {
			kubeConfig := strings.Split(line, " ")[1]
			keyValue := strings.Split(kubeConfig, "=")

			return keyValue[0], keyValue[1]
		}
	}

	return "", "" // add error management
}

func isValidValue(value string, validValues []string) bool {
	for _, validValue := range validValues {
		if validValue == value {
			return true
		}
	}

	return false
}
