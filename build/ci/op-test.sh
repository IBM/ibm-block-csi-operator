#!/bin/bash
set -e

OP_TEST_SCRIPT_URL=${OP_TEST_SCRIPT_URL-"https://raw.githubusercontent.com/operator-framework/operator-test-playbooks/master/upstream/test/test.sh"}

bash <(curl -sL $OP_TEST_SCRIPT_URL) $*
rc=$?
[[ $rc -eq 0 ]] && echo "rc=$rc"  || { echo "Error: rc=$rc"; exit $rc; }
