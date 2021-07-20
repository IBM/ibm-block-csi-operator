#!/bin/bash -e

#
# Copyright 2019 IBM Corp.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Add the copyright header for all fo files if not.

CURRENT_PATH=$(dirname "$BASH_SOURCE")
PROTECT_ROOT=$CURRENT_PATH/..
VENDOR_PATH=$CURRENT_PATH/../vendor
BOILERPLATE=$CURRENT_PATH/boilerplate.go.txt
excluded_files=("tools.go" "matan")

function contains()
{
    local i
    for i in "${@}"
	do
        [[ "$i" == "$1" ]] && return 0;  # 0 is true
    done
    return 1  # 1 is false
}

echo "add copyright header"

for file in $(find $PROTECT_ROOT -not -path "$VENDOR_PATH/*" -not -name "zz_generated*" -type f -name \*.go); do
  if [[ $(grep -n "\/\*" -m 1 $file | cut -f1 -d:) == 1 ]] && [[ $(grep -n "Copyright" -m 1 $file | cut -f1 -d:) == 2 ]]
  then
    # the file already has a copyright.
	continue
  else
    if !(contains $file_name "${excluded_files[@]}")
    then
      cat $BOILERPLATE > $file.tmp;
      echo "" >> $file.tmp;
      cat $file >> $file.tmp;
      mv $file.tmp $file;
    fi
  fi
done