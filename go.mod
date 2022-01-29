module github.com/ubports/nuntium

go 1.13

require (
	github.com/jessevdk/go-flags v1.5.0
	launchpad.net/go-dbus v1.0.0-20140208094800-gubd5md7cro3mtxa
	launchpad.net/go-xdg v0.0.0-20140208094800-000000000010
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087
	launchpad.net/udm v0.0.0-20140721093638-000000000009
)

replace launchpad.net/go-dbus => github.com/z3ntu/go-dbus v0.0.0-20170220120108-c022b8b2e127

replace launchpad.net/udm => gitlab.com/sap_nocops/golang-udm v0.0.0-20220129171329-13c5f4438d5d
