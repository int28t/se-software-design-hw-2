#!/bin/sh
set -eu

pos_project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$pos_project_dir"

pos_demo_dir=$(mktemp -d /tmp/pos-demo.XXXXXX)
sed "s|@EXPORT_DIR@|$pos_demo_dir|g" docs/demo/session.txt > "$pos_demo_dir/session.txt"
go run -buildvcs=false ./cmd/pos < "$pos_demo_dir/session.txt" > "$pos_demo_dir/console.txt"

cat "$pos_demo_dir/console.txt"
printf '\nФайлы демонстрации: %s\n' "$pos_demo_dir"
for pos_report in final incomplete updated; do
    printf '\n%s.csv\n' "$pos_report"
    cat "$pos_demo_dir/$pos_report.csv"
done
