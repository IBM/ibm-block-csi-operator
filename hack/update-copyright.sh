#!/bin/bash

CURRENT_PATH=$(dirname "$BASH_SOURCE")
PROTECT_ROOT=$CURRENT_PATH/..
VENDOR_PATH=$CURRENT_PATH/../vendor
BOILERPLATE=$CURRENT_PATH/boilerplate.go.txt

for file in $(find $PROTECT_ROOT -not -path "$VENDOR_PATH/*" -type f -name \*.go); do
  if [[ $(grep -n "\/\*" -m 1 $file | cut -f1 -d:) == 1 ]] && [[ $(grep -n "Copyright" -m 1 $file | cut -f1 -d:) == 2 ]]
  then
    # the file already has a copyright.
	continue
  else
    cat $BOILERPLATE > $file.tmp;
    echo "" >> $file.tmp;
    cat $file >> $file.tmp;
    mv $file.tmp $file;
  fi
done