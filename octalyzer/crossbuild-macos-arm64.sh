#!/bin/sh
# TODO(modernize/phase-3): After GOPATH flatten, the local source layout
# changed from GoPath/src/paleotronic.com/* to the repository root. The
# remote target ($BASE/microm8/GoPath/src/paleotronic.com/) may still
# expect the old layout — update the remote side or rewrite this script
# before relying on it.

HOST="10.0.0.234"
ACCT="melodygriffiths"
BASE="Code"

REPO_ROOT="$(git rev-parse --show-toplevel)"
rsync -av \
    --exclude=.git \
    --exclude=flatpak \
    --exclude=microM8 \
    --exclude=coverage.out \
    "$REPO_ROOT/" "$ACCT@$HOST:$BASE/microm8/GoPath/src/paleotronic.com/"
#ssh "$ACCT@$HOST" "cd ${BASE}/microm8/GoPath/src/paleotronic.com/octalyzer && ./mac-build.sh"
