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

    go get github.com/ubuntu-phonedations/nuntium/cmd/nuntium-decode-cli

Refer to the tool's [documentation](http://godoc.org/github.com/ubuntu-phonedations/nuntium/cmd/nuntium-decode-cli)
for more information.


### nuntium-preferred-context

*Needs implementation*

This tool allows reading or writing the preferred context `nuntium` will use
when trying to activate a context.

Install it by running:

    go get github.com/ubuntu-phonedations/nuntium/cmd/nuntium-preferred-context

Refer to the tool's
[documentation](http://godoc.org/github.com/ubuntu-phonedations/nuntium/cmd/nuntium-preferred-context)
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

*Needs implementation*

This tool is meant to inject a push notification message through the
`ReceiveNotification` method nuntium exposes on the system bus and 
simulate a black box workflow as described in 
[Receiving an MMS](architecture.md#receiving-an-mms).

It will track if the correct `MessageAdded` signal was raised

Install it by running:

    go get github.com/ubuntu-phonedations/nuntium/cmd/nuntium-inject-push

Refer to the tool's
[documentation](http://godoc.org/github.com/ubuntu-phonedations/nuntium/cmd/nuntium-inject-push)
for more information.


### nuntium-stub-send

*Needs implementation*

This tool verifies and simulates sending an MMS simulating what would happen
when [Sending an MMS](architecture.md#sending-an-mms)

It will track that the correct signals are raised over the bus.

Install it by running:

    go get github.com/ubuntu-phonedations/nuntium/cmd/nuntium-stub-send

Refer to the tool's
[documentation](http://godoc.org/github.com/ubuntu-phonedations/nuntium/cmd/nuntium-stub-send)
for more information.


## Installing on Ubuntu

On Ubuntu (vivid+), all of the `nuntium-*` tools can be installed by

    sudo apt install nuntium-tools
