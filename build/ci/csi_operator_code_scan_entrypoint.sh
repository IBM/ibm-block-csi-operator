#!/bin/bash

if [ -z "$EXCLUDE" ]
then
    gosec -fmt=junit-xml -log /results/logs -out /results/operator_code_scan_results.xml /operator
else
    gosec -exclude=${EXCLUDE} -fmt=junit-xml -log /results/logs -out /results/operator_code_scan_results.xml /operator
fi

