#!/bin/ash

set -ex

# if thrown flags immediately,
# assume they want to run the main application
if [ "${1:0:1}" = '-' ]; then
	set -- /usr/bin/cert-manager-csi-driver "$@"
fi

exec "$@"
