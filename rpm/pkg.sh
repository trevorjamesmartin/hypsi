#!/bin/sh
NAME=hypsi
VERSION=$(grep "VERSION = " ../main.go|awk '{print $4}'|xargs echo)
FORMAT=tar.gz
PREFIX=$NAME-$VERSION
OUTPUT=$PREFIX.$FORMAT

if ! command -v rpmbuild 2>&1 > /dev/null
then
        echo "rpmbuild could not be found"
        exit 1
fi

if ! command -v git 2>&1 > /dev/null
then
        echo "git could not be found"
        exit 1
fi


pushd ..
git archive --prefix=$PREFIX/ --format=$FORMAT -o rpm/$OUTPUT HEAD
popd

rm ~/rpmbuild/SOURCES/$OUTPUT

mv $OUTPUT ~/rpmbuild/SOURCES

rpmbuild -bs $NAME.spec

rpmbuild -ba $NAME.spec

# copr-cli build $NAME ~/rpmbuild/SRPMS/$PREFIX-0.fc41.src.rpm

