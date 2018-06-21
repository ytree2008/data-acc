package main

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fakewarp"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fakewarp/actions"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/urfave/cli"
	"log"
	"strings"
)

var testKeystore keystoreregistry.Keystore
var reader fileio.Reader

func getKeystore() keystoreregistry.Keystore {
	// TODO must be a better way to test this, proper factory?
	keystore := testKeystore
	if keystore == nil {
		keystore = etcdregistry.NewKeystore()
	}
	if reader == nil {
		reader = fileio.NewReader()
	}
	return keystore
}

func getActions(keystore keystoreregistry.Keystore) actions.FakewarpActions {
	volReg := keystoreregistry.NewVolumeRegistry(keystore)
	poolReg := keystoreregistry.NewPoolRegistry(keystore)
	return actions.NewFakewarpActions(poolReg, volReg, reader)
}

func createPersistent(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).CreatePersistentBuffer(c)
}

func showInstances(_ *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).ShowInstances()
}

func showSessions(_ *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).ShowSessions()
}

func listPools(_ *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).ListPools()
}

func showConfigurations(_ *cli.Context) error {
	fmt.Print(fakewarp.GetConfigurations())
	return nil
}

// TODO needs deleting, its a duplicate
func checkRequiredStrings(c *cli.Context, flags ...string) {
	errors := []string{}
	for _, flag := range flags {
		if str := c.String(flag); str == "" {
			errors = append(errors, flag)
		}
	}
	if len(errors) > 0 {
		log.Fatalf("Please provide these required parameters: %s", strings.Join(errors, ", "))
	}
}

func teardown(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).DeleteBuffer(c)
}

func jobProcess(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).ValidateJob(c)
}

func setup(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).CreatePerJobBuffer(c)
}

func realSize(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).RealSize(c)
}

func dataIn(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job")
	fmt.Printf("--token %s --job %s\n", c.String("token"), c.String("job"))

	keystore := getKeystore()
	defer keystore.Close()
	volReg := keystoreregistry.NewVolumeRegistry(keystore)

	volume, err := volReg.Volume(registry.VolumeName(c.String("token")))
	if err != nil {
		return err
	}

	if volume.SizeBricks == 0 {
		log.Println("skipping datain for:", volume.Name)
		return nil
	}

	err = volReg.UpdateState(volume.Name, registry.DataInRequested)
	if err != nil {
		return err
	}
	return volReg.WaitForState(volume.Name, registry.DataInComplete)
}

func paths(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job", "pathfile")
	fmt.Printf("--token %s --job %s --pathfile %s\n",
		c.String("token"), c.String("job"), c.String("pathfile"))
	// TODO get paths from the volume, and write out paths to given file
	return nil
}

func preRun(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job", "nodehostnamefile")
	fmt.Printf("--token %s --job %s --nodehostnamefile %s\n",
		c.String("token"), c.String("job"), c.String("nodehostnamefile"))

	keystore := getKeystore()
	defer keystore.Close()
	volReg := keystoreregistry.NewVolumeRegistry(keystore)

	volume, err := volReg.Volume(registry.VolumeName(c.String("token")))
	if err != nil {
		return err
	}

	if volume.SizeBricks == 0 {
		log.Println("skipping prerun for:", volume.Name)
		return nil
	}

	err = volReg.UpdateState(volume.Name, registry.MountRequested)
	if err != nil {
		return err
	}
	return volReg.WaitForState(volume.Name, registry.MountComplete)
}

func postRun(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job")
	fmt.Printf("--token %s --job %s\n",
		c.String("token"), c.String("job"))

	keystore := getKeystore()
	defer keystore.Close()
	volReg := keystoreregistry.NewVolumeRegistry(keystore)

	volume, err := volReg.Volume(registry.VolumeName(c.String("token")))
	if err != nil {
		return err
	}

	if volume.SizeBricks == 0 {
		log.Println("skipping postrun for:", volume.Name)
		return nil
	}

	err = volReg.UpdateState(volume.Name, registry.UnmountRequested)
	if err != nil {
		return err
	}
	return volReg.WaitForState(volume.Name, registry.UnmountComplete)
}

func dataOut(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job")
	fmt.Printf("--token %s --job %s\n",
		c.String("token"), c.String("job"))

	keystore := getKeystore()
	defer keystore.Close()
	volReg := keystoreregistry.NewVolumeRegistry(keystore)

	volume, err := volReg.Volume(registry.VolumeName(c.String("token")))
	if err != nil {
		return err
	}

	if volume.SizeBricks == 0 {
		log.Println("skipping data_out for:", volume.Name)
		return nil
	}

	err = volReg.UpdateState(volume.Name, registry.DataOutRequested)
	if err != nil {
		return err
	}
	return volReg.WaitForState(volume.Name, registry.DataOutComplete)
}
