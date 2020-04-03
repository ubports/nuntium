#!/bin/sh

build_dir=$1
gopkg=$2

gopkg_path=$build_dir/usr/share/gocode/src/$gopkg

rm -r \
	$gopkg_path/test \
	$gopkg_path/storage \
	$gopkg_path/scripts \
	$gopkg_path/docs \
	$gopkg_path/.travis.yml \
	$gopkg_path/.gitignore \
	$gopkg_path/README.md \
	$gopkg_path/cmd \
	$gopkg_path/Jenkinsfile

[ -f $build_dir/usr/bin/*/test ] && rm $build_dir/usr/bin/*/test
[ -f $build_dir/usr/bin/test ] && rm $build_dir/usr/bin/test

exit 0
