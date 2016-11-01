package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
)

type tConfig struct {
	Debug    bool    `json:"-"`
	NewSetup bool    `json:"-"`
	HomeDir  string  `json:"-"`
	PGPKeyID string  `json:"key"`
	Tokens   []*tOTP `json:"tokens,omitempty"`
}

func checkConfiguration(debug bool) {
	var (
		TmpConfig tConfig
	)

	usr, err := user.Current()

	if err != nil {
		fmt.Printf("Failed to read user dir. Error: %s", err.Error())
	}

	//setting HomeDir to pass along to ReadGlobalConfiguration.
	Config = &TmpConfig
	Config.Debug = debug
	Config.HomeDir = usr.HomeDir
	err = Config.ReadGlobalConfiguration()

	if err != nil && !Config.NewSetup {
		os.Exit(1)
	}

	if Config.NewSetup {
		err = Config.CreateNewSetup()

		if err != nil {
			os.Exit(2)
		}
	}

	if Config.Debug {
		fmt.Printf("Using configuration: %s", spew.Sdump(Config))
	}
}

// ReadGlobalConfiguration reads the general config file into a tConfig struct.
func (c *tConfig) ReadGlobalConfiguration() error {
	var (
		err       error
		tmpByte   []byte
		TmpConfig tConfig
	)

	_, err = ioutil.ReadFile(c.HomeDir + "/.gotp/config")

	if err != nil {
		// Check if file is missing or permissions don't allow read
		if err.Error() == "open "+c.HomeDir+"/.gotp/config: no such file or directory" {
			c.NewSetup = true
		} else {
			color.Red("::Error opening config file: %s\n", err.Error())
			return err
		}
	} else {
		// Reading config file into tConfig.
		tmpByte, err = ioutil.ReadFile(c.HomeDir + "/.gotp/config")

		if err != nil {
			color.Red("::Failed to open config file: %s\n", err.Error())
			return err
		}

		err = json.Unmarshal(tmpByte, &TmpConfig)

		if err != nil {
			color.Red("::Failed to read config file: %s\n", err.Error())
			return err
		}

		// Setting HomeDir and re-assigning the global tConfig to TmpConfig
		TmpConfig.HomeDir = c.HomeDir
		TmpConfig.Debug = c.Debug

		Config = &TmpConfig

		// Load keys from seperate file
		err = Config.ReadTokens()

		if err != nil {
			color.Red("::Failed to read OTP config file: %s\n", err.Error())
			return err
		}

	}
	return nil
}

// CreateNewSetup leads the user through the process of creating
// a new general configuration file and setting up gOTP for
// proper usage.
func (c *tConfig) CreateNewSetup() error {
	var (
		err     error
		tmpByte []byte
	)

	fmt.Printf("::Starting Initial Configuration\n")

	// Creating config dir if possible.
	err = os.Mkdir(c.HomeDir+"/.gotp/", 0700)

	if err != nil {
		// We only want to fail when the dir couldn't be created.
		if err.Error() != "mkdir "+c.HomeDir+"/.gotp/: file exists" {
			color.Red("::Failed to create config dir: %s\n", err.Error())
			return err
		}
	} else {
		fmt.Printf("::Created config dir\n")
	}

	tmpByte, err = json.Marshal(c)

	if err != nil {
		color.Red("::Failed to encode config struct: %s\n", err.Error())
		return err
	}

	// Now we can write the config file to disk.
	err = ioutil.WriteFile(c.HomeDir+"/.gotp/config", tmpByte, 0600)

	return nil
}

func (c *tConfig) ReadTokens() error {
	var (
		err  error
		data []byte
	)

	data, err = ioutil.ReadFile(c.HomeDir + "/.gotp/token")

	if err != nil {
		// It's okay to have no token config (yet)
		if err.Error() == "open "+c.HomeDir+"/.gotp/token: no such file or directory" {
			return nil
		}

		color.Red("::Failed to open OTP config file: %s\n", err.Error())
		return err
	}

	err = json.Unmarshal(data, &c.Tokens)

	if err != nil {
    if err.Error() == "open "+c.HomeDir+"/.gotp/token: no such file or directory" {
      return nil
    }

		color.Red("::Failed to read OTP config file: %s\n", err.Error())
		return err
	}

	if Config.Debug {
		spew.Dump(c.Tokens)
	}

	return nil
}

func (c *tConfig) SaveTokens() error {
	var (
		err  error
		data []byte
	)

	data, err = json.Marshal(c.Tokens)

	if err != nil {
		color.Red("::Failed to encode OTP config struct: %s\n", err.Error())
		return err
	}

	err = ioutil.WriteFile(c.HomeDir+"/.gotp/token", data, 0600)

	if err != nil {
		color.Red("::Failed to write OTP config file: %s\n", err.Error())
		return err
	}

	return nil
}
