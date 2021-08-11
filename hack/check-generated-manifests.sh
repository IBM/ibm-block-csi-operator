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

ibm_operator_name=ibm-block-csi-operator
cd ..
cp -r $ibm_operator_name ./$ibm_operator_name-copy
cd $ibm_operator_name-copy/
hack/update-crds.sh
cp -r bin/ ../$ibm_operator_name/
cd ..
diff -qr $ibm_operator_name $ibm_operator_name-copy/
rm -rm $ibm_operator_name-copy/
cd $ibm_operator_name
