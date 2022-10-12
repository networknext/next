#!/usr/bin/env bash
find . -name "*.cpp" | xargs sed -i .original -e 's/\t/    /g'
find . -name "*.c" | xargs sed -i .original -e 's/\t/    /g'
find . -name "*.h" | xargs sed -i .original -e 's/\t/    /g'
find . -name "*.original" -delete
