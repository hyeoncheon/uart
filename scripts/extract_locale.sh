#!/bin/bash
# vim: set ts=2 sw=2 expandtab:

langs="en-us ko-kr"

for lang in $langs; do
  echo "\n\nLANG: $lang"

  find templates -type f |xargs cat \
  |sed 's/t("/\nt("/g' \
  |grep 't("' \
  |sed 's/.*t("\([\._a-z0-9]*\)").*/- id: \1Xtranslation: "\1"/' \
  |while read line; do
    id=`echo $line|cut -dX -f1`
    grep -q "^$id\$" locales/*.$lang.yaml || echo "$line"
  done | sort -u | sed 's/X/\n/'

  grep 't(.*"\w*\(\.\w*\)*\.\w*"' `find actions -type f` \
  |grep -v "\.html" \
  |sed 's/.*"\(\w*\(\.\w*\)*\.\w*\)".*/- id: \1Xtranslation: "\1"/' \
  |while read line; do
    id=`echo $line|cut -dX -f1`
    grep -q "^$id\$" locales/*.$lang.yaml || echo "$line"
  done | sort -u | sed 's/X/\n/'
done
