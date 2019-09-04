#!/bin/bash

CURRENT_PATH=$(dirname "$BASH_SOURCE")
DEPLOY_PATH=$CURRENT_PATH/../deploy
CRD_PATH=$DEPLOY_PATH/crds

TARGET_FILE_NAME=ibm-block-csi-operator.yaml
TARGET_FILE=$DEPLOY_PATH/$TARGET_FILE_NAME

excluded_files=("csi_driver.yaml" $TARGET_FILE_NAME)

function contains()
{
    local i
    for i in "${@:2}"
	do
        [[ "$i" == "$1" ]] && return 0;  # 0 is true
    done
    return 1  # 1 is false
}

echo "" > $TARGET_FILE

for file_name in $(ls $CRD_PATH)
do
    file=$CRD_PATH/$file_name
    if test -f $file
    then
        if [[ $file == *_crd.yaml ]]
        then
            cat $file >> $TARGET_FILE
            printf "\n---\n" >> $TARGET_FILE
        else
            echo "skip $file_name"
        fi
    else
        echo "skip $file_name, it is not a file"
    fi
done

for file_name in $(ls $DEPLOY_PATH)
do
    file=$DEPLOY_PATH/$file_name
    if test -f $file
    then
        if !(contains $file_name "${excluded_files[@]}")
        then
            cat $file >> $TARGET_FILE
            printf "\n---\n" >> $TARGET_FILE
        else
            echo "skip $file_name"
        fi
    else
        echo "skip $file_name, it is not a file"
    fi
done

# delete the last "---"
# this is only work on Mac, for linux, use sed -i '$d' $TARGET_FILE
sed -i '' -e '$d' $TARGET_FILE
sed -i '' -e '$d' $TARGET_FILE
