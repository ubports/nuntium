#!/bin/sh

quit() {
	echo "ERROR: $1"
	exit 1
}

for input in *.msc
do
	output=$(echo $input | sed 's/msc$/png/')
	[ -n "$output" ] || quit "output file empty"
	mscgen -i "$input" -o "$output" -T png
	[ "$?" = 0 ] || quit "msgen failed to run"
done
