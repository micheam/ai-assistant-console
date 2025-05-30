#!/usr/bin/env bash
set -e

#
#/ Usage: prepare_release_body [-h] [-f FROM] [-t TO]
#/ Prepare GitHub release message body from given commit range.
#/
#/ Options:
#/   -f FROM        The tag or commit to start from (default: last tag).
#/   -t TO          The tag or commit to end at (default: HEAD).
#/   -h             show this message.
#/
#/ Examples:
#/    $ prepare_release_body -h
#/
#/    # Prepare release body from last tag to HEAD
#/    $ prepare_release_body
#/
#/    # Prepare release body from v0.1 to HEAD
#/    $ prepare_release_body -f v0.1
#/
#/    # Prepare release body from given range
#/    $ prepare_release_body -f v0.1 -t v0.2
#

usage() {
    grep '^#/' "${0}" | cut -c 3-
    echo ""
    exit 1
}

FROM="$(git describe --tags --abbrev=0 2>/dev/null || echo "HEAD")"
TO="HEAD"

_main() {
  local range="${FROM}..${TO}"
  if [ "$FROM" = "HEAD" ]; then
    range="HEAD"
  fi

  local format='{"hash":"%h","author":"%an","date":"%ad","message":"%s"}'
  local commits
  git log --pretty=format:"${format}" "${range}" |\
    jq -c 'select(.message | startswith("docs") | not)' |\
    jq -c 'select(.message | startswith("test") | not)' |\
    jq -r '"* [\(.hash)]: \(.message)"'


  # for commit in $(echo "${commits}" | jq -c '.[]'); do
  #   echo "$commit"
  # done
}

while getopts "hf:t:" opt; do
  case $opt in
    f) FROM="$OPTARG" ;;
    t) TO="$OPTARG" ;;
    h) usage ;;
    ?) usage ;;
  esac
done

shift $((OPTIND -1))

_main $@
