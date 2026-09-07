#!/bin/sh
# NeetoRecord CLI — session-start hook for Claude Code
# Lightweight auth liveness check. Always exits 0 (informational).

if ! command -v neetorecord >/dev/null 2>&1; then
  echo "NeetoRecord CLI is not installed or not on PATH."
  exit 0
fi

if neetorecord whoami >/dev/null 2>&1; then
  echo "NeetoRecord plugin active."
else
  echo "NeetoRecord CLI installed but not authenticated. Run 'neetorecord login' to authenticate."
fi

exit 0
