#!/bin/bash

## this will retrieve all of the .go files that have been 
## changed since the last commit

echo "Runing prehook"

STAGED_GO_FILES=$(git diff --cached --name-only -- '*.go')

## we can check to see if this is empty
if [[ $STAGED_GO_FILES == "" ]]; then
    echo "No Go Files to Update"

## otherwise we can do stuff with these changed go files
else
    for file in $STAGED_GO_FILES; do
        ## format our file
        go fmt $file
        ## add any potential changes from our formatting to the 
        ## commit
        git add $file
    done
fi

make swagger > /dev/null

git add docs
