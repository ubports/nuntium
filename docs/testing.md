# Testing nuntium

## Tools

These are generic debugging tools provided as part of nuntium or described as
external aids that can be used for troubleshooting. These do not stub
components, so are safe for operator testing.

### nuntium-decode-cli

This tool allows you to easily decode an *M-Retrieve.conf* displaying the
output of what is decoded to stdout with an option to drop the decoded
artifacts into a specific path.

Install it by running:

    go get github.com/ubports/nuntium/cmd/nuntium-decode-cli

Refer to the tool's [documentation](https://pkg.go.dev/github.com/ubports/nuntium/cmd/nuntium-decode-cli)
for more information.


### nuntium-preferred-context

*Needs implementation*

This tool allows reading or writing the preferred context `nuntium` will use
when trying to activate a context.

Install it by running:

    go get github.com/ubports/nuntium/cmd/nuntium-preferred-context

Refer to the tool's
[documentation](https://pkg.go.dev/github.com/ubports/nuntium/cmd/nuntium-preferred-context)
for more information.


### tcpdump

When doing operator testing and MMS debugging is needed, tcpdump can provide
the right tools to debug the issue.

To get it going, on the device, do:

     sudo mount -o remount,rw /
     sudo apt install tcpdump
     sudo tcpdump -w [file]

The capture `[file]` can be analyzed to better understand the problem.


### network-test-session

Additionally to `tcpdump`, a useful tool to pinpoint problems is
`network-test-session`, refer to the [documentation](https://github.com/sergiusens/network-test-session/blob/master/README.md)
for this tool on how to use it.


## Testing tools

This section describes a list of external tools to tests and a short summary
of what can be done with them. Some of these tools remove the operator from
the picture, so in the case of doing operator testing, these really should be
avoided.

These tools don't yet mock or stub `ofono`'s APN contexts logic but override
the logic in code completely if needed.


### nuntium-inject-push

This tool is meant to inject a push notification message through the
`ReceiveNotification` method nuntium exposes on the system bus and 
simulate a black box workflow as described in 
[Receiving an MMS](architecture.md#receiving-an-mms).

Install it by running:

    go get github.com/ubports/nuntium/cmd/nuntium-inject-push

This tool does not mock out ofono completely, but instead what it creates a
local server to serve an mms that would be passed on from the Push
Notification with a Content Location such that this local server would be
used to fetch the MMS.

If no MMS is specified a in tool *m-retrieve.conf* would be used. Once the
content is served once, the tool exits.

The agent is registered on the system bus, as such, it should be run like:

     sudo nuntium-inject-push --end-point :1.356

where `:1.356` is the dbus name that nuntium has on the system bus, this can
be discovered by looking at the nuntium logs.

To use a specifc mms, just use the cli switch like:

     sudo nuntium-inject-push --end-point :1.356


### nuntium-stub-send

*Needs implementation*

This tool verifies and simulates sending an MMS simulating what would happen
when [Sending an MMS](architecture.md#sending-an-mms)

It will track that the correct signals are raised over the bus.

Install it by running:

    go get github.com/ubports/nuntium/cmd/nuntium-stub-send

Refer to the tool's
[documentation](https://pkg.go.dev/github.com/ubports/nuntium/cmd/nuntium-stub-send)
for more information.
