#!/usr/bin/env sh

# 1. Setup

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

SUCCESS="${GREEN}✔︎${NC} "
FAILURE="${RED}✘${NC} "

gopath=`go env GOPATH`
mb="$gopath/bin/monobuild"
cd ./test/fixtures/manifests-test

exit_status=0

# Test helper

assert_eq()
{
  title=$1
  actual=$2
  expected=$3

  if [ "$actual" = "$expected" ]; then
    printf "${SUCCESS} $title\n"
  else
    printf "${FAILURE} $title\n"
    printf "${GREEN}Expected${NC}:\n\"$expected\"\n"
    printf "${RED}Actual${NC}:\n\"$actual\"\n"
    exit_status=1
  fi
}

# Tests


printf "Print command:\n"

# monobuild print
actual=$($mb print)
expected="app1: 
app2: 
app3: 
app4: 
libs/lib1: 
libs/lib2: 
libs/lib3: 
stack1: app1, app2, app3"

assert_eq "monobuild print" "$actual" "$expected"

# monobuild print --top-level
actual=$($mb print --top-level)
expected="app4: 
stack1: "

assert_eq "monobuild print --top-level" "$actual" "$expected"

# monobuild print --dependencies
actual=$($mb print --dependencies)
expected="app1: libs/lib1, libs/lib2
app2: libs/lib2, libs/lib3
app3: libs/lib3
app4: 
libs/lib1: libs/lib3
libs/lib2: libs/lib3
libs/lib3: 
stack1: app1, app2, app3"

assert_eq "monobuild print --dependencies" "$actual" "$expected"

# monobuild print --dependencies --top-level
actual=$($mb print --dependencies --top-level)
expected="app4: 
stack1: "

assert_eq "monobuild print --dependencies --top-level" "$actual" "$expected"

# monobuild print --dependencies --scope app1
actual=$($mb print --dependencies --scope app1)
expected="app1: libs/lib1, libs/lib2
libs/lib1: libs/lib3
libs/lib2: libs/lib3
libs/lib3: "

assert_eq "monobuild print --dependencies --scope app1" "$actual" "$expected"

# monobuild print --full
actual=$($mb print --full)
expected="app1: libs/lib1, libs/lib2
app2: libs/lib2, libs/lib3
app3: libs/lib3
app4: 
libs/lib1: libs/lib3
libs/lib2: libs/lib3
libs/lib3: 
stack1: !app1, !app2, !app3"

assert_eq "monobuild print --dependencies" "$actual" "$expected"

# monobuild diff 
printf "\nDiff command:\n"

changes="libs/lib2/change.txt
app4/app.bin"

# monobuild diff
actual=$(echo "$changes" | $mb diff -)
expected="app1: 
app2: 
app4: 
libs/lib2: 
stack1: app1, app2"

assert_eq "monobuild diff" "$actual" "$expected"

# monobuild diff --scope app2
changes="libs/lib2/change.txt
app4/app.bin"

actual=$(echo "$changes" | $mb diff --scope app2 -)
expected="app2: 
libs/lib2: "

assert_eq "monobuild diff --scope app2" "$actual" "$expected"

# monobuild diff --rebuild-strong
changes="libs/lib2/change.txt
app4/app.bin"

actual=$(echo "$changes" | $mb diff --rebuild-strong -)
expected="app1: 
app2: 
app3: 
app4: 
libs/lib2: 
stack1: app1, app2, app3"

assert_eq "monobuild diff --rebuild-strong" "$actual" "$expected"

# monobuild diff --rebuild-strong --scope app2
changes="libs/lib2/change.txt
app4/app.bin"

actual=$(echo "$changes" | $mb diff --rebuild-strong --scope app2 -)
expected="app2: 
libs/lib2: "

assert_eq "monobuild diff --rebuild-strong --scope app2" "$actual" "$expected"

# monobuild diff --top-level
changes="libs/lib2/change.txt
app4/app.bin"

actual=$(echo "$changes" | $mb diff --top-level -)
expected="app4: 
stack1: "

assert_eq "monobuild diff --top-level" "$actual" "$expected"

# monobuild diff --dependencies
actual=$(echo "$changes" | $mb diff --dependencies -)
expected="app1: libs/lib2
app2: libs/lib2
app4: 
libs/lib2: 
stack1: app1, app2"

assert_eq "monobuild diff --dependencies" "$actual" "$expected"

# monobuild diff --dependencies --scope app2
actual=$(echo "$changes" | $mb diff --dependencies --scope app2 -)
expected="app2: libs/lib2
libs/lib2: "

assert_eq "monobuild diff --dependencies --scope app2" "$actual" "$expected"

# monobuild diff --dependencies --rebuild-strong
changes="libs/lib2/change.txt
app4/app.bin"

actual=$(echo "$changes" | $mb diff --dependencies --rebuild-strong -)
expected="app1: libs/lib2
app2: libs/lib2
app3: 
app4: 
libs/lib2: 
stack1: app1, app2, app3"

assert_eq "monobuild diff --dependencies --rebuild-strong" "$actual" "$expected"

# monobuild diff --dependencies --rebuild-strong --scope app2
# Scope has priority
changes="libs/lib2/change.txt
app4/app.bin"

actual=$(echo "$changes" | $mb diff --dependencies --rebuild-strong --scope app2 -)
expected="app2: libs/lib2
libs/lib2: "

assert_eq "monobuild diff --dependencies --rebuild-strong --scope app2" "$actual" "$expected"

# monobuild diff --dependencies --top-level
# --top-level trumps --dependencies
changes="libs/lib2/change.txt
app4/app.bin"

actual=$(echo "$changes" | $mb diff --dependencies --top-level -)
expected="app4: 
stack1: "

assert_eq "monobuild diff --dependencies --top-level" "$actual" "$expected"

# monobuild diff --dependencies --top-level --scope app4
# --top-level trumps --dependencies
changes="libs/lib2/change.txt
app4/app.bin"

actual=$(echo "$changes" | $mb diff --dependencies --top-level --scope app4 -)
expected="app4: "

assert_eq "monobuild diff --dependencies --top-level --scope app4" "$actual" "$expected"

# monobuild diff --full
actual=$(echo "$changes" | $mb diff --full -)
expected="app1: libs/lib2
app2: libs/lib2
app4: 
libs/lib2: 
stack1: !app1, !app2"

assert_eq "monobuild diff --full" "$actual" "$expected"

# Return a status based on success
exit $exit_status

