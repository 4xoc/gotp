package main

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
	"github.com/hgfischer/go-otp"
	"github.com/urfave/cli"
)

type tOTP struct {
	Type        string `json:"type"`        //totp or hotp
	Issuer      string `json:"issuer"`      //identifies the issuer
	User        string `json:"user"`        //idefifies a user if given
	Secret      string `json:"secret"`      //shared secret base32 encoded
	Label       string `json:"Label"`       //label for the account
	Description string `json:"description"` //optional description for the token usage
	Length      uint8  `json:"length"`      //number of digits/characters to return
	Counter     uint64 `json:"counter"`     //counter for HOTPs
	Base32      bool   `json:"base32"`      //true if secret is base32 encoded
}

func AddOTP(c *cli.Context) {
	var (
		err error
		otp tOTP
		uri string
	)

	// Using flags for setup
	if c.NumFlags() == 0 {
		color.Red("::Missing arguments")
	}

	if c.String("uri") == "" {
		color.Red("::Missing URI value")
		return
	} else {
		uri = c.String("uri")
	}

	if c.String("description") != "" {
		otp.Description = c.String("description")
	}

	if c.String("label") != "" {
		otp.Label = c.String("label")
	}

	err = parseURI(&uri, &otp)

	if err != nil {
		return
	}

	if Config.Debug {
		spew.Dump(otp)
	}

	// Final check if anything is fishy
	err = otp.Verify()

	if err != nil {
		return
	}

	Config.Tokens = append(Config.Tokens, &otp)

	// Save new OTP to config
	err = Config.SaveTokens()

	if err != nil {
		return
	}

	fmt.Printf("Successfully added new OTP called '%s'\n", otp.Label)
}

func GetOTP(c *cli.Context) {
	var (
		label string
		i     int
		totp  otp.TOTP
		hotp  otp.HOTP
	)

	label = c.Args().Get(0)

	if label == "" {
		color.Red("::Label argument is missing")
		return
	}

	for i = range Config.Tokens {
		if label == Config.Tokens[i].Label {
			if Config.Tokens[i].Type == "totp" {
				totp = otp.TOTP{Secret: Config.Tokens[i].Secret, IsBase32Secret: Config.Tokens[i].Base32}
				fmt.Printf("%s\n", totp.Get())
			} else {
				hotp = otp.HOTP{Secret: Config.Tokens[i].Secret, IsBase32Secret: Config.Tokens[i].Base32}
				fmt.Printf("%s\n", hotp.Get())
			}
		}
	}
}

func ListOTP(c *cli.Context) {
	var (
		i int
	)

	fmt.Printf("-------------------------------------------------------------------------------------------------------\n")
	fmt.Printf("|  %-20s|  %-20s|  %-20s|  %-30s|\n", "Label", "Issuer", "User", "Description")
	fmt.Printf("|----------------------|----------------------|----------------------|--------------------------------|\n")

	for i = range Config.Tokens {
		fmt.Printf("|  %-20s|  %-20s|  %-20s|  %-30s|\n", Config.Tokens[i].Label, Config.Tokens[i].Issuer, Config.Tokens[i].User, Config.Tokens[i].Description)
	}
	fmt.Printf("-------------------------------------------------------------------------------------------------------\n")
}

func DeleteOTP(c *cli.Context) {
	var (
		label string
		i     int
		resp  string
	)

	label = c.Args().Get(0)

	if label == "" {
		color.Red("::Label argument is missing")
		return
	}

	for i = range Config.Tokens {
		if label == Config.Tokens[i].Label {
			if !c.Bool("force") {
				// Prompt user for confirmation

				for true {
					fmt.Printf("Are you sure you want to delete OTP with label '%s'? [y/n]: ", label)
					fmt.Scanln(&resp)

					if resp == "n" {
						return
					}

					if resp == "y" {
						break
					}
				}
			}

			Config.Tokens[i] = Config.Tokens[len(Config.Tokens)-1]
			Config.Tokens = Config.Tokens[:len(Config.Tokens)-1]

			_ = Config.SaveTokens()

			fmt.Printf("Successfully deleted OTP with label '%s'\n", label)

			break
		}
	}
}

func parseURI(uri *string, otp *tOTP) error {
	var (
		err      error
		data     *url.URL
		values   url.Values
		tmpName  []string
		tmpLabel string
	)

	data, err = url.Parse(*uri)

	if err != nil {
		color.Red("::Could not parse otpauth URI")
		return err
	}

	values, err = url.ParseQuery(data.RawQuery)

	if err != nil {
		color.Red("::Could not parse otpauth URI query")
		return err
	}

	// TODO: proper base32 detection
	// most likely the secret is base32 encoded
	otp.Base32 = true

	otp.Type = data.Host
	tmpLabel = data.Path[1:] //trim slash

	tmpName = strings.Split(tmpLabel, ":")

	if len(tmpLabel) > 1 {
		// Actually split the string.
		if otp.Label == "" {
			// Only use the given label if none has been provided by the user.
			otp.Label = tmpName[0]
		}

		otp.User = tmpName[1]
	}

	if values["issuer"] != nil {
		// Issuer isn't mandatory thus checking if it is set or not.
		otp.Issuer = values["issuer"][0]
	}

	// Though it should be set everytime, better check for secret really existing
	if values["secret"] != nil {
		otp.Secret = values["secret"][0]
	} else {
		color.Red("::Missing shared secret")
		return err
	}

	// Since we have the verify function we skip error checking here
	// so we have all error messages at one place.
	if values["digits"] != nil {
		// reusing counter since we've allocated that memory anyway
		otp.Counter, _ = strconv.ParseUint(values["digits"][0], 10, 8)
		otp.Length = uint8(otp.Counter)
		otp.Counter = 0
	}

	if values["counter"] != nil {
		otp.Counter, _ = strconv.ParseUint(values["counter"][0], 10, 64)
	}

	return nil
}

// Verify checks if a given tOTP has all necessary variables set.
func (otp *tOTP) Verify() error {
	var (
		i           int
		errorExists bool
	)

	if otp.Label == "" {
		color.Red("::Label is missing")
		errorExists = true
	}

	if otp.Issuer == "" {
		color.Red("::Issuer is missing")
		errorExists = true
	}

	if otp.User == "" {
		color.Red("::User is missing")
		errorExists = true
	}

	if otp.Secret == "" {
		color.Red("::Secret is missing")
		errorExists = true
	}

	if otp.Length == 0 {
		color.Red("::Length isn't set")
		errorExists = true
	}

	if otp.Type == "hotp" && otp.Counter == 0 {
		color.Red("::Counter isn't set")
		errorExists = true
	}

	// Also check if a config for a specific label already exists.
	for i = range Config.Tokens {
		if Config.Tokens[i].Label == otp.Label {
			color.Red("::An OTP with label '%s' already exists, aborted\n", otp.Label)
			errorExists = true
			break
		}
	}

	if errorExists {
		return errors.New("")
	}

	return nil
}
