# gOTP

This tool allows OTP management and generating hash or time based OTPs.

## Installation

Compile using `go build -o gotp *.go`

## How to use it

`gotp l` List all existing configurations  
`gotp a -l google -u "otpauth://totp/Google:myGmailAddress%40googlemail.com?secret=aStringThatRepresentsTheSecret&issuer=Google&digits=6"` Add a new OTP with the label 'google'  
`gotp g google` Returns the current OTP  
`gotp d google` Deletes OTP called 'google' (with confirmation prompt)  

## Troubleshooting

There's currently no automated way of checking if a secret is base32 encoded or not. Because of this it is assumed by default that the secret has been submitted as a base32 encoded string. If you encounter problems with OTPs not being correct try changing the "base32" value in `~/.gotp/token`.

## TODOs

- base32 detection  
- PGP encryption support for token config file  
- token lifetime countdown  
- Tabulator auto-completion  

## Used 3rd party modules

Special thanks to the creators of these modules!  

https://github.com/davecgh/go-spew  
https://github.com/fatih/color  
https://github.com/hgfischer/go-otp  
https://github.com/urfave/cli  
