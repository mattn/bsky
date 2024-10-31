#!/bin/bash
# This script is an example of handling renaming of golang packages and repositories.
set -ex
ORIGINAL="github.com/mattn/bsky"
NEW="github.com/jlewi/bsctl"

# Handle renaming of module
find ./ -name "*.go"  -exec  sed -i ".bak" "s/${ORIGINAL}/${NEW}/g" {} ";"
# Find and update all go.mod files
find ./ -name "go.mod"  -exec sed -i ".bak" "s/${ORIGINAL}/${NEW}/g" {} ";"

# These rule updates all go files
find ./ -name "*.go"  -exec  sed -i ".bak" "s/pkg.loadConfig/pkg.LoadConfig/g" {} ";"
find ./ -name "*.go"  -exec  sed -i ".bak" "s/pkg.stringp/pkg.Stringp/g" {} ";"
find ./ -name "*.go"  -exec  sed -i ".bak" "s/pkg.int64p/pkg.Int64p/g" {} ";"
find ./ -name "*.go"  -exec  sed -i ".bak" "s/stringp/Stringp/g" {} ";"
find ./ -name "*.go"  -exec  sed -i ".bak" "s/int64p/Int64p/g" {} ";"
find ./ -name "*.go"  -exec  sed -i ".bak" "s/makeXRPCC/MakeXRPCC/g" {} ";"
find ./ -name "*.go"  -exec  sed -i ".bak" "s/timep/Timep/g" {} ";"
find ./ -name "*.go"  -exec  sed -i ".bak" "s/printPost/PrintPost/g" {} ";"
find ./ -name "*.go"  -exec  sed -i ".bak" "s/config/Config/g" {} ";"
find ./ -name "*.go"  -exec  sed -i ".bak" "s/cfg.verbose/cfg.Verbose/g" {} ";"
find ./ -name "*.go"  -exec  sed -i ".bak" "s/cfg.prefix/cfg.Prefix/g" {} ";"

find ./ -name "*.bak" -exec rm {} ";"