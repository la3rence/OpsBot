#!/bin/bash
# ln -s $PWD/pre-commit.sh .git/hooks/pre-commit
go test ./...
RESULT=$?
if [[ $RESULT != 0 ]]; then
  echo "REJECTING COMMIT (test failed with status: $RESULT)"
  exit 1
fi

go fmt ./...
exit 0
